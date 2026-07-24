package relay

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

var (
	websocketStartID = identifier.ID{ClientID: "START", Clock: -1}
	websocketEndID   = identifier.ID{ClientID: "END", Clock: -2}
)

func startWebSocketTestServer(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(NewRelayServer("").ws))
	t.Cleanup(server.Close)

	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func dialJoinedClient(t *testing.T, url string, roomID string, clientID string) *websocket.Conn {
	t.Helper()

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Dial() returned an unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Errorf("Close() returned an unexpected error: %v", err)
		}
	})

	message := envelopeBytes(t, protocol.TypeJoin, map[string]string{
		"room":      roomID,
		"client_id": clientID,
	})

	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		t.Fatalf("WriteMessage(join) returned an unexpected error: %v", err)
	}

	messageType, _, err := readWebSocketMessage(t, conn)
	if err != nil {
		t.Fatalf("ReadMessage(join ack) returned an unexpected error: %v", err)
	}
	if messageType != websocket.TextMessage {
		t.Fatalf("join ack message type = %d; expected %d", messageType, websocket.TextMessage)
	}

	return conn
}

func readWebSocketMessage(t *testing.T, conn *websocket.Conn) (int, []byte, error) {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("SetReadDeadline() returned an unexpected error: %v", err)
	}

	return conn.ReadMessage()
}

func envelopeBytes(t *testing.T, messageType string, payload any) []byte {
	t.Helper()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal(payload) returned an unexpected error: %v", err)
	}

	message, err := json.Marshal(protocol.Envelope{
		Version:     protocol.SupportedVersion,
		MessageType: messageType,
		Payload:     payloadBytes,
	})
	if err != nil {
		t.Fatalf("json.Marshal(envelope) returned an unexpected error: %v", err)
	}

	return message
}

func insertUpdateBytes(t *testing.T, clock int, character any) []byte {
	t.Helper()

	operation := map[string]any{
		"type":      protocol.OpInsert,
		"new_id":    identifier.ID{ClientID: "sender", Clock: clock},
		"origin_id": websocketStartID,
		"right_id":  websocketEndID,
		"character": character,
	}

	payload := protocol.UpdatePayload{
		Operation: mustMarshalRawMessage(t, operation),
	}

	return envelopeBytes(t, protocol.TypeUpdate, payload)
}

func mustMarshalRawMessage(t *testing.T, value any) json.RawMessage {
	t.Helper()

	bytes, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal(value) returned an unexpected error: %v", err)
	}

	return bytes
}

func TestWebSocketBroadcastsInsertCharactersAsStrings(t *testing.T) {
	url := startWebSocketTestServer(t)
	sender := dialJoinedClient(t, url, "room-1", "sender")
	receiver := dialJoinedClient(t, url, "room-1", "receiver")

	for clock, character := range []string{"H", "ñ", "😀"} {
		message := insertUpdateBytes(t, clock, character)

		if err := sender.WriteMessage(websocket.TextMessage, message); err != nil {
			t.Fatalf("WriteMessage(update %q) returned an unexpected error: %v", character, err)
		}

		messageType, received, err := readWebSocketMessage(t, receiver)
		if err != nil {
			t.Fatalf("ReadMessage(update %q) returned an unexpected error: %v", character, err)
		}
		if messageType != websocket.TextMessage {
			t.Fatalf("update message type = %d; expected %d", messageType, websocket.TextMessage)
		}
		if string(received) != string(message) {
			t.Fatalf("broadcasted message = %s; expected %s", received, message)
		}
	}
}

func TestWebSocketRejectsInvalidInsertCharacters(t *testing.T) {
	tests := []struct {
		name      string
		character any
	}{
		{name: "empty string", character: ""},
		{name: "two runes", character: "ab"},
		{name: "number", character: 72},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := startWebSocketTestServer(t)
			conn := dialJoinedClient(t, url, "room-1", "sender")

			if err := conn.WriteMessage(websocket.TextMessage, insertUpdateBytes(t, 0, tt.character)); err != nil {
				t.Fatalf("WriteMessage(update) returned an unexpected error: %v", err)
			}

			messageType, received, err := readWebSocketMessage(t, conn)
			if err != nil {
				t.Fatalf("ReadMessage(error) returned an unexpected error: %v", err)
			}
			if messageType != websocket.TextMessage {
				t.Fatalf("error message type = %d; expected %d", messageType, websocket.TextMessage)
			}

			var envelope protocol.Envelope
			if err := json.Unmarshal(received, &envelope); err != nil {
				t.Fatalf("json.Unmarshal(error envelope) returned an unexpected error: %v", err)
			}
			if envelope.MessageType != protocol.TypeError {
				t.Fatalf("error envelope type = %q; expected %q", envelope.MessageType, protocol.TypeError)
			}

			var payload protocol.ErrorPayload
			if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
				t.Fatalf("json.Unmarshal(error payload) returned an unexpected error: %v", err)
			}
			if payload.Code != protocol.InvalidPayload {
				t.Fatalf("error code = %q; expected %q", payload.Code, protocol.InvalidPayload)
			}
		})
	}
}
