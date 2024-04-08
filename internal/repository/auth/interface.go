package auth

import (
	"context"

	"github.com/monobearotaku/online-chat-api/internal/domain"
	"github.com/monobearotaku/online-chat-api/internal/domain/credentials"
	"github.com/monobearotaku/online-chat-api/internal/domain/user"
	"github.com/monobearotaku/online-chat-api/internal/postgres"
)

type Repo interface {
	WithTx(tx postgres.Tx) Repo
	CreateUser(ctx context.Context, cred credentials.Credentials) error
	GetUser(ctx context.Context, login domain.Login) (user.User, error)
}
