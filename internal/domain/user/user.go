package user

import "github.com/monobearotaku/online-chat-api/internal/domain"

type User struct {
	ID           int64
	Login        domain.Login
	PasswordHash domain.Password
}
