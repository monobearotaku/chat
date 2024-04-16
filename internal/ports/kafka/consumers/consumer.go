package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/monobearotaku/online-chat-api/internal/config"
	chatDomain "github.com/monobearotaku/online-chat-api/internal/domain/chat"
	"github.com/monobearotaku/online-chat-api/internal/pkg/slices"
	"github.com/monobearotaku/online-chat-api/internal/service/chat"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	r           *kafka.Reader
	chatService chat.Service
}

func NewConsumer(config config.Config, chatService chat.Service) *Consumer {
	return &Consumer{
		r: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     slices.FromElenent(config.Kafka.Broker),
			Topic:       config.Kafka.Topic,
			GroupID:     "chat",
			StartOffset: kafka.LastOffset,
		}),
		chatService: chatService,
	}
}

func (c *Consumer) Consume(ctx context.Context) {
	for {
		msg, err := c.r.ReadMessage(ctx)
		if err != nil {
			fmt.Println("Error reading message:", err)
			continue
		}

		chatMessage := chatDomain.Message{}

		err = json.Unmarshal(msg.Value, &chatMessage)
		if err != nil {
			fmt.Println("Error unmarshaling message:", err)
			continue
		}

		c.chatService.SendMessage(ctx, string(msg.Key), chatMessage)
	}
}
