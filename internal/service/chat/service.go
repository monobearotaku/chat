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
	"github.com/monobearotaku/online-chat-api/internal/domain"
	"github.com/monobearotaku/online-chat-api/internal/kafka/producer"
	"github.com/monobearotaku/online-chat-api/internal/pkg/slices"
	"github.com/monobearotaku/online-chat-api/internal/service/tokenizer"
	chatv1 "github.com/monobearotaku/online-chat-api/proto/chat/v1"

	chatDomain "github.com/monobearotaku/online-chat-api/internal/domain/chat"
	"github.com/monobearotaku/online-chat-api/internal/domain/token"

	"github.com/monobearotaku/online-chat-api/internal/postgres"
	"github.com/monobearotaku/online-chat-api/internal/repository/auth"
	"github.com/monobearotaku/online-chat-api/internal/repository/chat"
)

type connection struct {
	stream   chatv1.ChatService_ConnectToChatServer
	userID   int64
	userUuid string
}

type connections []connection

type chatService struct {
	chat chat.Repo
	auth auth.Repo

	tokenizer  tokenizer.Tokenizer
	txBeginner postgres.TxBeginner

	producer producer.Producer

	chatIdToStream map[int64]connections
	clientToChatId map[int64]map[int64]struct{}

	mu *sync.Mutex
}

func NewChatService(chat chat.Repo, auth auth.Repo, tokenizer tokenizer.Tokenizer, producer producer.Producer, txBeginner postgres.TxBeginner) Service {
	return &chatService{
		chat:           chat,
		auth:           auth,
		tokenizer:      tokenizer,
		txBeginner:     txBeginner,
		producer:       producer,
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
		if err == nil {
			err := tx.Commit(ctx)
			if err != nil {
				_ = tx.Rollback(ctx)
			}
		}
	}()

	cht := chatDomain.Chat{}
	cht.Name = name

	cht, err = c.chat.WithTx(tx).CreateChat(ctx, cht)
	if err != nil {
		if errors.Is(err, chatDomain.ErrChatAlreadyExists) {
			return chatDomain.Chat{}, err
		}

		return chatDomain.Chat{}, fmt.Errorf("Chat.Service.CreateChat creating chat: %w", err)
	}

	err = c.chat.WithTx(tx).AddUserToChat(ctx, cht.ID, ownerID, chatDomain.Owner)
	if err != nil {
		return chatDomain.Chat{}, fmt.Errorf("Chat.Service.AddUserToChat adding owner to chat: %w", err)
	}

	return cht, nil
}

func (c *chatService) ValidateChat(ctx context.Context, userID, chatID int64) (err error) {
	chtUsers, err := c.chat.GetChatUsers(ctx, chatID)
	if err != nil {
		if errors.Is(err, chatDomain.ErrChatNotFound) {
			return err
		}

		return fmt.Errorf("Chat.Service.ValidateChat failed to get chat users by id: %w", err)
	}

	if !chtUsers.Contains(userID) {
		return chatDomain.ErrChatHaveNoUser
	}

	return nil
}

func (c *chatService) AddUserToChat(ctx context.Context, ownerID int64, chatID int64, userID int64) error {
	chtUsers, err := c.chat.GetChatUsers(ctx, chatID)
	if err != nil {
		if errors.Is(err, chatDomain.ErrChatNotFound) {
			return err
		}

		return fmt.Errorf("Chat.Service.AddUserToChat failed to get chat users by id: %w", err)
	}

	if !chtUsers.IsOwner(ownerID) {
		return chatDomain.ErrUserNotOwner
	}

	_, err = c.auth.GetUserById(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return err
		}

		return fmt.Errorf("Chat.Service.AddUserToChat failed to get user id: %w", err)
	}

	err = c.chat.AddUserToChat(ctx, chatID, userID, chatDomain.Member)
	if err != nil {
		return fmt.Errorf("Chat.Service.AddUserToChat failed to add user to chat: %w", err)
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
	currentUser, err := c.auth.GetUserById(ctx, userID)
	if err != nil {
		return fmt.Errorf("Chat.Service.StartMessaging failed to get user id:%w", err)
	}

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

		if msg.Message == "" {
			continue
		}

		domainMsg := chatDomain.Message{
			UserID: currentUser.ID,
			ChatID: chatID,
			Msg:    msg.Message,
			Login:  currentUser.Login.String(),
		}

		err = c.chat.SaveMessage(ctx, domainMsg)
		if err != nil {
			return fmt.Errorf("Chat.Service.StartMessaging failed to save msg:%w", err)
		}

		err = c.producer.Produce(ctx, userUuid, domainMsg)
		if err != nil {
			fmt.Println("Chat.Service.StartMessaging failed to save msg:", err)
		}
	}

	return nil
}

func (c *chatService) SendMessage(ctx context.Context, uuid string, msg chatDomain.Message) {
	go func() {
		c.sendMsg(uuid, msg)
	}()
}

func (c *chatService) sendMsg(uuid string, msg chatDomain.Message) {
	c.mu.Lock()
	allConnections := c.chatIdToStream[msg.ChatID]
	c.mu.Unlock()

	for _, connect := range allConnections {
		if connect.userUuid != uuid {
			connect.stream.Send(&chatv1.ChatMessageResponse{
				Message: msg.Msg,
				UserId:  msg.UserID,
				ChatId:  msg.ChatID,
				Login:   msg.Login,
			})
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
