package client

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
	"github.com/john6604/yata-collaborative-editor/internal/sync"
)

func RemoteMessageLoop(doc *document.Document, conn *websocket.Conn) {

	for {

		messageType, message, errMessage := conn.ReadMessage()

		if errMessage != nil {
			break
		}

		if messageType != websocket.TextMessage {
			continue
		}

		fmt.Println(string(message))

		version, typeMessage, envelope, errEnvelope := protocol.DecodeEnvelope(message)

		if errEnvelope != nil {
			fmt.Println(errEnvelope)
			continue
		}

		if version != protocol.SupportedVersion {
			continue
		}

		if typeMessage != protocol.TypeUpdate {
			continue
		}

		updatePayload, errPayload := protocol.DecodeUpdate(envelope)

		if errPayload != nil {
			fmt.Println(errPayload)
			continue
		}

		convertedOperation, errConversion := sync.ConvertUpdateOperation(updatePayload)

		if errConversion != nil {
			fmt.Println(errConversion)
			continue
		}

		errApply := sync.ApplyConvertedOperation(doc, convertedOperation)

		if errApply != nil {
			fmt.Println(errApply)
			continue
		}

		fmt.Println(doc.String())
	}
}
