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

const testWebSocketOrigin = "http://localhost:5173"

func startWebSocketTestServer(t *testing.T) string {
	t.Helper()

	relayServer, err := NewRelayServer(t.TempDir()+"/relay.db", ":0")
	if err != nil {
		t.Fatalf("NewRelayServer() returned an unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := relayServer.Storage.CloseDB(); err != nil {
			t.Errorf("CloseDB() returned an unexpected error: %v", err)
		}
	})

	server := httptest.NewServer(http.HandlerFunc(relayServer.ws))
	t.Cleanup(server.Close)

	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func dialJoinedClient(t *testing.T, url string, roomID string, clientID string) *websocket.Conn {
	return dialJoinedClientWithName(t, url, roomID, clientID, clientID)
}

func dialJoinedClientWithName(t *testing.T, url string, roomID string, clientID string, name string) *websocket.Conn {
	t.Helper()

	header := http.Header{}
	header.Set("Origin", testWebSocketOrigin)

	conn, _, err := websocket.DefaultDialer.Dial(url, header)
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
		"name":      name,
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

	assertNextPresenceContains(t, conn, clientID, name)

	return conn
}

func assertNextPresenceContains(t *testing.T, conn *websocket.Conn, clientID string, name string) {
	t.Helper()

	presence := readNextPresence(t, conn)

	for _, user := range presence.Users {
		if user.Id == clientID && user.Name == name {
			return
		}
	}

	t.Fatalf("presence users did not contain client %q with name %q", clientID, name)
}

func readNextPresence(t *testing.T, conn *websocket.Conn) protocol.PresenceOp {
	t.Helper()

	messageType, received, err := readWebSocketMessage(t, conn)
	if err != nil {
		t.Fatalf("ReadMessage(presence) returned an unexpected error: %v", err)
	}
	if messageType != websocket.TextMessage {
		t.Fatalf("presence message type = %d; expected %d", messageType, websocket.TextMessage)
	}

	var envelope protocol.Envelope
	if err := json.Unmarshal(received, &envelope); err != nil {
		t.Fatalf("json.Unmarshal(presence envelope) returned an unexpected error: %v", err)
	}
	if envelope.MessageType != protocol.TypeUpdate {
		t.Fatalf("presence envelope type = %q; expected %q", envelope.MessageType, protocol.TypeUpdate)
	}

	var payload protocol.UpdatePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("json.Unmarshal(presence update payload) returned an unexpected error: %v", err)
	}

	var presence protocol.PresenceOp
	if err := json.Unmarshal(payload.Operation, &presence); err != nil {
		t.Fatalf("json.Unmarshal(presence operation) returned an unexpected error: %v", err)
	}
	if presence.Type != protocol.OpPresence {
		t.Fatalf("presence operation type = %q; expected %q", presence.Type, protocol.OpPresence)
	}

	return presence
}

func assertPresenceUsers(t *testing.T, presence protocol.PresenceOp, expectedUsers map[string]string) {
	t.Helper()

	gotUsers := make(map[string]string)
	for _, user := range presence.Users {
		gotUsers[user.Id] = user.Name
	}

	if len(gotUsers) != len(expectedUsers) {
		t.Fatalf("%d presence users were expected, but got %d", len(expectedUsers), len(gotUsers))
	}

	for clientID, name := range expectedUsers {
		gotName, exists := gotUsers[clientID]
		if !exists {
			t.Fatalf("presence users did not contain client %q", clientID)
		}
		if gotName != name {
			t.Fatalf("presence user %q name = %q; expected %q", clientID, gotName, name)
		}
	}
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

func TestWebSocketBroadcastsPresenceWithActiveUsers(t *testing.T) {
	url := startWebSocketTestServer(t)
	alice := dialJoinedClientWithName(t, url, "room-1", "client-A", "Alice")
	dialJoinedClientWithName(t, url, "room-1", "client-B", "Bob")

	presence := readNextPresence(t, alice)
	assertPresenceUsers(t, presence, map[string]string{
		"client-A": "Alice",
		"client-B": "Bob",
	})
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
