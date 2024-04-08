package domain

import (
	"errors"
	"unicode/utf8"
)

var (
	ErrPasswordTooShort = errors.New("Password too short (must be more than 7 characters)")
	ErrPasswordTooLong  = errors.New("Password too long must be less that 30 characters")
)

type Password string

func (p Password) String() string {
	return string(p)
}

func (p Password) Bytes() []byte {
	return []byte(p)
}

func (p Password) Validate() error {
	if utf8.RuneCountInString(string(p)) < 8 {
		return ErrPasswordTooShort
	}

	if utf8.RuneCountInString(string(p)) > 30 {
		return ErrPasswordTooLong
	}

	return nil
}
