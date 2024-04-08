package auth

import (
	"context"

	"github.com/monobearotaku/online-chat-api/internal/domain/credentials"
	"github.com/monobearotaku/online-chat-api/internal/domain/token"
)

type Service interface {
	SignIn(ctx context.Context, cred credentials.Credentials) (token.Token, error)
	SignUp(ctx context.Context, cred credentials.Credentials) (token.Token, error)
}
