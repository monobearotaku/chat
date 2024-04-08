package credentials

import (
	"errors"

	"github.com/monobearotaku/online-chat-api/internal/domain"
)

var (
	ErrWrongCreds = errors.New("Login or Password are incorrect")
)

type Credentials struct {
	Login    domain.Login
	Password domain.Password
}

func NewCredentials(login, password string) Credentials {
	return Credentials{
		Login:    domain.Login(login),
		Password: domain.Password(password),
	}
}

func (c Credentials) Validate() error {
	return errors.Join(c.Login.Validate(), c.Password.Validate())
}
