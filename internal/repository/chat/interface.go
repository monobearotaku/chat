package chat

import (
	"context"

	"github.com/monobearotaku/online-chat-api/internal/domain/chat"
	"github.com/monobearotaku/online-chat-api/internal/postgres"
)

type Repo interface {
	WithTx(postgres.Tx) Repo
	GetById(context.Context, int64) (chat.Chat, error)
	GetByName(context.Context, string) (chat.Chat, error)
	CreateChat(context.Context, chat.Chat) (chat.Chat, error)
}
