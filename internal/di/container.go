package di

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
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
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
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

	logger log.Logger

	closeFunctions []func()
}

func NewDiContainer(ctx context.Context) *DiContainer {
	config := config.ParseConfig()

	logger := log.NewJSONLogger(os.Stderr)
	logger = log.With(logger, "service", "chat")

	closeFunctions := make([]func(), 0)

	grpcListener, err := net.Listen("tcp", ":8000")
	if err != nil {
		level.Error(logger).Log("error", fmt.Errorf("failed to start grpc dialer: %v", err))
	}

	httpListener, err := net.Listen("tcp", ":8001")
	if err != nil {
		level.Error(logger).Log("error", fmt.Errorf("failed to start http dialer: %v", err))
	}

	kafkaProducer := producer.NewProducer(config)

	db, closeFunc := postgres.NewDbConnection(ctx, config)
	closeFunctions = append(closeFunctions, closeFunc)

	authRepo := auth_repo.NewAuthRepo(db)
	chatRepo := chat_repo.NewChatRepo(db)

	tokenizer := tokenizer.NewTokenizer()

	authService := auth.NewAuthService(authRepo, tokenizer, db)
	chatService := chat.NewChatService(chatRepo, authRepo, tokenizer, kafkaProducer, db)

	kafkaConcumer := consumer.NewConsumer(config, chatService, logger)

	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Minute,
		PermitWithoutStream: true,
	}

	kaserver := keepalive.ServerParameters{
		Time:    10 * time.Minute,
		Timeout: 20 * time.Second,
	}

	dialer := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kaserver),
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			recovery.UnaryServerInterceptor(),
			logging.UnaryServerInterceptor(
				interceptorLogger(logger),
				logging.WithFieldsFromContext(generateLogFields),
				logging.WithLogOnEvents(
					logging.FinishCall,
				),
			),
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
			recovery.StreamServerInterceptor(),
			logging.StreamServerInterceptor(
				interceptorLogger(logger),
				logging.WithFieldsFromContext(generateLogFields),
				logging.WithLogOnEvents(
					logging.StartCall,
					logging.FinishCall,
				),
			),
		),

		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	tracer, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.Tracer.Url)))
	if err != nil {
		level.Error(logger).Log("error", fmt.Errorf("failed to start http tracer: %v", err))
	}

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("chat"),
		),
	)
	if err != nil {
		level.Error(logger).Log("error", fmt.Errorf("failed to start tracer resource: %v", err))
	}

	otel.SetTracerProvider(tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
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
		logger:         logger,
		closeFunctions: closeFunctions,
		kafkaConcumer:  kafkaConcumer,
	}
}

func (di *DiContainer) Run(ctx context.Context) {
	go func() {
		level.Info(di.logger).Log("message", fmt.Sprintf("gRCP started on port: %s", di.grpcListener.Addr().String()))
		_ = di.server.Serve(di.grpcListener)
	}()

	go func() {
		level.Info(di.logger).Log("message", "kafka consumer started")
		di.kafkaConcumer.Consume(ctx)
	}()

	go func() {
		level.Info(di.logger).Log("message", fmt.Sprintf("metrics started on port: %s", di.grpcListener.Addr().String()))
		di.mux.Serve(di.httpListener)
	}()

}

func (di *DiContainer) Stop() {
	level.Info(di.logger).Log("message", "stopping application")

	_ = di.server.GracefulStop
	_ = di.grpcListener.Close()
	_ = di.httpListener.Close()
}

func generateLogFields(ctx context.Context) logging.Fields {
	if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
		return logging.Fields{"traceID", span.TraceID().String()}
	}
	return nil
}

func interceptorLogger(l log.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			_ = level.Debug(l).Log(fields...)
		case logging.LevelInfo:
			_ = level.Info(l).Log(fields...)
		case logging.LevelWarn:
			_ = level.Warn(l).Log(fields...)
		case logging.LevelError:
			_ = level.Error(l).Log(fields...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
