package relay

type ClientSession struct {
	clientID string
	roomID   string
}

func NewClientSession(clientID string, roomID string) *ClientSession {
	clientSession := ClientSession{clientID: clientID}
	clientSession.roomID = roomID
	return &clientSession
}
