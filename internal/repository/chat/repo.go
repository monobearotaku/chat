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
			name
		FROM chats
		WHERE id = $1
	`

	cht := chat.Chat{}

	err := c.db.QueryRow(ctx, query, chatId).Scan(&cht.ID, &cht.Name)
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
			name
		FROM chats
		WHERE name = $1
	`

	cht := chat.Chat{}

	err := c.db.QueryRow(ctx, query, name).Scan(&cht.ID, &cht.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return chat.Chat{}, chat.ErrChatNotFound
		}

		return chat.Chat{}, err
	}

	return cht, nil
}

func (c *chatRepo) GetChatUsers(ctx context.Context, chatID int64) (chat.ChatUsers, error) {
	const query = `
		SELECT 
			user_id,
			role
		FROM users_to_chats
		WHERE chat_id = $1
	`

	rows, err := c.db.Query(ctx, query, chatID)
	if err != nil {
		return chat.ChatUsers{}, err
	}

	users := chat.ChatUsers{}

	for rows.Next() {
		var userID int64
		var role string

		err = rows.Scan(&userID, &role)
		if err != nil {
			return chat.ChatUsers{}, err
		}

		users.Users = append(users.Users, chat.ChatUser{
			UserID: userID,
			Role:   chat.Role(role),
		})
	}

	return users, nil
}

func (c *chatRepo) CreateChat(ctx context.Context, cht chat.Chat) (chat.Chat, error) {
	const query = `
		INSERT INTO chats(name)
		VALUES ($1)
		ON CONFLICT(name) DO NOTHING
		RETURNING id
	`

	err := c.db.QueryRow(ctx, query, cht.Name).Scan(&cht.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return chat.Chat{}, chat.ErrChatAlreadyExists
		}

		return chat.Chat{}, err
	}

	return cht, nil
}

func (c *chatRepo) AddUserToChat(ctx context.Context, chatID int64, userID int64, role chat.Role) error {
	const query = `
		INSERT INTO users_to_chats(chat_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (chat_id, user_id) DO NOTHING
	`

	_, err := c.db.Exec(ctx, query, chatID, userID, role.String())
	if err != nil {
		return err
	}

	return nil
}

func (c *chatRepo) SaveMessage(ctx context.Context, msg chat.Message) error {
	const query = `
		INSERT INTO messages(chat_id, user_id, message)
		VALUES ($1, $2, $3)
		ON CONFLICT (chat_id, user_id) DO NOTHING
	`

	_, err := c.db.Exec(ctx, query, msg.ChatID, msg.UserID, msg.Msg)
	if err != nil {
		return err
	}

	return nil
}
