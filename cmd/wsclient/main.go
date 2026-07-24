package main

import (
	"fmt"
	"os"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/client"
	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/sync"
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

	errInsert, operationID := doc.InsertElement(0, 'H')

	if errInsert != nil {
		fmt.Println(errInsert)
		os.Exit(0)
	}

	fmt.Println("Local document: ", doc.String())

	operation := doc.InsertLog[operationID]

	if operation == nil {
		fmt.Println("Operation is empty")
		os.Exit(0)
	}

	dataOperation, errOperation := sync.EncodeInsertOperation(*operation)

	if errOperation != nil {
		fmt.Println(errOperation)
		os.Exit(0)
	}

	errSending := conn.WriteMessage(websocket.TextMessage, dataOperation)

	if errSending != nil {
		fmt.Println(errSending)
		os.Exit(0)
	}

	fmt.Println("Waiting for remote messages...")

	client.RemoteMessageLoop(doc, conn)

}
