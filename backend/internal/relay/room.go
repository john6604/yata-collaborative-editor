package relay

type Room struct {
	id      string
	clients map[string]*ClientSession
}

func NewRoom(id string) *Room {
	room := Room{id: id}
	room.clients = map[string]*ClientSession{}
	return &room
}
