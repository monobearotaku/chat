package producer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/monobearotaku/online-chat-api/internal/domain/event"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(broker),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
		},
	}
}

func (p *Producer) Produce(ctx context.Context, message event.Event) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	kafkaMsg := kafka.Message{
		Key:   []byte(message.Key()),
		Value: msgBytes,
	}

	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		return err
	}

	return nil
}
