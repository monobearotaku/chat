package chat

import (
	"context"

	"github.com/monobearotaku/online-chat-api/internal/domain/token"

	"github.com/monobearotaku/online-chat-api/internal/domain/chat"
	chatv1 "github.com/monobearotaku/online-chat-api/proto/chat/v1"
)

type Service interface {
	JoinChat(context.Context, int64, int64) (token.Token, error)
	CreateChat(context.Context, int64, string) (chat.Chat, error)
	ValidateChat(context.Context, int64, int64) error
	StartMessaging(context.Context, int64, int64, chatv1.ChatService_ConnectToChatServer) error
	AddUserToChat(context.Context, int64, int64, int64) error
	SendMessage(context.Context, string, chat.Message)
}
