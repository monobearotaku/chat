package v1

import (
	"github.com/monobearotaku/online-chat-api/internal/service/auth"
	auth_v1 "github.com/monobearotaku/online-chat-api/proto/auth/v1"
	"google.golang.org/grpc"
)

type AuthV1 struct {
	auth_v1.UnimplementedAuthServiceServer
	authService auth.Service
}

func NewAuthV1(dialer grpc.ServiceRegistrar, authService auth.Service) *AuthV1 {
	server := AuthV1{
		authService: authService,
	}

	auth_v1.RegisterAuthServiceServer(dialer, &server)
	return &server
}
