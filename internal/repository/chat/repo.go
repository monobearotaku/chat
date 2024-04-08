package chat

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/monobearotaku/online-chat-api/internal/domain/chat"
	"github.com/monobearotaku/online-chat-api/internal/postgres"
)

type chatRepo struct {
	db postgres.QueryExecer
}

func NewChatRepo(db postgres.QueryExecer) Repo {
	return &chatRepo{
		db: db,
	}
}

func (c *chatRepo) WithTx(tx postgres.Tx) Repo {
	return &chatRepo{
		db: tx,
	}
}

func (c *chatRepo) GetById(ctx context.Context, chatId int64) (chat.Chat, error) {
	const query = `
		SELECT
			id,
			user_id,
			chat_owner,
			name
		FROM chat
		WHERE id = $1
	`

	cht := chat.Chat{}

	err := c.db.QueryRow(ctx, query, chatId).Scan(&cht.ID, &cht.UserIDs, &cht.OwnerID, &cht.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return chat.Chat{}, chat.ErrChatNotFound
		}

		return chat.Chat{}, err
	}

	return cht, nil
}

func (c *chatRepo) GetByName(ctx context.Context, name string) (chat.Chat, error) {
	const query = `
		SELECT
			id,
			user_id,
			chat_owner,
			name
		FROM chat
		WHERE name = $1
	`

	cht := chat.Chat{}

	err := c.db.QueryRow(ctx, query, name).Scan(&cht.ID, &cht.UserIDs, &cht.OwnerID, &cht.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return chat.Chat{}, chat.ErrChatNotFound
		}

		return chat.Chat{}, err
	}

	return cht, nil
}

func (c *chatRepo) CreateChat(ctx context.Context, cht chat.Chat) (chat.Chat, error) {
	const query = `
		INSERT INTO chat(user_id, name, chat_owner)
		VALUES ($1, $2, $3)
		ON CONFLICT(name) DO NOTHING
		RETURNING id
	`

	err := c.db.QueryRow(ctx, query, cht.UserIDs, cht.Name, cht.OwnerID).Scan(&cht.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return chat.Chat{}, chat.ErrChatAlreadyExists
		}

		return chat.Chat{}, err
	}

	return cht, nil
}
