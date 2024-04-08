package tokenizer

import (
	"context"

	"github.com/monobearotaku/online-chat-api/internal/domain/token"
	"github.com/monobearotaku/online-chat-api/internal/domain/token/data"
)

type Tokenizer interface {
	CreateToken(context.Context, data.TokenData) (token.Token, error)
	ValidateToken(context.Context, token.Token) error
	extractTokenData(context.Context, token.Token) (data.TokenData, error)
	ValidateAndExtractData(context.Context, token.Token) (data.TokenData, error)
}
