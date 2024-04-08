package token

import "errors"

var (
	ErrInvalidToken = errors.New("Invalid JWT token")
	ErrTokenExpired = errors.New("JWT token expired")
)

type Token string

func (t Token) String() string {
	return string(t)
}
