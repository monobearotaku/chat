package chat

import "errors"

var (
	ErrChatNotFound      = errors.New("Chat not found")
	ErrChatAlreadyExists = errors.New("Chat already exists")
	ErrChatHaveNoUser    = errors.New("User not in chat")
	ErrChatInvalidName   = errors.New("Invalid chat name")
)

type Chat struct {
	ID      int64
	Name    string
	UserIDs []int64
	OwnerID int64
}
