package storage

import (
	"reflect"
	"strings"
	"testing"

	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

var (
	testStartID = identifier.ID{ClientID: "START", Clock: -1}
	testEndID   = identifier.ID{ClientID: "END", Clock: -2}
)

func openTestStorage(t *testing.T) *Storage {
	t.Helper()

	storage := &Storage{}
	if err := storage.OpenDB(t.TempDir() + "/snapshot.db"); err != nil {
		t.Fatalf("OpenDB() returned an unexpected error: %v", err)
	}

	t.Cleanup(func() {
		if err := storage.CloseDB(); err != nil {
			t.Errorf("CloseDB() returned an unexpected error: %v", err)
		}
	})

	return storage
}

func insertByte(t *testing.T, doc *document.Document, character byte) identifier.ID {
	t.Helper()

	err, id := doc.InsertElement(doc.VisibleLength(), character)
	if err != nil {
		t.Fatalf("InsertElement(%d, %q) returned an unexpected error: %v", doc.VisibleLength(), character, err)
	}

	return id
}

func assertInsertLogsEqual(t *testing.T, got, want map[identifier.ID]*protocol.InsertOperation) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("InsertLog mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func assertDeleteLogsEqual(t *testing.T, got, want map[identifier.ID]*protocol.DeleteOperation) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("DeleteLog mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestSnapshotRoundTripBasic(t *testing.T) {
	storage := openTestStorage(t)

	original := document.NewDocument()
	original.ClientID = "snapshot-client"
	original.Clock = 0

	idA := insertByte(t, original, 'A')
	idB := insertByte(t, original, 'B')
	idC := insertByte(t, original, 'C')

	if err := original.Delete(1); err != nil {
		t.Fatalf("Delete(1) returned an unexpected error: %v", err)
	}

	if err := storage.SaveSnapshot(original); err != nil {
		t.Fatalf("SaveSnapshot() returned an unexpected error: %v", err)
	}

	reconstructed := document.NewDocument()
	if err := reconstructed.ReconstructDocument(storage); err != nil {
		t.Fatalf("ReconstructDocument() returned an unexpected error: %v", err)
	}

	if got, want := reconstructed.VisibleContent(), original.VisibleContent(); got != want {
		t.Errorf("VisibleContent() = %q; expected %q", got, want)
	}
	if got, want := reconstructed.VisibleLength(), original.VisibleLength(); got != want {
		t.Errorf("VisibleLength() = %d; expected %d", got, want)
	}
	if got, want := reconstructed.PrintInternal(), original.PrintInternal(); got != want {
		t.Errorf("PrintInternal() = %q; expected %q", got, want)
	}
	if got, want := reconstructed.CharacterCounter, original.CharacterCounter; got != want {
		t.Errorf("CharacterCounter = %d; expected %d", got, want)
	}
	if got, want := reconstructed.Clock, original.Clock; got != want {
		t.Errorf("Clock = %d; expected %d", got, want)
	}
	if got, want := reconstructed.ClientID, original.ClientID; got != want {
		t.Errorf("ClientID = %q; expected %q", got, want)
	}

	assertInsertLogsEqual(t, reconstructed.InsertLog, original.InsertLog)
	assertDeleteLogsEqual(t, reconstructed.DeleteLog, original.DeleteLog)

	if got := reconstructed.ElementsByID[idB]; got == nil {
		t.Fatalf("deleted element %v was not reconstructed", idB)
	} else if !got.IsDeleted {
		t.Errorf("element %v was reconstructed without its tombstone", idB)
	}

	if !strings.Contains(reconstructed.PrintInternal(), "B(X)") {
		t.Errorf("PrintInternal() = %q; expected the B(X) tombstone", reconstructed.PrintInternal())
	}

	for _, id := range []identifier.ID{idA, idB, idC} {
		if reconstructed.ElementsByID[id] == nil {
			t.Errorf("element %v is missing after reconstruction", id)
		}
	}

	if got := len(reconstructed.PendingInserts); got != 0 {
		t.Errorf("PendingInserts has %d entries; expected 0", got)
	}
	if got := len(reconstructed.PendingDeletes); got != 0 {
		t.Errorf("PendingDeletes has %d entries; expected 0", got)
	}
}

func TestSnapshotRoundTripKeepsUnresolvablePendings(t *testing.T) {
	storage := openTestStorage(t)

	original := document.NewDocument()
	original.ClientID = "snapshot-owner"

	missingOrigin := identifier.ID{ClientID: "remote", Clock: 8}
	pendingInsertID := identifier.ID{ClientID: "remote", Clock: 9}
	pendingDeleteID := identifier.ID{ClientID: "remote", Clock: 12}

	if err := original.RemoteInsert(pendingInsertID, missingOrigin, testEndID, 'P'); err == nil || err.Error() != "Pending value." {
		t.Fatalf("RemoteInsert() error = %v; expected Pending value.", err)
	}
	if err := original.RemoteDelete(pendingDeleteID); err != nil {
		t.Fatalf("RemoteDelete() returned an unexpected error: %v", err)
	}

	if err := storage.SaveSnapshot(original); err != nil {
		t.Fatalf("SaveSnapshot() returned an unexpected error: %v", err)
	}

	reconstructed := document.NewDocument()
	if err := reconstructed.ReconstructDocument(storage); err != nil {
		t.Fatalf("ReconstructDocument() returned an unexpected error: %v", err)
	}

	if got := reconstructed.VisibleContent(); got != "" {
		t.Errorf("VisibleContent() = %q; expected an empty document", got)
	}
	if got := reconstructed.VisibleLength(); got != 0 {
		t.Errorf("VisibleLength() = %d; expected 0", got)
	}

	if _, exists := reconstructed.PendingInserts[pendingInsertID]; !exists {
		t.Errorf("pending insert %v was not reconstructed", pendingInsertID)
	}
	if _, exists := reconstructed.PendingDeletes[pendingDeleteID]; !exists {
		t.Errorf("pending delete %v was not reconstructed", pendingDeleteID)
	}
	if _, exists := reconstructed.ElementsByID[pendingInsertID]; exists {
		t.Errorf("unresolvable insert %v was unexpectedly integrated", pendingInsertID)
	}
	if _, exists := reconstructed.ElementsByID[pendingDeleteID]; exists {
		t.Errorf("missing delete target %v was unexpectedly integrated", pendingDeleteID)
	}

	if _, exists := reconstructed.InsertLog[pendingInsertID]; !exists {
		t.Errorf("InsertLog does not contain pending insert %v", pendingInsertID)
	}
	if _, exists := reconstructed.DeleteLog[pendingDeleteID]; !exists {
		t.Errorf("DeleteLog does not contain pending delete %v", pendingDeleteID)
	}
}

func TestSnapshotRoundTripResolvesCausalPendings(t *testing.T) {
	storage := openTestStorage(t)

	original := document.NewDocument()
	original.ClientID = "snapshot-owner"
	original.Clock = 3

	idA := identifier.ID{ClientID: "remote", Clock: 0}
	idB := identifier.ID{ClientID: "remote", Clock: 1}
	idC := identifier.ID{ClientID: "remote", Clock: 2}

	// The snapshot deliberately contains the operations in the logs but not in
	// the linked list. ReconstructDocument must rebuild the pending buffers and
	// resolve the complete causal chain after hydration.
	original.InsertLog[idA] = protocol.NewInsertOperation(idA, testStartID, testEndID, 'A')
	original.InsertLog[idB] = protocol.NewInsertOperation(idB, idA, testEndID, 'B')
	original.InsertLog[idC] = protocol.NewInsertOperation(idC, idB, testEndID, 'C')
	original.DeleteLog[idC] = protocol.NewDeleteOperation(idC)

	if err := storage.SaveSnapshot(original); err != nil {
		t.Fatalf("SaveSnapshot() returned an unexpected error: %v", err)
	}

	reconstructed := document.NewDocument()
	if err := reconstructed.ReconstructDocument(storage); err != nil {
		t.Fatalf("ReconstructDocument() returned an unexpected error: %v", err)
	}

	if got, want := reconstructed.VisibleContent(), "AB"; got != want {
		t.Errorf("VisibleContent() = %q; expected %q", got, want)
	}
	if got, want := reconstructed.VisibleLength(), 2; got != want {
		t.Errorf("VisibleLength() = %d; expected %d", got, want)
	}
	if got, want := reconstructed.PrintInternal(), "START -> A -> B -> C(X) -> END"; got != want {
		t.Errorf("PrintInternal() = %q; expected %q", got, want)
	}
	if got := len(reconstructed.PendingInserts); got != 0 {
		t.Errorf("PendingInserts has %d entries; expected 0", got)
	}
	if got := len(reconstructed.PendingDeletes); got != 0 {
		t.Errorf("PendingDeletes has %d entries; expected 0", got)
	}

	if element := reconstructed.ElementsByID[idC]; element == nil {
		t.Fatalf("element %v was not integrated", idC)
	} else if !element.IsDeleted {
		t.Errorf("element %v was integrated but its pending delete was not applied", idC)
	}
}
