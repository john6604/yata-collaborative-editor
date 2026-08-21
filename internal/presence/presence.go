package presence

type PresenceUser struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func NewPresenceUser(id string, name string) *PresenceUser {
	presenceUser := PresenceUser{
		Id:   id,
		Name: name,
	}

	return &presenceUser
}
