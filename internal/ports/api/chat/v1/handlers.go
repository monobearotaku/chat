package v1

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"

	"github.com/monobearotaku/online-chat-api/internal/domain/token"
	chatv1 "github.com/monobearotaku/online-chat-api/proto/chat/v1"
)

func (c *ChatV1) JoinChat(ctx context.Context, req *chatv1.JoinChatRequest) (*chatv1.JoinChatResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md["authentication"]

	if len(values) < 1 {
		return nil, token.ErrInvalidToken
	}

	data, err := c.tokenizer.ValidateAndExtractData(ctx, token.Token(values[0]))
	if err != nil {
		return nil, err
	}

	userID, err := strconv.ParseInt(data["userID"], 10, 64)
	if err != nil {
		return nil, token.ErrInvalidToken
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

	md, _ := metadata.FromIncomingContext(ctx)
	values := md["session"]

	if len(values) < 1 {
		return token.ErrInvalidToken
	}

	data, err := c.tokenizer.ValidateAndExtractData(ctx, token.Token(values[0]))
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(data["userID"], 10, 64)
	if err != nil {
		return token.ErrInvalidToken
	}

	chatID, err := strconv.ParseInt(data["chatID"], 10, 64)
	if err != nil {
		return token.ErrInvalidToken
	}

	err = c.chatService.StartMessaging(ctx, userID, chatID, stream)
	if err != nil {
		return err
	}

	return nil
}

func (c *ChatV1) CreateChat(ctx context.Context, req *chatv1.CreateChatRequest) (*chatv1.CreateChatResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md["authentication"]
	if len(values) < 1 {
		return nil, token.ErrInvalidToken
	}

	data, err := c.tokenizer.ValidateAndExtractData(ctx, token.Token(values[0]))
	if err != nil {
		return nil, err
	}

	userID, err := strconv.ParseInt(data["userID"], 10, 64)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	newChat, err := c.chatService.CreateChat(ctx, userID, req.ChatName)
	if err != nil {
		return nil, err
	}

	return &chatv1.CreateChatResponse{
		ChatId: newChat.ID,
	}, nil
}
