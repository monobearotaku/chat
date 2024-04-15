package chat

import "errors"

var (
	ErrChatNotFound      = errors.New("Chat not found")
	ErrChatAlreadyExists = errors.New("Chat already exists")
	ErrChatHaveNoUser    = errors.New("User not in chat")
	ErrChatInvalidName   = errors.New("Invalid chat name")
	ErrUserNotOwner      = errors.New("User is not an owner")
)

type Chat struct {
	ID   int64
	Name string
}

type ChatUser struct {
	UserID int64
	Role   Role
}

type ChatUsers struct {
	ID    int64
	Users []ChatUser
}

func (cu ChatUsers) Contains(userID int64) bool {
	for _, item := range cu.Users {
		if item.UserID == userID {
			return true
		}
	}

	return false
}

func (cu ChatUsers) IsOwner(userID int64) bool {
	for _, item := range cu.Users {
		if item.UserID == userID && item.Role == Owner {
			return true
		}
	}

	return false
}
