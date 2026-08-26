package relay

import (
	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func SendErrorMessage(code string, message string, conn *websocket.Conn) {

	bytes, err := protocol.EncodeErrorPayload(code, message)

	if err != nil {
		return
	}

	errSend := conn.WriteMessage(websocket.TextMessage, bytes)

	if errSend != nil {
		return
	}

}
