package client

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
	internalSync "github.com/john6604/yata-collaborative-editor/internal/sync"
)

func RemoteMessageLoop(doc *document.Document, conn *websocket.Conn, mutex *sync.Mutex, writeMutex *sync.Mutex) {

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

		var operationEnvelope protocol.OperationEnvelope

		errOpEnvelope := json.Unmarshal(updatePayload.Operation, &operationEnvelope)

		if errOpEnvelope != nil {
			fmt.Println(errOpEnvelope)
			continue
		}

		switch operationEnvelope.Type {
		case protocol.OpInsert, protocol.OpDelete:
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
		case protocol.OpSync1:
			fmt.Println("received sync_step1")
			var sync1Operation protocol.SyncOp1

			err := json.Unmarshal(updatePayload.Operation, &sync1Operation)

			if err != nil {
				fmt.Println(err)
				continue
			}

			remoteVector := internalSync.Vector{
				StateVectors: sync1Operation.VectorState,
				DeleteSet:    sync1Operation.DeleteSet,
			}

			mutex.Lock()

			missingInserts, missingDeletes := internalSync.ComputeDelta(*doc, remoteVector)
			delta := internalSync.ComputeSerializedDelta(*doc, missingInserts, missingDeletes)

			mutex.Unlock()

			envelope, errEnvelope := internalSync.EncodeSyncStep2(delta)

			if errEnvelope != nil {
				fmt.Println(errEnvelope)
				continue
			}

			writeMutex.Lock()

			errSend := conn.WriteMessage(websocket.TextMessage, envelope)
			if errSend != nil {
				fmt.Println(errSend)
				writeMutex.Unlock()
				continue
			}

			writeMutex.Unlock()

		case protocol.OpSync2:
			fmt.Println("received sync_step2")

			var sync2Operation protocol.SyncOp2

			err := json.Unmarshal(updatePayload.Operation, &sync2Operation)

			if err != nil {
				fmt.Println(err)
				continue
			}

			if sync2Operation.Delta == nil {
				fmt.Println("missing field")
				continue
			}

			mutex.Lock()

			errDelta := doc.IntegrateDelta(*sync2Operation.Delta)

			if errDelta != nil {
				fmt.Println(errDelta)
				mutex.Unlock()
				continue
			}

			docContent := doc.String()

			mutex.Unlock()

			fmt.Println(docContent)

		default:
			fmt.Println("unsupported operation")
			continue
		}
	}
}
