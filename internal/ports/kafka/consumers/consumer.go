package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/google/uuid"
	"github.com/monobearotaku/online-chat-api/internal/config"
	chatDomain "github.com/monobearotaku/online-chat-api/internal/domain/chat"
	"github.com/monobearotaku/online-chat-api/internal/pkg/slices"
	"github.com/monobearotaku/online-chat-api/internal/service/chat"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	r           *kafka.Reader
	chatService chat.Service
	logger      log.Logger
}

func NewConsumer(config config.Config, chatService chat.Service, logger log.Logger) *Consumer {
	return &Consumer{
		r: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     slices.FromElenent(config.Kafka.Broker),
			Topic:       config.Kafka.Topic,
			GroupID:     uuid.NewString(),
			StartOffset: kafka.LastOffset,
		}),
		chatService: chatService,
		logger:      logger,
	}
}

func (c *Consumer) Consume(ctx context.Context) {
	for {
		msg, err := c.r.ReadMessage(ctx)
		if err != nil {
			level.Error(c.logger).Log("error", fmt.Errorf("error reading message:%w", err))
			continue
		}

		chatMessage := chatDomain.Message{}

		err = json.Unmarshal(msg.Value, &chatMessage)
		if err != nil {
			level.Error(c.logger).Log("error", fmt.Errorf("error unmarshaling message:%w", err))
			continue
		}

		go func(key string, msg chatDomain.Message) {
			c.chatService.SendMessage(ctx, key, msg)
		}(string(msg.Key), chatMessage)
	}
}
