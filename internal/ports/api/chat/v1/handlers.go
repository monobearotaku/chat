package v1

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"

	"github.com/monobearotaku/online-chat-api/internal/domain/token"
	chatv1 "github.com/monobearotaku/online-chat-api/proto/chat/v1"
)

func (c *ChatV1) extractUserId(ctx context.Context) (int64, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md["authentication"]

	if len(values) < 1 {
		return 0, token.ErrInvalidToken
	}

	data, err := c.tokenizer.ValidateAndExtractData(ctx, token.Token(values[0]))
	if err != nil {
		return 0, err
	}

	userID, err := strconv.ParseInt(data["userID"], 10, 64)
	if err != nil {
		return 0, token.ErrInvalidToken
	}

	return userID, nil
}

func (c *ChatV1) extrextractChatAndUserIdFromSession(ctx context.Context) (int64, int64, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md["session"]

	if len(values) < 1 {
		return 0, 0, token.ErrInvalidToken
	}

	data, err := c.tokenizer.ValidateAndExtractData(ctx, token.Token(values[0]))
	if err != nil {
		return 0, 0, err
	}

	chatID, err := strconv.ParseInt(data["chatID"], 10, 64)
	if err != nil {
		return 0, 0, token.ErrInvalidToken
	}

	userID, err := strconv.ParseInt(data["userID"], 10, 64)
	if err != nil {
		return 0, 0, token.ErrInvalidToken
	}

	return chatID, userID, nil
}

func (c *ChatV1) JoinChat(ctx context.Context, req *chatv1.JoinChatRequest) (*chatv1.JoinChatResponse, error) {
	userID, err := c.extractUserId(ctx)
	if err != nil {
		return nil, err
	}

	tkn, err := c.chatService.JoinChat(ctx, userID, req.ChatId)
	if err != nil {
		return nil, err
	}

	return &chatv1.JoinChatResponse{
		Session: tkn.String(),
	}, nil
}

func (c *ChatV1) ConnectToChat(stream chatv1.ChatService_ConnectToChatServer) error {
	ctx := stream.Context()

	chatID, userID, err := c.extrextractChatAndUserIdFromSession(ctx)
	if err != nil {
		return err
	}

	err = c.chatService.StartMessaging(ctx, userID, chatID, stream)
	if err != nil {
		return err
	}

	return nil
}

func (c *ChatV1) CreateChat(ctx context.Context, req *chatv1.CreateChatRequest) (*chatv1.CreateChatResponse, error) {
	userID, err := c.extractUserId(ctx)
	if err != nil {
		return nil, err
	}

	newChat, err := c.chatService.CreateChat(ctx, userID, req.ChatName)
	if err != nil {
		return nil, err
	}

	return &chatv1.CreateChatResponse{
		ChatId: newChat.ID,
	}, nil
}

func (c *ChatV1) AddUserToChat(ctx context.Context, req *chatv1.AddUserToChatRequest) (*chatv1.AddUserToChatResponse, error) {
	ownerID, err := c.extractUserId(ctx)
	if err != nil {
		return nil, err
	}

	err = c.chatService.AddUserToChat(ctx, ownerID, req.ChatId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &chatv1.AddUserToChatResponse{}, nil
}
