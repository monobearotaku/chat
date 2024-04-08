package domain

import (
	"errors"
	"unicode/utf8"
)

var (
	ErrLoginTooShort = errors.New("Login too short (must be more than 4 characters)")
	ErrLoginTooLong  = errors.New("Login too long must be less that 20 characters")
)

type Login string

func (l Login) String() string {
	return string(l)
}

func (l Login) Validate() error {
	if utf8.RuneCountInString(string(l)) <= 4 {
		return ErrLoginTooShort
	}

	if utf8.RuneCountInString(string(l)) > 20 {
		return ErrLoginTooLong
	}

	return nil
}
