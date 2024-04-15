package producer

import (
	"context"
)

type Producer interface {
	Produce(ctx context.Context, key string, msg interface{}) error
}
