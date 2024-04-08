package event

type MessageFromUser struct {
	UserID  int64
	Login   string
	Message string
}
