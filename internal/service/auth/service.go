package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/monobearotaku/online-chat-api/internal/domain"
	"github.com/monobearotaku/online-chat-api/internal/domain/credentials"
	"github.com/monobearotaku/online-chat-api/internal/domain/token"
	"github.com/monobearotaku/online-chat-api/internal/postgres"
	"github.com/monobearotaku/online-chat-api/internal/repository/auth"
	"github.com/monobearotaku/online-chat-api/internal/service/tokenizer"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	auth       auth.Repo
	tokenizer  tokenizer.Tokenizer
	txBeginner postgres.TxBeginner
}

func NewAuthService(auth auth.Repo, tokenizer tokenizer.Tokenizer, txBeginner postgres.TxBeginner) Service {
	return &authService{
		auth:       auth,
		tokenizer:  tokenizer,
		txBeginner: txBeginner,
	}
}

func (s *authService) SignUp(ctx context.Context, cred credentials.Credentials) (newToken token.Token, err error) {
	cred, err = s.securePassword(cred)
	if err != nil {
		return "", fmt.Errorf("Auth.Service.SignUp hashin pass: %w", err)
	}

	tx, err := s.txBeginner.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", fmt.Errorf("Auth.Service.SignUp begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			if err := tx.Commit(ctx); err != nil {
				_ = tx.Rollback(ctx)
			}
		}
	}()

	err = s.auth.WithTx(tx).CreateUser(ctx, cred)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return "", domain.ErrAlreadyExists
		}

		return "", fmt.Errorf("Auth.Service.SignUp creating user: %w", err)
	}

	usr, err := s.auth.WithTx(tx).GetUser(ctx, cred.Login)
	if err != nil {
		return "", fmt.Errorf("Auth.Service.SignUp geting user: %w", err)
	}

	newToken, err = s.tokenizer.CreateToken(ctx, map[string]string{
		"userID": strconv.FormatInt(usr.ID, 10),
		"login":  usr.Login.String(),
	})

	if err != nil {
		return "", fmt.Errorf("Auth.Service.SignUp creating token: %w", err)
	}

	return newToken, nil
}

func (s *authService) SignIn(ctx context.Context, cred credentials.Credentials) (newToken token.Token, err error) {
	usr, err := s.auth.GetUser(ctx, cred.Login)
	if err != nil {
		return "", fmt.Errorf("Auth.Service.SignIn getting user data: %w", err)
	}

	if !s.comparePassword(cred.Password, usr.PasswordHash) {
		return "", credentials.ErrWrongCreds
	}

	newToken, err = s.tokenizer.CreateToken(ctx, map[string]string{
		"userID": strconv.FormatInt(usr.ID, 10),
		"login":  usr.Login.String(),
	})

	if err != nil {
		return "", fmt.Errorf("Auth.Service.SignUp creating token: %w", err)
	}

	return newToken, nil
}

func (s *authService) securePassword(cred credentials.Credentials) (credentials.Credentials, error) {
	bytes, err := bcrypt.GenerateFromPassword(cred.Password.Bytes(), 14)
	if err != nil {
		return cred, err
	}

	return credentials.Credentials{
		Login:    cred.Login,
		Password: domain.Password(bytes),
	}, nil
}

func (s *authService) comparePassword(password, hash domain.Password) bool {
	return bcrypt.CompareHashAndPassword(hash.Bytes(), password.Bytes()) == nil
}
