package chat

import (
	"context"

	"github.com/monobearotaku/online-chat-api/internal/domain/chat"
	"github.com/monobearotaku/online-chat-api/internal/postgres"
)

type Repo interface {
	WithTx(tx postgres.Tx) Repo
	GetById(ctx context.Context, chatID int64) (chat.Chat, error)
	GetChatUsers(ctx context.Context, chatID int64) (chat.ChatUsers, error)
	GetByName(ctx context.Context, name string) (chat.Chat, error)
	CreateChat(ctx context.Context, chat chat.Chat) (chat.Chat, error)
	AddUserToChat(ctx context.Context, chatID int64, userID int64, role chat.Role) error
	SaveMessage(ctx context.Context, msg chat.Message) error
}
