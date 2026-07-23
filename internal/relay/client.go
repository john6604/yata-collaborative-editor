package relay

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ClientSession struct {
	clientID   string
	roomID     string
	webSocket  *websocket.Conn
	writeMutex sync.Mutex
}

func NewClientSession(clientID string, roomID string, conn *websocket.Conn) *ClientSession {
	clientSession := ClientSession{clientID: clientID}
	clientSession.roomID = roomID
	clientSession.webSocket = conn
	return &clientSession
}

func (c *ClientSession) Send(message []byte) error {

	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	errSend := c.webSocket.WriteMessage(websocket.TextMessage, message)

	if errSend != nil {
		return errSend
	}

	return nil
}
