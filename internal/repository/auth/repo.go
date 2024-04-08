package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/monobearotaku/online-chat-api/internal/domain"
	"github.com/monobearotaku/online-chat-api/internal/domain/credentials"
	"github.com/monobearotaku/online-chat-api/internal/domain/user"
	"github.com/monobearotaku/online-chat-api/internal/postgres"
)

type authRepo struct {
	db postgres.QueryExecer
}

func NewAuthRepo(db postgres.QueryExecer) Repo {
	return &authRepo{
		db: db,
	}
}

func (a *authRepo) WithTx(tx postgres.Tx) Repo {
	return &authRepo{
		db: tx,
	}
}

func (a *authRepo) CreateUser(ctx context.Context, cred credentials.Credentials) error {
	const query = `
		INSERT INTO users(login, password) 
		VALUES($1, $2)
		ON CONFLICT(login) DO NOTHING
	`

	res, err := a.db.Exec(ctx, query, cred.Login, cred.Password)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return domain.ErrAlreadyExists
	}

	return nil
}

func (a *authRepo) GetUser(ctx context.Context, login domain.Login) (user.User, error) {
	const query = `
		SELECT 
			id,
			login,
			password
		FROM users 
		WHERE login = $1
	`

	usr := user.User{}

	err := a.db.QueryRow(ctx, query, login).Scan(&usr.ID, &usr.Login, &usr.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, domain.ErrNotFound
		}

		return user.User{}, err
	}

	return usr, nil
}
