package client

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
	internalSync "github.com/john6604/yata-collaborative-editor/internal/sync"
)

func RemoteMessageLoop(doc *document.Document, conn *websocket.Conn, mutex *sync.Mutex) {

	for {

		messageType, message, errMessage := conn.ReadMessage()

		if errMessage != nil {
			break
		}

		if messageType != websocket.TextMessage {
			continue
		}

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

		convertedOperation, errConversion := internalSync.ConvertUpdateOperation(updatePayload)

		if errConversion != nil {
			fmt.Println(errConversion)
			continue
		}

		mutex.Lock()

		errApply := internalSync.ApplyConvertedOperation(doc, convertedOperation)

		if errApply != nil {
			fmt.Println(errApply)
			mutex.Unlock()
			continue
		}

		fmt.Println(doc.String())

		mutex.Unlock()
	}
}
