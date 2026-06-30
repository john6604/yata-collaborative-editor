package document_test

import (
	"testing"

	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/storage"
	crdtsync "github.com/john6604/yata-collaborative-editor/internal/sync"
)

var integrationEndID = identifier.ID{ClientID: "END", Clock: -2}

func insertIntegrationByte(t *testing.T, doc *document.Document, character byte) identifier.ID {
	t.Helper()

	err, id := doc.InsertElement(doc.VisibleLength(), character)
	if err != nil {
		t.Fatalf("InsertElement(%d, %q) returned an unexpected error: %v", doc.VisibleLength(), character, err)
	}
	return id
}

func integrateRemoteInsert(t *testing.T, doc *document.Document, id, originID, rightID identifier.ID, content byte) {
	t.Helper()

	if err := doc.RemoteInsert(id, originID, rightID, content); err != nil {
		t.Fatalf("RemoteInsert(%v) returned an unexpected error: %v", id, err)
	}
}

func vectorFor(doc *document.Document) crdtsync.Vector {
	vector := crdtsync.Vector{}
	vector.GenerateStateVector(*doc)
	vector.GenerateDeleteSet(*doc)
	return vector
}

func syncFromTo(t *testing.T, source, target *document.Document) {
	t.Helper()

	targetVector := vectorFor(target)
	missingInserts, missingDeletes := crdtsync.ComputeDelta(*source, targetVector)
	delta := crdtsync.ComputeSerializedDelta(*source, missingInserts, missingDeletes)

	if err := target.IntegrateDelta(delta); err != nil {
		t.Fatalf("IntegrateDelta() returned an unexpected error: %v", err)
	}
}

func assertDocumentsConverged(t *testing.T, left, right *document.Document) {
	t.Helper()

	if got, want := right.VisibleContent(), left.VisibleContent(); got != want {
		t.Errorf("visible state did not converge: got %q; expected %q", got, want)
	}
	if got, want := right.VisibleLength(), left.VisibleLength(); got != want {
		t.Errorf("visible length did not converge: got %d; expected %d", got, want)
	}
	if got, want := right.PrintInternal(), left.PrintInternal(); got != want {
		t.Errorf("internal state did not converge:\n got: %s\nwant: %s", got, want)
	}
	if got := len(right.PendingInserts); got != 0 {
		t.Errorf("target has %d pending inserts after sync; expected 0", got)
	}
	if got := len(right.PendingDeletes); got != 0 {
		t.Errorf("target has %d pending deletes after sync; expected 0", got)
	}
}

func buildSourceAndLaggingReplica(t *testing.T) (*document.Document, *document.Document) {
	t.Helper()

	source := document.NewDocument()
	source.ClientID = "client-A"
	source.Clock = 0

	idA := insertIntegrationByte(t, source, 'A')
	insertIntegrationByte(t, source, 'B')
	insertIntegrationByte(t, source, 'C')

	if err := source.Delete(1); err != nil {
		t.Fatalf("Delete(1) returned an unexpected error: %v", err)
	}

	target := document.NewDocument()
	integrateRemoteInsert(t, target, idA, identifier.ID{ClientID: "START", Clock: -1}, integrationEndID, 'A')

	return source, target
}

func TestSyncEndToEndWithoutPersistence(t *testing.T) {
	source, target := buildSourceAndLaggingReplica(t)

	syncFromTo(t, source, target)

	assertDocumentsConverged(t, source, target)
}

func TestSyncEndToEndAfterSnapshotReconstruction(t *testing.T) {
	source, target := buildSourceAndLaggingReplica(t)

	sourceStorage := &storage.Storage{}
	if err := sourceStorage.OpenDB(t.TempDir() + "/source.db"); err != nil {
		t.Fatalf("source OpenDB() returned an unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := sourceStorage.CloseDB(); err != nil {
			t.Errorf("source CloseDB() returned an unexpected error: %v", err)
		}
	})

	targetStorage := &storage.Storage{}
	if err := targetStorage.OpenDB(t.TempDir() + "/target.db"); err != nil {
		t.Fatalf("target OpenDB() returned an unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := targetStorage.CloseDB(); err != nil {
			t.Errorf("target CloseDB() returned an unexpected error: %v", err)
		}
	})

	if err := sourceStorage.SaveSnapshot(source); err != nil {
		t.Fatalf("source SaveSnapshot() returned an unexpected error: %v", err)
	}
	if err := targetStorage.SaveSnapshot(target); err != nil {
		t.Fatalf("target SaveSnapshot() returned an unexpected error: %v", err)
	}

	reconstructedSource := document.NewDocument()
	if err := reconstructedSource.ReconstructDocument(sourceStorage); err != nil {
		t.Fatalf("source ReconstructDocument() returned an unexpected error: %v", err)
	}
	reconstructedTarget := document.NewDocument()
	if err := reconstructedTarget.ReconstructDocument(targetStorage); err != nil {
		t.Fatalf("target ReconstructDocument() returned an unexpected error: %v", err)
	}

	syncFromTo(t, reconstructedSource, reconstructedTarget)
	assertDocumentsConverged(t, reconstructedSource, reconstructedTarget)

	// Persist the synchronized target once more and verify convergence survives
	// another process restart.
	if err := targetStorage.SaveSnapshot(reconstructedTarget); err != nil {
		t.Fatalf("post-sync SaveSnapshot() returned an unexpected error: %v", err)
	}

	restartedTarget := document.NewDocument()
	if err := restartedTarget.ReconstructDocument(targetStorage); err != nil {
		t.Fatalf("post-sync ReconstructDocument() returned an unexpected error: %v", err)
	}
	assertDocumentsConverged(t, reconstructedSource, restartedTarget)
}
