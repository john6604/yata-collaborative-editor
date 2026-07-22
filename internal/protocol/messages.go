package protocol

import (
	"encoding/json"
	"errors"
	"strings"
)

var SupportedVersion = 1
var TypeJoin = "join"
var TypeJoinAck = "join_ack"
var TypeError = "error"
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

func DecodeJoinMessage(message []byte) (int, string, json.RawMessage, error) {

	var msg Envelope

	err := json.Unmarshal(message, &msg)

	if err != nil {
		return 0, "", nil, err
	}

	return msg.Version, msg.MessageType, msg.Payload, nil
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
