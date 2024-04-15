package di

import (
	"context"
	"fmt"
	"log"
	"net"

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
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type DiContainer struct {
	chatV1Server *chat_v1.ChatV1
	authV1Server *auth_v1.AuthV1

	kafkaConcumer *consumer.Consumer

	listener net.Listener
	server   *grpc.Server

	closeFunctions []func()
}

func NewDiContainer(ctx context.Context) *DiContainer {
	config := config.ParseConfig()

	closeFunctions := make([]func(), 0)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Panicf("Failed to start grpc dialer: %v", err)
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

	dialer := grpc.NewServer()
	reflection.Register(dialer)

	authV1 := auth_v1.NewAuthV1(dialer, authService)
	chatV1 := chat_v1.NewChatV1(dialer, chatService, tokenizer)

	return &DiContainer{
		chatV1Server:   chatV1,
		authV1Server:   authV1,
		listener:       listener,
		server:         dialer,
		closeFunctions: closeFunctions,
		kafkaConcumer:  kafkaConcumer,
	}
}

func (di *DiContainer) Run(ctx context.Context) {
	go func() {
		_ = di.server.Serve(di.listener)
	}()

	go func() {
		di.kafkaConcumer.Consume(ctx)
	}()

	fmt.Println("Di container started on port:", di.listener.Addr().String())
}

func (di *DiContainer) Stop() {
	fmt.Println("Stopping app")

	_ = di.server.GracefulStop
	_ = di.listener.Close()
}
