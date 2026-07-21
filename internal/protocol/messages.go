package protocol

import (
	"encoding/json"
	"errors"
	"strings"
)

var SupportedVersion = 1
var TypeJoin = "join"
var TypeJoinAck = "join_ack"

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
