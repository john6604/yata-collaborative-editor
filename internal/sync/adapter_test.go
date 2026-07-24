package sync

import (
	"encoding/json"
	"testing"

	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

var (
	adapterStartID = identifier.ID{ClientID: "START", Clock: -1}
	adapterEndID   = identifier.ID{ClientID: "END", Clock: -2}
)

func mustDecodeUpdatePayload(t *testing.T, payload string) protocol.UpdatePayload {
	t.Helper()

	update, err := protocol.DecodeUpdate([]byte(payload))
	if err != nil {
		t.Fatalf("DecodeUpdate() returned an unexpected error: %v", err)
	}

	return update
}

func updatePayloadFromOperation(t *testing.T, operation string) protocol.UpdatePayload {
	t.Helper()

	if !json.Valid([]byte(operation)) {
		t.Fatalf("operation JSON is invalid: %s", operation)
	}

	return protocol.UpdatePayload{Operation: json.RawMessage(operation)}
}

func assertEmptyConvertedOperation(t *testing.T, operation ConvertedOperation) {
	t.Helper()

	if operation.Type != "" {
		t.Errorf("ConvertedOperation.Type = %q; expected an empty value", operation.Type)
	}
	if operation.Insert != (protocol.InsertOperation{}) {
		t.Errorf("ConvertedOperation.Insert = %#v; expected an empty value", operation.Insert)
	}
	if operation.Delete != (protocol.DeleteOperation{}) {
		t.Errorf("ConvertedOperation.Delete = %#v; expected an empty value", operation.Delete)
	}
}

func convertedInsert(newID identifier.ID, originID identifier.ID, rightID identifier.ID, content rune) ConvertedOperation {
	return ConvertedOperation{
		Type: protocol.OpInsert,
		Insert: protocol.InsertOperation{
			NewID:    newID,
			OriginID: originID,
			RightID:  rightID,
			Content:  content,
		},
	}
}

func convertedDelete(targetID identifier.ID) ConvertedOperation {
	return ConvertedOperation{
		Type: protocol.OpDelete,
		Delete: protocol.DeleteOperation{
			TargetID: targetID,
		},
	}
}

func mustEncodeAndConvertRoundTrip(t *testing.T, operation ConvertedOperation) ConvertedOperation {
	t.Helper()

	message, err := EncodeUpdateOperation(operation)
	if err != nil {
		t.Fatalf("EncodeUpdateOperation() returned an unexpected error: %v", err)
	}

	version, messageType, payload, err := protocol.DecodeEnvelope(message)
	if err != nil {
		t.Fatalf("DecodeEnvelope() returned an unexpected error: %v", err)
	}
	if version != protocol.SupportedVersion {
		t.Fatalf("version = %d; expected %d", version, protocol.SupportedVersion)
	}
	if messageType != protocol.TypeUpdate {
		t.Fatalf("message type = %q; expected %q", messageType, protocol.TypeUpdate)
	}

	updatePayload, err := protocol.DecodeUpdate(payload)
	if err != nil {
		t.Fatalf("DecodeUpdate() returned an unexpected error: %v", err)
	}

	converted, err := ConvertUpdateOperation(updatePayload)
	if err != nil {
		t.Fatalf("ConvertUpdateOperation() returned an unexpected error: %v", err)
	}

	return converted
}

func TestConvertUpdateOperationInsertASCII(t *testing.T) {
	payload := mustDecodeUpdatePayload(t, `{"op":{"type":"insert","new_id":{"client_id":"A","clock":1},"origin_id":{"client_id":"root","clock":-2},"right_id":{"client_id":"root","clock":-1},"character":"H"}}`)

	converted, err := ConvertUpdateOperation(payload)
	if err != nil {
		t.Fatalf("ConvertUpdateOperation() returned an unexpected error: %v", err)
	}

	if converted.Type != protocol.OpInsert {
		t.Fatalf("ConvertedOperation.Type = %q; expected %q", converted.Type, protocol.OpInsert)
	}
	if got, want := converted.Insert.NewID, (identifier.ID{ClientID: "A", Clock: 1}); got != want {
		t.Errorf("Insert.NewID = %v; expected %v", got, want)
	}
	if got, want := converted.Insert.OriginID, (identifier.ID{ClientID: "root", Clock: -2}); got != want {
		t.Errorf("Insert.OriginID = %v; expected %v", got, want)
	}
	if got, want := converted.Insert.RightID, (identifier.ID{ClientID: "root", Clock: -1}); got != want {
		t.Errorf("Insert.RightID = %v; expected %v", got, want)
	}
	if got, want := converted.Insert.Content, 'H'; got != want {
		t.Errorf("Insert.Content = %q; expected %q", got, want)
	}
}

func TestConvertUpdateOperationInsertUnicode(t *testing.T) {
	tests := []struct {
		name      string
		character string
		expected  rune
	}{
		{name: "enye", character: `\u00f1`, expected: '\u00f1'},
		{name: "emoji", character: `\ud83d\ude00`, expected: '\U0001F600'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := mustDecodeUpdatePayload(t, `{"op":{"type":"insert","new_id":{"client_id":"A","clock":1},"origin_id":{"client_id":"root","clock":-2},"right_id":{"client_id":"root","clock":-1},"character":"`+tt.character+`"}}`)

			converted, err := ConvertUpdateOperation(payload)
			if err != nil {
				t.Fatalf("ConvertUpdateOperation() returned an unexpected error: %v", err)
			}

			if converted.Type != protocol.OpInsert {
				t.Fatalf("ConvertedOperation.Type = %q; expected %q", converted.Type, protocol.OpInsert)
			}
			if converted.Insert.Content != tt.expected {
				t.Errorf("Insert.Content = %q; expected %q", converted.Insert.Content, tt.expected)
			}
		})
	}
}

func TestConvertUpdateOperationRejectsInsertWithTwoCharacters(t *testing.T) {
	payload := updatePayloadFromOperation(t, `{"type":"insert","new_id":{"client_id":"A","clock":1},"origin_id":{"client_id":"root","clock":-2},"right_id":{"client_id":"root","clock":-1},"character":"ab"}`)

	converted, err := ConvertUpdateOperation(payload)
	if err == nil {
		t.Fatal("ConvertUpdateOperation() returned nil; expected an error")
	}
	if err.Error() != protocol.InvalidPayload {
		t.Fatalf("ConvertUpdateOperation() error = %q; expected %q", err.Error(), protocol.InvalidPayload)
	}
	assertEmptyConvertedOperation(t, converted)
}

func TestConvertUpdateOperationDelete(t *testing.T) {
	payload := mustDecodeUpdatePayload(t, `{"op":{"type":"delete","target_id":{"client_id":"A","clock":1}}}`)

	converted, err := ConvertUpdateOperation(payload)
	if err != nil {
		t.Fatalf("ConvertUpdateOperation() returned an unexpected error: %v", err)
	}

	if converted.Type != protocol.OpDelete {
		t.Fatalf("ConvertedOperation.Type = %q; expected %q", converted.Type, protocol.OpDelete)
	}
	if got, want := converted.Delete.TargetID, (identifier.ID{ClientID: "A", Clock: 1}); got != want {
		t.Errorf("Delete.TargetID = %v; expected %v", got, want)
	}
}

func TestConvertUpdateOperationRejectsDeleteWithoutTargetID(t *testing.T) {
	payload := updatePayloadFromOperation(t, `{"type":"delete"}`)

	converted, err := ConvertUpdateOperation(payload)
	if err == nil {
		t.Fatal("ConvertUpdateOperation() returned nil; expected an error")
	}
	if err.Error() != protocol.MissingField {
		t.Fatalf("ConvertUpdateOperation() error = %q; expected %q", err.Error(), protocol.MissingField)
	}
	assertEmptyConvertedOperation(t, converted)
}

func TestConvertUpdateOperationRejectsUnknownOperation(t *testing.T) {
	payload := updatePayloadFromOperation(t, `{"type":"banana"}`)

	converted, err := ConvertUpdateOperation(payload)
	if err == nil {
		t.Fatal("ConvertUpdateOperation() returned nil; expected an error")
	}
	if err.Error() != protocol.InvalidPayload {
		t.Fatalf("ConvertUpdateOperation() error = %q; expected %q", err.Error(), protocol.InvalidPayload)
	}
	assertEmptyConvertedOperation(t, converted)
}

func TestConvertUpdateOperationRejectsSyncStep1AsUnsupportedConversion(t *testing.T) {
	payload := mustDecodeUpdatePayload(t, `{"op":{"type":"sync_step1","vector_state":{},"delete_set":{}}}`)

	converted, err := ConvertUpdateOperation(payload)
	if err == nil {
		t.Fatal("ConvertUpdateOperation() returned nil; expected an error")
	}
	if err.Error() != "unsupported_conversion" {
		t.Fatalf("ConvertUpdateOperation() error = %q; expected %q", err.Error(), "unsupported_conversion")
	}
	assertEmptyConvertedOperation(t, converted)
}

func TestConvertUpdateOperationRejectsSyncStep2AsUnsupportedConversion(t *testing.T) {
	payload := mustDecodeUpdatePayload(t, `{"op":{"type":"sync_step2","delta":{}}}`)

	converted, err := ConvertUpdateOperation(payload)
	if err == nil {
		t.Fatal("ConvertUpdateOperation() returned nil; expected an error")
	}
	if err.Error() != "unsupported_conversion" {
		t.Fatalf("ConvertUpdateOperation() error = %q; expected %q", err.Error(), "unsupported_conversion")
	}
	assertEmptyConvertedOperation(t, converted)
}

func TestApplyConvertedOperationRemoteInsert(t *testing.T) {
	doc := document.NewDocument()
	idH := identifier.ID{ClientID: "A", Clock: 1}

	err := ApplyConvertedOperation(doc, convertedInsert(idH, adapterStartID, adapterEndID, 'H'))
	if err != nil {
		t.Fatalf("ApplyConvertedOperation(insert) returned an unexpected error: %v", err)
	}

	if got, want := doc.VisibleContent(), "H"; got != want {
		t.Fatalf("VisibleContent() = %q; expected %q", got, want)
	}
}

func TestApplyConvertedOperationTwoRemoteInsertsInOrder(t *testing.T) {
	doc := document.NewDocument()
	idH := identifier.ID{ClientID: "A", Clock: 1}
	idI := identifier.ID{ClientID: "A", Clock: 2}

	if err := ApplyConvertedOperation(doc, convertedInsert(idH, adapterStartID, adapterEndID, 'H')); err != nil {
		t.Fatalf("ApplyConvertedOperation(insert H) returned an unexpected error: %v", err)
	}
	if err := ApplyConvertedOperation(doc, convertedInsert(idI, idH, adapterEndID, 'i')); err != nil {
		t.Fatalf("ApplyConvertedOperation(insert i) returned an unexpected error: %v", err)
	}

	if got, want := doc.VisibleContent(), "Hi"; got != want {
		t.Fatalf("VisibleContent() = %q; expected %q", got, want)
	}
}

func TestApplyConvertedOperationRemoteDelete(t *testing.T) {
	doc := document.NewDocument()
	idH := identifier.ID{ClientID: "A", Clock: 1}

	if err := ApplyConvertedOperation(doc, convertedInsert(idH, adapterStartID, adapterEndID, 'H')); err != nil {
		t.Fatalf("ApplyConvertedOperation(insert) returned an unexpected error: %v", err)
	}
	if err := ApplyConvertedOperation(doc, convertedDelete(idH)); err != nil {
		t.Fatalf("ApplyConvertedOperation(delete) returned an unexpected error: %v", err)
	}

	if got, want := doc.VisibleContent(), ""; got != want {
		t.Fatalf("VisibleContent() = %q; expected %q", got, want)
	}
	if element := doc.ElementsByID[idH]; element == nil {
		t.Fatalf("element %v was not found after delete", idH)
	} else if !element.IsDeleted {
		t.Fatalf("element %v exists but was not marked as deleted", idH)
	}
}

func TestApplyConvertedOperationRejectsUnknownType(t *testing.T) {
	doc := document.NewDocument()

	err := ApplyConvertedOperation(doc, ConvertedOperation{Type: "banana"})
	if err == nil {
		t.Fatal("ApplyConvertedOperation() returned nil; expected an error")
	}
	if err.Error() != "unsupported_operation" {
		t.Fatalf("ApplyConvertedOperation() error = %q; expected %q", err.Error(), "unsupported_operation")
	}
	if got, want := doc.VisibleContent(), ""; got != want {
		t.Fatalf("VisibleContent() = %q; expected %q", got, want)
	}
}

func TestApplyConvertedOperationKeepsMissingDependencyInsertPending(t *testing.T) {
	doc := document.NewDocument()
	idH := identifier.ID{ClientID: "A", Clock: 1}
	idI := identifier.ID{ClientID: "A", Clock: 2}

	err := ApplyConvertedOperation(doc, convertedInsert(idI, idH, adapterEndID, 'i'))
	if err != nil {
		t.Fatalf("ApplyConvertedOperation(pending insert) returned an unexpected error: %v", err)
	}

	if got, want := doc.VisibleContent(), ""; got != want {
		t.Fatalf("VisibleContent() = %q; expected %q", got, want)
	}
	if got, want := len(doc.PendingInserts), 1; got != want {
		t.Fatalf("len(PendingInserts) = %d; expected %d", got, want)
	}
}

func TestApplyConvertedOperationKeepsMissingDeletePending(t *testing.T) {
	doc := document.NewDocument()
	missingID := identifier.ID{ClientID: "A", Clock: 99}

	err := ApplyConvertedOperation(doc, convertedDelete(missingID))
	if err != nil {
		t.Fatalf("ApplyConvertedOperation(delete missing ID) returned an unexpected error: %v", err)
	}

	if got, want := doc.VisibleContent(), ""; got != want {
		t.Fatalf("VisibleContent() = %q; expected %q", got, want)
	}
	if got, want := len(doc.PendingDeletes), 1; got != want {
		t.Fatalf("len(PendingDeletes) = %d; expected %d", got, want)
	}
}

func TestApplyConvertedOperationResolvesPendingInsert(t *testing.T) {
	doc := document.NewDocument()
	idH := identifier.ID{ClientID: "A", Clock: 1}
	idI := identifier.ID{ClientID: "A", Clock: 2}

	if err := ApplyConvertedOperation(doc, convertedInsert(idI, idH, adapterEndID, 'i')); err != nil {
		t.Fatalf("ApplyConvertedOperation(pending insert) returned an unexpected error: %v", err)
	}
	if err := ApplyConvertedOperation(doc, convertedInsert(idH, adapterStartID, adapterEndID, 'H')); err != nil {
		t.Fatalf("ApplyConvertedOperation(unblocking insert) returned an unexpected error: %v", err)
	}

	if got, want := doc.VisibleContent(), "Hi"; got != want {
		t.Fatalf("VisibleContent() = %q; expected %q", got, want)
	}
	if got, want := len(doc.PendingInserts), 0; got != want {
		t.Fatalf("len(PendingInserts) = %d; expected %d", got, want)
	}
}

func TestEncodeUpdateOperationRoundTripInsert(t *testing.T) {
	idH := identifier.ID{ClientID: "B", Clock: 1}
	original := convertedInsert(idH, adapterStartID, adapterEndID, 'H')

	converted := mustEncodeAndConvertRoundTrip(t, original)

	if converted.Type != protocol.OpInsert {
		t.Fatalf("ConvertedOperation.Type = %q; expected %q", converted.Type, protocol.OpInsert)
	}
	if got, want := converted.Insert.NewID, idH; got != want {
		t.Errorf("Insert.NewID = %v; expected %v", got, want)
	}
	if got, want := converted.Insert.OriginID, adapterStartID; got != want {
		t.Errorf("Insert.OriginID = %v; expected %v", got, want)
	}
	if got, want := converted.Insert.RightID, adapterEndID; got != want {
		t.Errorf("Insert.RightID = %v; expected %v", got, want)
	}
	if got, want := converted.Insert.Content, 'H'; got != want {
		t.Errorf("Insert.Content = %q; expected %q", got, want)
	}
}

func TestEncodeUpdateOperationRoundTripUnicodeInsert(t *testing.T) {
	tests := []struct {
		name    string
		content rune
	}{
		{name: "enye", content: '\u00f1'},
		{name: "emoji", content: '\U0001F600'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := identifier.ID{ClientID: "B", Clock: 1}
			original := convertedInsert(id, adapterStartID, adapterEndID, tt.content)

			converted := mustEncodeAndConvertRoundTrip(t, original)

			if converted.Type != protocol.OpInsert {
				t.Fatalf("ConvertedOperation.Type = %q; expected %q", converted.Type, protocol.OpInsert)
			}
			if converted.Insert.Content != tt.content {
				t.Errorf("Insert.Content = %q; expected %q", converted.Insert.Content, tt.content)
			}
		})
	}
}

func TestEncodeUpdateOperationRoundTripDelete(t *testing.T) {
	targetID := identifier.ID{ClientID: "B", Clock: 1}
	original := convertedDelete(targetID)

	converted := mustEncodeAndConvertRoundTrip(t, original)

	if converted.Type != protocol.OpDelete {
		t.Fatalf("ConvertedOperation.Type = %q; expected %q", converted.Type, protocol.OpDelete)
	}
	if got, want := converted.Delete.TargetID, targetID; got != want {
		t.Errorf("Delete.TargetID = %v; expected %v", got, want)
	}
}

func TestEncodeUpdateOperationRejectsUnsupportedType(t *testing.T) {
	message, err := EncodeUpdateOperation(ConvertedOperation{Type: "banana"})

	if err == nil {
		t.Fatal("EncodeUpdateOperation() returned nil; expected an error")
	}
	if err.Error() != "unsupported_operation" {
		t.Fatalf("EncodeUpdateOperation() error = %q; expected %q", err.Error(), "unsupported_operation")
	}
	if message != nil {
		t.Fatalf("EncodeUpdateOperation() bytes = %s; expected nil", message)
	}
}
