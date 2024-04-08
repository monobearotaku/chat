package chat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/monobearotaku/online-chat-api/internal/pkg/slices"
	"github.com/monobearotaku/online-chat-api/internal/service/tokenizer"
	chatv1 "github.com/monobearotaku/online-chat-api/proto/chat/v1"

	chatDomain "github.com/monobearotaku/online-chat-api/internal/domain/chat"
	"github.com/monobearotaku/online-chat-api/internal/domain/token"

	"github.com/monobearotaku/online-chat-api/internal/postgres"
	"github.com/monobearotaku/online-chat-api/internal/repository/chat"
)

type connection struct {
	stream   chatv1.ChatService_ConnectToChatServer
	userID   int64
	userUuid string
}

type connections []connection

type chatService struct {
	chat       chat.Repo
	tokenizer  tokenizer.Tokenizer
	txBeginner postgres.TxBeginner

	chatIdToStream map[int64]connections
	clientToChatId map[int64]map[int64]struct{}

	mu *sync.Mutex
}

func NewChatService(chat chat.Repo, tokenizer tokenizer.Tokenizer, txBeginner postgres.TxBeginner) Service {
	return &chatService{
		chat:           chat,
		tokenizer:      tokenizer,
		txBeginner:     txBeginner,
		chatIdToStream: make(map[int64]connections),
		clientToChatId: make(map[int64]map[int64]struct{}),
		mu:             &sync.Mutex{},
	}
}

func (c *chatService) CreateChat(ctx context.Context, ownerID int64, name string) (newChat chatDomain.Chat, err error) {
	if utf8.RuneCountInString(name) == 0 {
		return chatDomain.Chat{}, chatDomain.ErrChatInvalidName
	}

	tx, err := c.txBeginner.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return chatDomain.Chat{}, fmt.Errorf("Chat.Service.CreateChat begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			if err := tx.Commit(ctx); err != nil {
				_ = tx.Rollback(ctx)
			}
		}
	}()

	cht := chatDomain.Chat{}

	cht.Name = name
	cht.OwnerID = ownerID
	cht.UserIDs = append(cht.UserIDs, cht.OwnerID)

	newChat, err = c.chat.WithTx(tx).CreateChat(ctx, cht)
	if err != nil {
		if errors.Is(err, chatDomain.ErrChatAlreadyExists) {
			return chatDomain.Chat{}, err
		}

		return chatDomain.Chat{}, fmt.Errorf("Chat.Service.CreateChat creating chat: %w", err)
	}

	return newChat, nil
}

func (c *chatService) ValidateChat(ctx context.Context, userID, chatID int64) (err error) {
	cht, err := c.chat.GetById(ctx, chatID)
	if err != nil {
		if errors.Is(err, chatDomain.ErrChatNotFound) {
			return err
		}

		return fmt.Errorf("Chat.Service.ValidateChat failed to get chat by id: %w", err)
	}

	if !slices.Contains(cht.UserIDs, userID) {
		return chatDomain.ErrChatHaveNoUser
	}

	return nil
}

func (c *chatService) JoinChat(ctx context.Context, userID, chatID int64) (token.Token, error) {
	err := c.ValidateChat(ctx, userID, chatID)
	if err != nil {
		return "", err
	}

	return c.tokenizer.CreateToken(ctx, map[string]string{
		"userID": strconv.FormatInt(userID, 10),
		"chatID": strconv.FormatInt(chatID, 10),
	})
}

func (c *chatService) StartMessaging(ctx context.Context, userID, chatID int64, stream chatv1.ChatService_ConnectToChatServer) error {
	userUuid := uuid.NewString()

	c.addConnection(stream, userID, chatID, userUuid)
	defer c.removeConnection(userID, userUuid)

	for {
		msg, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		c.sendMsg(&chatv1.ChatMessageResponse{
			UserId:  userID,
			Message: msg.Message,
			ChatId:  chatID,
		})
	}

	return nil
}

func (c *chatService) sendMsg(msg *chatv1.ChatMessageResponse) {
	c.mu.Lock()
	allConnections := c.chatIdToStream[msg.ChatId]
	c.mu.Unlock()

	for _, connect := range allConnections {
		if connect.userID != msg.UserId {
			connect.stream.Send(msg)
		}
	}
}

func (c *chatService) addConnection(stream chatv1.ChatService_ConnectToChatServer, userID, chatID int64, userUuid string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.clientToChatId[userID]; !ok {
		c.clientToChatId[userID] = map[int64]struct{}{}
	}
	c.clientToChatId[userID][chatID] = struct{}{}

	c.chatIdToStream[chatID] = append(c.chatIdToStream[chatID], connection{
		userUuid: userUuid,
		userID:   userID,
		stream:   stream,
	})
}

func (c *chatService) removeConnection(userID int64, userUuid string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ids, ok := c.clientToChatId[userID]
	if !ok {
		return
	}

	delete(c.clientToChatId, userID)

	for key := range ids {
		conns := c.chatIdToStream[key]

		conns = slices.RemoveFunc(conns, func(i int) bool {
			return conns[i].userUuid == userUuid
		})

		if len(c.chatIdToStream[key]) == 0 {
			delete(c.chatIdToStream, key)
		}
	}
}
