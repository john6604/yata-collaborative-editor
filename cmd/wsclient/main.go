package main

import (
	"fmt"
	"os"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/client"
	"github.com/john6604/yata-collaborative-editor/internal/document"
)

func main() {

	doc := document.NewDocument()

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8181/ws", nil)

	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	fmt.Println("Connected to relay...")

	defer conn.Close()

	jsonStr := `{"version":1,"type":"join","payload":{"room":"room-1","client_id":"client-B"}}`
	data := []byte(jsonStr)

	errSend := conn.WriteMessage(websocket.TextMessage, data)

	if errSend != nil {
		fmt.Println(errSend)
		os.Exit(0)
	}

	_, message, errMessage := conn.ReadMessage()

	if errMessage != nil {
		fmt.Println(errMessage)
		os.Exit(0)
	}

	fmt.Println(string(message))

	fmt.Println("Waiting for remote messages...")

	client.RemoteMessageLoop(doc, conn)

}
