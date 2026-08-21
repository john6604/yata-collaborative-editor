package relay

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

type ClientSession struct {
	clientID   string
	roomID     string
	name       string
	webSocket  *websocket.Conn
	writeMutex sync.Mutex
}

func NewClientSession(clientID string, roomID string, name string, conn *websocket.Conn) *ClientSession {
	clientSession := ClientSession{clientID: clientID}
	clientSession.roomID = roomID
	clientSession.name = name
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

func (c *ClientSession) SendErrorMessage(code string, message string) error {

	bytes, err := protocol.EncodeErrorPayload(code, message)

	if err != nil {
		return err
	}

	errSend := c.Send(bytes)

	if errSend != nil {
		return errSend
	}

	return nil
}
