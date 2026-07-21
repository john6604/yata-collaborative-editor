package relay

import "github.com/gorilla/websocket"

type ClientSession struct {
	clientID  string
	roomID    string
	webSocket *websocket.Conn
}

func NewClientSession(clientID string, roomID string, conn *websocket.Conn) *ClientSession {
	clientSession := ClientSession{clientID: clientID}
	clientSession.roomID = roomID
	clientSession.webSocket = conn
	return &clientSession
}
