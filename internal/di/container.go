package di

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/monobearotaku/online-chat-api/internal/config"
	auth_v1 "github.com/monobearotaku/online-chat-api/internal/ports/api/auth/v1"
	chat_v1 "github.com/monobearotaku/online-chat-api/internal/ports/api/chat/v1"
	"github.com/monobearotaku/online-chat-api/internal/postgres"
	auth_repo "github.com/monobearotaku/online-chat-api/internal/repository/auth"
	chat_repo "github.com/monobearotaku/online-chat-api/internal/repository/chat"
	"github.com/monobearotaku/online-chat-api/internal/service/auth"
	"github.com/monobearotaku/online-chat-api/internal/service/chat"
	"github.com/monobearotaku/online-chat-api/internal/service/tokenizer"
	"google.golang.org/grpc"
)

type DiContainer struct {
	chatV1Server *chat_v1.ChatV1
	authV1Server *auth_v1.AuthV1

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

	db, closeFunc := postgres.NewDbConnection(ctx, config)
	closeFunctions = append(closeFunctions, closeFunc)

	authRepo := auth_repo.NewAuthRepo(db)
	chatRepo := chat_repo.NewChatRepo(db)

	tokenizer := tokenizer.NewTokenizer()

	authService := auth.NewAuthService(authRepo, tokenizer, db)
	chatService := chat.NewChatService(chatRepo, tokenizer, db)

	dialer := grpc.NewServer()

	authV1 := auth_v1.NewAuthV1(dialer, authService)
	chatV1 := chat_v1.NewChatV1(dialer, chatService, tokenizer)

	return &DiContainer{
		chatV1Server:   chatV1,
		authV1Server:   authV1,
		listener:       listener,
		server:         dialer,
		closeFunctions: closeFunctions,
	}
}

func (di *DiContainer) Run() {
	go func() {
		_ = di.server.Serve(di.listener)
	}()

	fmt.Println("Di container started on port:", di.listener.Addr().String())
}

func (di *DiContainer) Stop() {
	fmt.Println("Stopping app")

	_ = di.server.GracefulStop
	_ = di.listener.Close()
}
