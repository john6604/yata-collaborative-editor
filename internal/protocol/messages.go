package protocol

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/john6604/yata-collaborative-editor/internal/identifier"
)

var SupportedVersion = 1

var TypeJoin = "join"
var TypeJoinAck = "join_ack"
var TypeError = "error"
var TypeUpdate = "update"
var TypeSnapshot = "snapshot"

var OpInsert = "insert"
var OpDelete = "delete"
var OpSync1 = "sync_step1"
var OpSync2 = "sync_step2"
var OpSnapshot = "snapshot"
var OpPresence = "presence"

var InternalError = "internal_error"
var InternalErrorMessage = "Server internal error."
var Unauthorized = "unauthorized"
var UnauthorizedMessage = "The system could not identify a client."
var NotJoined = "not_joined"
var NotJoinedMessage = "The current client is not joined to any room."
var InvalidPayload = "invalid_payload"
var InvalidPayloadMessage = "The payload is not valid."
var MissingField = "missing_field"
var MissingFieldMessage = "The information is not complete."
var ExpectedMessageCode = "expected_text_message"
var ExpectedMessage = "Only text message type is accepted."
var UnsupportedVersion = "unsupported_version"
var UnsupportedVersionMessage = "The protocol version is not supported."
var ExpectedJoin = "expected_join"
var ExpectedJoinMessage = "The first message must be join."
var DuplicateClient = "duplicate_client"
var DuplicateClientMessage = "The client has already joined the room."
var UnknownMessageCode = "unknown_message_type"
var UnknownMessage = "Unkown message type."

type Envelope struct {
	Version     int             `json:"version"`
	MessageType string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
}

type JoinPayload struct {
	RoomID   string `json:"room"`
	ClientID string `json:"client_id"`
}

type JoinAckPayload struct {
	Room     string `json:"room"`
	ClientID string `json:"client_id"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type UpdatePayload struct {
	Operation json.RawMessage `json:"op"`
}

type OperationEnvelope struct {
	Type string `json:"type"`
}

type InsertOp struct {
	Type      string        `json:"type"`
	NewID     identifier.ID `json:"new_id"`
	OriginID  identifier.ID `json:"origin_id"`
	RightID   identifier.ID `json:"right_id"`
	Character string        `json:"character"`
}

type DeleteOp struct {
	Type     string        `json:"type"`
	TargetID identifier.ID `json:"target_id"`
}

type SyncOp1 struct {
	Type        string           `json:"type"`
	VectorState map[string]int   `json:"vector_state"`
	DeleteSet   map[string][]int `json:"delete_set"`
}

type SyncOp2 struct {
	Type  string `json:"type"`
	Delta *Delta `json:"delta"`
}

type SnapshotOp struct {
	Type  string `json:"type"`
	Delta *Delta `json:"delta"`
}

type PresenceOp struct {
	Type  string   `json:"type"`
	Users []string `json:"users"`
}

func DecodeEnvelope(message []byte) (int, string, json.RawMessage, error) {

	var msg Envelope

	err := json.Unmarshal(message, &msg)

	if err != nil {
		return 0, "", nil, err
	}

	return msg.Version, msg.MessageType, msg.Payload, nil
}

func DecodeUpdate(message []byte) (UpdatePayload, error) {

	var payload UpdatePayload

	err := json.Unmarshal(message, &payload)

	if err != nil {
		return UpdatePayload{}, err
	}

	if len(payload.Operation) == 0 {
		return UpdatePayload{}, errors.New("missing_field")
	}

	var operationEnvelope OperationEnvelope

	errDecode := json.Unmarshal(payload.Operation, &operationEnvelope)

	if errDecode != nil {
		return UpdatePayload{}, errDecode
	}

	formattedType := strings.TrimSpace(operationEnvelope.Type)

	if len(formattedType) == 0 {
		return UpdatePayload{}, errors.New("invalid_payload")
	}

	if formattedType != OpInsert && formattedType != OpDelete && formattedType != OpSync1 && formattedType != OpSync2 && formattedType != OpSnapshot && formattedType != OpPresence {
		return UpdatePayload{}, errors.New("invalid_payload")
	}

	switch formattedType {
	case OpInsert:
		_, errInsert := DecodeInsert(payload)
		if errInsert != nil {
			return UpdatePayload{}, errInsert
		}
	case OpDelete:
		_, errDelete := DecodeDelete(payload)
		if errDelete != nil {
			return UpdatePayload{}, errDelete
		}
	case OpSync1:
		_, errSync1 := DecodeSync1(payload)
		if errSync1 != nil {
			return UpdatePayload{}, errSync1
		}
	case OpSync2:
		_, errSync2 := DecodeSync2(payload)
		if errSync2 != nil {
			return UpdatePayload{}, errSync2
		}
	case OpSnapshot:
		_, errSnapshot := DecodeSnapshot(payload)
		if errSnapshot != nil {
			return UpdatePayload{}, errSnapshot
		}
	case OpPresence:
		_, errPresence := DecodePresence(payload)
		if errPresence != nil {
			return UpdatePayload{}, errPresence
		}
	}

	return payload, nil
}

func DecodeJoin(message json.RawMessage) (string, string, error) {

	var joinPayload JoinPayload

	err := json.Unmarshal(message, &joinPayload)

	if err != nil {
		return "", "", err
	}

	roomID := joinPayload.RoomID
	clientID := joinPayload.ClientID

	formattedRoomID := strings.TrimSpace(roomID)
	formattedClientID := strings.TrimSpace(clientID)

	if formattedClientID == "" {
		return "", "", errors.New("Client ID is empty.")
	}

	if formattedRoomID == "" {
		return "", "", errors.New("Room is empty.")
	}

	return formattedRoomID, formattedClientID, nil
}

func DecodeJoinAck(message json.RawMessage) (string, string, error) {

	var joinPayload JoinAckPayload

	err := json.Unmarshal(message, &joinPayload)

	if err != nil {
		return "", "", err
	}

	roomID := joinPayload.Room
	clientID := joinPayload.ClientID

	formattedRoomID := strings.TrimSpace(roomID)
	formattedClientID := strings.TrimSpace(clientID)

	if formattedClientID == "" {
		return "", "", errors.New("Client ID is empty.")
	}

	if formattedRoomID == "" {
		return "", "", errors.New("Room is empty.")
	}

	return formattedRoomID, formattedClientID, nil
}

func EncodeJoinAck(room string, clientID string) ([]byte, error) {

	joinAckPayload := JoinAckPayload{
		Room:     room,
		ClientID: clientID,
	}

	payload, errPayload := json.Marshal(joinAckPayload)

	if errPayload != nil {
		return nil, errPayload
	}

	envelopeAck := Envelope{
		Version:     SupportedVersion,
		MessageType: TypeJoinAck,
		Payload:     payload,
	}

	ackPackage, errAck := json.Marshal(envelopeAck)

	if errAck != nil {
		return nil, errAck
	}

	return ackPackage, nil
}

func EncodeErrorPayload(code string, message string) ([]byte, error) {

	errorPayload := ErrorPayload{
		Code:    code,
		Message: message,
	}

	payload, errPayload := json.Marshal(errorPayload)

	if errPayload != nil {
		return nil, errPayload
	}

	envelope := Envelope{
		Version:     SupportedVersion,
		MessageType: TypeError,
		Payload:     payload,
	}

	errPackage, err := json.Marshal(envelope)

	if err != nil {
		return nil, err
	}

	return errPackage, nil
}

func DecodeInsert(payloadInsert UpdatePayload) (InsertOp, error) {

	var insertBytes InsertOp

	err := json.Unmarshal(payloadInsert.Operation, &insertBytes)

	if err != nil {
		return InsertOp{}, err
	}

	if insertBytes.NewID.ClientID == "" {
		return InsertOp{}, errors.New("missing_field")
	}

	if insertBytes.OriginID.ClientID == "" {
		return InsertOp{}, errors.New("missing_field")
	}

	if insertBytes.RightID.ClientID == "" {
		return InsertOp{}, errors.New("missing_field")
	}

	characters := []rune(insertBytes.Character)

	if len(characters) == 0 {
		return InsertOp{}, errors.New("missing_field")
	}

	if len(characters) != 1 {
		return InsertOp{}, errors.New("invalid_payload")
	}

	return insertBytes, nil
}

func DecodeDelete(payloadDelete UpdatePayload) (DeleteOp, error) {

	var deleteBytes DeleteOp

	err := json.Unmarshal(payloadDelete.Operation, &deleteBytes)

	if err != nil {
		return DeleteOp{}, err
	}

	if deleteBytes.TargetID.ClientID == "" {
		return DeleteOp{}, errors.New("missing_field")
	}

	return deleteBytes, nil
}

func DecodeSync1(payloadSync1 UpdatePayload) (SyncOp1, error) {

	var sync1Bytes SyncOp1

	err := json.Unmarshal(payloadSync1.Operation, &sync1Bytes)

	if err != nil {
		return SyncOp1{}, err
	}

	if sync1Bytes.VectorState == nil {
		return SyncOp1{}, errors.New("missing_field")
	}

	if sync1Bytes.DeleteSet == nil {
		return SyncOp1{}, errors.New("missing_field")
	}

	return sync1Bytes, nil
}

func DecodeSync2(payloadSync2 UpdatePayload) (SyncOp2, error) {

	var sync2Bytes SyncOp2

	err := json.Unmarshal(payloadSync2.Operation, &sync2Bytes)

	if err != nil {
		return SyncOp2{}, err
	}

	if sync2Bytes.Delta == nil {
		return SyncOp2{}, errors.New("missing_field")
	}

	return sync2Bytes, nil
}

func DecodeSnapshot(payloadSnapshot UpdatePayload) (SnapshotOp, error) {

	var snapshotBytes SnapshotOp

	err := json.Unmarshal(payloadSnapshot.Operation, &snapshotBytes)

	if err != nil {
		return SnapshotOp{}, err
	}

	if snapshotBytes.Delta == nil {
		return SnapshotOp{}, errors.New("missing_field")
	}

	return snapshotBytes, nil
}

func DecodePresence(payloadPresence UpdatePayload) (PresenceOp, error) {

	var presenceBytes PresenceOp

	err := json.Unmarshal(payloadPresence.Operation, &presenceBytes)
	if err != nil {
		return PresenceOp{}, err
	}

	if presenceBytes.Users == nil {
		return PresenceOp{}, errors.New("missing_field")
	}

	return presenceBytes, nil
}
