package di

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/monobearotaku/online-chat-api/internal/config"
	"github.com/monobearotaku/online-chat-api/internal/kafka/producer"
	auth_v1 "github.com/monobearotaku/online-chat-api/internal/ports/api/auth/v1"
	chat_v1 "github.com/monobearotaku/online-chat-api/internal/ports/api/chat/v1"
	consumer "github.com/monobearotaku/online-chat-api/internal/ports/kafka/consumers"
	"github.com/monobearotaku/online-chat-api/internal/postgres"
	auth_repo "github.com/monobearotaku/online-chat-api/internal/repository/auth"
	chat_repo "github.com/monobearotaku/online-chat-api/internal/repository/chat"
	"github.com/monobearotaku/online-chat-api/internal/service/auth"
	"github.com/monobearotaku/online-chat-api/internal/service/chat"
	"github.com/monobearotaku/online-chat-api/internal/service/tokenizer"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type DiContainer struct {
	chatV1Server *chat_v1.ChatV1
	authV1Server *auth_v1.AuthV1

	kafkaConcumer *consumer.Consumer

	grpcListener net.Listener
	httpListener net.Listener

	server *grpc.Server
	mux    *http.Server

	closeFunctions []func()
}

func NewDiContainer(ctx context.Context) *DiContainer {
	config := config.ParseConfig()

	closeFunctions := make([]func(), 0)

	grpcListener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Printf("Failed to start grpc dialer: %v\n", err)
	}

	httpListener, err := net.Listen("tcp", ":8001")
	if err != nil {
		fmt.Printf("Failed to start http dialer: %v\n", err)
	}

	kafkaProducer := producer.NewProducer(config)

	db, closeFunc := postgres.NewDbConnection(ctx, config)
	closeFunctions = append(closeFunctions, closeFunc)

	authRepo := auth_repo.NewAuthRepo(db)
	chatRepo := chat_repo.NewChatRepo(db)

	tokenizer := tokenizer.NewTokenizer()

	authService := auth.NewAuthService(authRepo, tokenizer, db)
	chatService := chat.NewChatService(chatRepo, authRepo, tokenizer, kafkaProducer, db)

	kafkaConcumer := consumer.NewConsumer(config, chatService)

	dialer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
			recovery.StreamServerInterceptor(),
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	tracer, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.Tracer.Url)))
	if err != nil {
		fmt.Printf("Failed to start http tracer: %v\n", err)
	}

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("chat"),
		),
	)

	otel.SetTracerProvider(tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(tracer),
		tracesdk.WithResource(r),
	))

	reflection.Register(dialer)
	grpc_prometheus.Register(dialer)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	authV1 := auth_v1.NewAuthV1(dialer, authService)
	chatV1 := chat_v1.NewChatV1(dialer, chatService, tokenizer)

	return &DiContainer{
		chatV1Server: chatV1,
		authV1Server: authV1,
		grpcListener: grpcListener,
		httpListener: httpListener,
		server:       dialer,
		mux: &http.Server{
			Handler: mux,
		},
		closeFunctions: closeFunctions,
		kafkaConcumer:  kafkaConcumer,
	}
}

func (di *DiContainer) Run(ctx context.Context) {
	go func() {
		fmt.Println("Di container started on port:", di.grpcListener.Addr().String())
		_ = di.server.Serve(di.grpcListener)
	}()

	go func() {
		di.kafkaConcumer.Consume(ctx)
	}()

	go func() {
		fmt.Println("Metrics started on port:", di.httpListener.Addr().String())
		di.mux.Serve(di.httpListener)
	}()

}

func (di *DiContainer) Stop() {
	fmt.Println("Stopping app")

	_ = di.server.GracefulStop
	_ = di.grpcListener.Close()
	_ = di.httpListener.Close()
}
