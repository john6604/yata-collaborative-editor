package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
	internalSync "github.com/john6604/yata-collaborative-editor/internal/sync"
)

type ConnectionState uint8

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateSyncing
	StateOnline
	StateClosing
)

type CollaborativeClient struct {
	ServerURL string
	RoomID    string
	ClientID  string

	Doc      *document.Document
	DocMutex sync.Mutex

	Conn      *websocket.Conn
	State     ConnectionState
	ConnMutex sync.RWMutex

	WriteMutex sync.Mutex

	OfflineQueue [][]byte

	ReconnectSignal chan struct{}
	StopSignal      chan struct{}
	StopOnce        sync.Once
}

func NewCollaborativeClient(serverURL string, roomID string, clientID string) (*CollaborativeClient, error) {

	formattedServer := strings.TrimSpace(serverURL)
	formattedRoom := strings.TrimSpace(roomID)
	formattedClient := strings.TrimSpace(clientID)

	if formattedServer == "" || formattedRoom == "" || formattedClient == "" {
		return nil, errors.New("invalid flags")
	}

	client := CollaborativeClient{
		ServerURL:       formattedServer,
		RoomID:          formattedRoom,
		ClientID:        formattedClient,
		Doc:             document.NewDocument(),
		State:           StateDisconnected,
		OfflineQueue:    make([][]byte, 0),
		ReconnectSignal: make(chan struct{}, 1),
		StopSignal:      make(chan struct{}),
	}

	return &client, nil
}

func (client *CollaborativeClient) GetState() ConnectionState {
	client.ConnMutex.RLock()
	clientState := client.State
	client.ConnMutex.RUnlock()

	return clientState
}

func (client *CollaborativeClient) ConnectionSnapshot() (*websocket.Conn, ConnectionState) {
	client.ConnMutex.RLock()
	conn := client.Conn
	state := client.State
	client.ConnMutex.RUnlock()

	return conn, state
}

func (client *CollaborativeClient) SetConnection(conn *websocket.Conn, state ConnectionState) {
	client.ConnMutex.Lock()
	client.Conn = conn
	client.State = state
	client.ConnMutex.Unlock()
}

func (client *CollaborativeClient) EnqueueOffline(message []byte) {

	messageCopy := append([]byte(nil), message...)

	client.ConnMutex.Lock()
	client.OfflineQueue = append(client.OfflineQueue, messageCopy)
	client.ConnMutex.Unlock()
}

func (client *CollaborativeClient) OfflineQueueLength() int {
	client.ConnMutex.RLock()
	length := len(client.OfflineQueue)
	client.ConnMutex.RUnlock()

	return length
}

func (client *CollaborativeClient) SendOrQueue(message []byte) error {

	if len(message) == 0 {
		return errors.New("empty message")
	}

	conn, state := client.ConnectionSnapshot()

	if state != StateOnline || conn == nil {
		client.EnqueueOffline(message)
		return nil
	}

	client.WriteMutex.Lock()
	errSend := conn.WriteMessage(websocket.TextMessage, message)
	client.WriteMutex.Unlock()

	if errSend == nil {
		return nil
	}

	client.EnqueueOffline(message)
	client.MarkDisconnected(conn)

	return nil
}

func (client *CollaborativeClient) MarkDisconnected(conn *websocket.Conn) {

	shouldReconnect := false

	if conn == nil {
		return
	}

	client.ConnMutex.Lock()

	if client.Conn == conn {
		client.Conn = nil
		if client.State != StateClosing {
			client.State = StateDisconnected
			shouldReconnect = true
		}
	}

	client.ConnMutex.Unlock()

	conn.Close()

	if shouldReconnect {
		client.RequestReconnect()
	}
}

func (client *CollaborativeClient) RequestReconnect() {

	ch := client.ReconnectSignal

	select {
	case ch <- struct{}{}:
		fmt.Println("Signal sent")
	default:

	}
}

func (client *CollaborativeClient) BeginReconnect() bool {

	client.ConnMutex.Lock()

	if client.State == StateDisconnected && client.Conn == nil {
		client.State = StateConnecting
		client.ConnMutex.Unlock()
		return true
	}
	client.ConnMutex.Unlock()

	return false
}

func (client *CollaborativeClient) ReconnectLoop() {

	loop := true

	for loop {

		select {
		case <-client.ReconnectSignal:
			beginReconnect := client.BeginReconnect()
			if beginReconnect {
				_, err := client.ReconnectOnce()
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("Reconnected and joined...")
				}
			}
		case <-client.StopSignal:
			loop = false
		}
	}
}

func (client *CollaborativeClient) DialAndJoin() (*websocket.Conn, error) {

	conn, _, err := websocket.DefaultDialer.Dial(client.ServerURL, nil)

	if err != nil {
		return nil, err
	}

	fmt.Printf("Connected to relay server...\n")

	msg := protocol.JoinPayload{
		RoomID:   client.RoomID,
		ClientID: client.ClientID,
	}

	joinPayload, errPayload := json.Marshal(msg)

	if errPayload != nil {
		conn.Close()
		return nil, errPayload
	}

	envelope := protocol.Envelope{
		Version:     protocol.SupportedVersion,
		MessageType: protocol.TypeJoin,
		Payload:     joinPayload,
	}

	envelopeBytes, errEnvelope := json.Marshal(envelope)

	if errEnvelope != nil {
		conn.Close()
		return nil, errEnvelope
	}

	client.WriteMutex.Lock()

	errSend := conn.WriteMessage(websocket.TextMessage, envelopeBytes)

	client.WriteMutex.Unlock()

	if errSend != nil {
		conn.Close()
		return nil, errSend
	}

	messageType, message, errResponse := conn.ReadMessage()

	if errResponse != nil {
		conn.Close()
		return nil, errResponse
	}

	validJoinAck := client.ValidateJoinAck(messageType, message)

	if validJoinAck != nil {
		conn.Close()
		return nil, validJoinAck
	}

	return conn, nil

}

func (client *CollaborativeClient) ValidateJoinAck(messageType int, message []byte) error {

	if messageType != websocket.TextMessage {
		return errors.New("unsupported message")
	}

	version, typeMessage, payload, errDecode := protocol.DecodeEnvelope(message)

	if errDecode != nil {
		return errDecode
	}

	if version != protocol.SupportedVersion {
		return errors.New("Unsupported version")
	}

	if typeMessage != protocol.TypeJoinAck {
		return errors.New("Must be join_ack package")
	}

	room, clientID, errJoinAck := protocol.DecodeJoinAck(payload)

	if errJoinAck != nil {
		return errJoinAck
	}

	if room != client.RoomID {
		return errors.New("Unknown room")
	}

	if clientID != client.ClientID {
		return errors.New("Unknown client")
	}

	return nil
}

func (client *CollaborativeClient) ReconnectOnce() (*websocket.Conn, error) {

	conn, err := client.DialAndJoin()

	if err != nil {
		client.ConnMutex.Lock()
		if client.State == StateConnecting {
			client.State = StateDisconnected
			client.Conn = nil
		}
		client.ConnMutex.Unlock()
		return nil, err
	}

	client.ConnMutex.Lock()
	if client.State != StateConnecting {
		client.ConnMutex.Unlock()
		conn.Close()
		return nil, errors.New("could not install reconnected connection")
	}

	client.State = StateSyncing
	client.Conn = conn

	client.ConnMutex.Unlock()

	return conn, nil
}

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
