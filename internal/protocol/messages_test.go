package protocol

import (
	"encoding/json"
	"testing"

	"github.com/john6604/yata-collaborative-editor/internal/identifier"
)

var (
	testStartID = identifier.ID{ClientID: "START", Clock: -1}
	testEndID   = identifier.ID{ClientID: "END", Clock: -2}
)

func testInsertUpdatePayload(t *testing.T, character any) UpdatePayload {
	t.Helper()

	operation := map[string]any{
		"type":      OpInsert,
		"new_id":    identifier.ID{ClientID: "client-A", Clock: 0},
		"origin_id": testStartID,
		"right_id":  testEndID,
		"character": character,
	}

	operationBytes, err := json.Marshal(operation)
	if err != nil {
		t.Fatalf("json.Marshal(operation) returned an unexpected error: %v", err)
	}

	return UpdatePayload{Operation: operationBytes}
}

func TestDecodeInsertAcceptsSingleRuneCharacterStrings(t *testing.T) {
	for _, character := range []string{"H", "ñ", "😀"} {
		t.Run(character, func(t *testing.T) {
			if err := DecodeInsert(testInsertUpdatePayload(t, character)); err != nil {
				t.Fatalf("DecodeInsert() returned an unexpected error: %v", err)
			}
		})
	}
}

func TestDecodeInsertRejectsInvalidCharacterPayloads(t *testing.T) {
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
			if err := DecodeInsert(testInsertUpdatePayload(t, tt.character)); err == nil {
				t.Fatal("DecodeInsert() returned nil; expected an error")
			}
		})
	}
}
