package tokenizer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/monobearotaku/online-chat-api/internal/domain/token"
	"github.com/monobearotaku/online-chat-api/internal/domain/token/data"
	"github.com/monobearotaku/online-chat-api/internal/pkg/slices/timeconv"
)

const (
	hmacSampleSecret  = "HelloWorld"
	validUntilKey     = "validUntil"
	tokenValidityTime = 72 * time.Hour
)

type tokenizer struct{}

func NewTokenizer() Tokenizer {
	return &tokenizer{}
}

func (t *tokenizer) CreateToken(ctx context.Context, tokenData data.TokenData) (token.Token, error) {
	tokenData[validUntilKey] = timeconv.TimeToString(time.Now().Add(tokenValidityTime))

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenData.ToMap())

	tokenString, err := tkn.SignedString([]byte(hmacSampleSecret))
	if err != nil {
		return "", err
	}

	return token.Token(tokenString), nil
}

func (t *tokenizer) ValidateToken(ctx context.Context, strToken token.Token) error {
	jwtClaims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(strToken.String(), &jwtClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(hmacSampleSecret), nil
	})
	if err != nil {
		return token.ErrInvalidToken
	}

	mapClaims := data.TokenDataFromMap(jwtClaims)
	valid, ok := mapClaims[validUntilKey]

	if !ok || timeconv.StringToTime(valid).Before(time.Now()) {
		return token.ErrTokenExpired
	}

	return nil
}

func (t *tokenizer) extractTokenData(ctx context.Context, strToken token.Token) (data.TokenData, error) {
	jwtClaims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(strToken.String(), &jwtClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(hmacSampleSecret), nil
	})
	if err != nil {
		return data.TokenData{}, token.ErrInvalidToken
	}

	return data.TokenDataFromMap(jwtClaims), nil
}

func (t *tokenizer) ValidateAndExtractData(ctx context.Context, tokenStr token.Token) (data.TokenData, error) {
	err := t.ValidateToken(ctx, tokenStr)
	if err != nil {
		return data.TokenData{}, err
	}

	tokenData, err := t.extractTokenData(ctx, tokenStr)
	if err != nil {
		if errors.Is(err, token.ErrInvalidToken) {
			return data.TokenData{}, err
		}

		return data.TokenData{}, fmt.Errorf("Tokenizer.Service.ValidateAndExtractData failed to extract token data: %w", err)
	}

	return tokenData, nil
}
