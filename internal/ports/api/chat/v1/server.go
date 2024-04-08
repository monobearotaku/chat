package v1

import (
	"github.com/monobearotaku/online-chat-api/internal/service/chat"
	"github.com/monobearotaku/online-chat-api/internal/service/tokenizer"
	chatv1 "github.com/monobearotaku/online-chat-api/proto/chat/v1"
	"google.golang.org/grpc"
)

type ChatV1 struct {
	chatv1.UnimplementedChatServiceServer
	chatService chat.Service
	tokenizer   tokenizer.Tokenizer
}

func NewChatV1(dialer grpc.ServiceRegistrar, chatService chat.Service, tokenizer tokenizer.Tokenizer) *ChatV1 {
	server := ChatV1{
		chatService: chatService,
		tokenizer:   tokenizer,
	}

	chatv1.RegisterChatServiceServer(dialer, &server)

	return &server
}
