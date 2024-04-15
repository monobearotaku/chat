package producer

import (
	"context"
	"encoding/json"

	"github.com/monobearotaku/online-chat-api/internal/config"
	"github.com/monobearotaku/online-chat-api/internal/pkg/slices"
	"github.com/segmentio/kafka-go"
)

type producer struct {
	w *kafka.Writer
}

func NewProducer(config config.Config) Producer {
	return &producer{
		w: kafka.NewWriter(
			kafka.WriterConfig{
				Brokers:  slices.FromElenent(config.Kafka.Broker),
				Topic:    config.Kafka.Topic,
				Balancer: &kafka.RoundRobin{},
			},
		),
	}
}

func (p *producer) Produce(ctx context.Context, key string, msg interface{}) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	kafkaMsg := kafka.Message{
		Key:   []byte(key),
		Value: bytes,
	}

	return p.w.WriteMessages(ctx, kafkaMsg)
}
