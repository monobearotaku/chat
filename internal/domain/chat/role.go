package chat

type Role string

const (
	Owner  Role = "owner"
	Member Role = "member"
)

func (r Role) String() string {
	return string(r)
}
