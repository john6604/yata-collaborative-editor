package document

import (
	"strings"
	"testing"
)

// Test: insertText inserts all bytes from the given text at the end of the document.
func insertText(t *testing.T, document *Document, text string) {
	t.Helper()

	for i := 0; i < len(text); i++ {
		character := text[i]
		index := document.VisibleLength()

		if err, _ := document.InsertElement(index, character); err != nil {
			t.Fatalf(
				"InsertElement(%d, %q) returned an unexpected error: %v",
				index,
				character,
				err,
			)
		}
	}
}

// Test: TestNewDocument verifies that a newly created document is empty and only contains the START and END sentinel nodes internally.
func TestNewDocument(t *testing.T) {
	document := NewDocument()

	if got := document.VisibleContent(); got != "" {
		t.Errorf("VisibleContent() = %q; expected %q", got, "")
	}

	if got := document.VisibleLength(); got != 0 {
		t.Errorf("VisibleLength() = %d; expected %d", got, 0)
	}

	if got := document.PrintInternal(); got != "START -> END" {
		t.Errorf(
			"PrintInternal() = %q; expected %q",
			got,
			"START -> END",
		)
	}
}

// Test: TestSequentialInsertion verifies that multiple bytes inserted sequentially produce the expected visible content and length.
func TestSequentialInsertion(t *testing.T) {
	document := NewDocument()

	insertText(t, document, "Hello")

	if got := document.VisibleContent(); got != "Hello" {
		t.Errorf("VisibleContent() = %q; expected %q", got, "Hello")
	}

	if got := document.VisibleLength(); got != 5 {
		t.Errorf("VisibleLength() = %d; expected %d", got, 5)
	}
}

// Test: TestInsertAtBeginning verifies that a byte can be inserted at the beginning of an existing document.
func TestInsertAtBeginning(t *testing.T) {
	document := NewDocument()
	insertText(t, document, "Hello")

	if err, _ := document.InsertElement(0, 'X'); err != nil {
		t.Fatalf("InsertElement(0, 'X') returned an unexpected error: %v", err)
	}

	if got := document.VisibleContent(); got != "XHello" {
		t.Errorf("VisibleContent() = %q; expected %q", got, "XHello")
	}

	if got := document.VisibleLength(); got != 6 {
		t.Errorf("VisibleLength() = %d; expected %d", got, 6)
	}
}

// Test: TestInsertInMiddle verifies that a byte can be inserted between existing visible elements.
func TestInsertInMiddle(t *testing.T) {
	document := NewDocument()
	insertText(t, document, "Hello")

	if err, _ := document.InsertElement(2, 'Y'); err != nil {
		t.Fatalf("InsertElement(2, 'Y') returned an unexpected error: %v", err)
	}

	if got := document.VisibleContent(); got != "HeYllo" {
		t.Errorf("VisibleContent() = %q; expected %q", got, "HeYllo")
	}

	if got := document.VisibleLength(); got != 6 {
		t.Errorf("VisibleLength() = %d; expected %d", got, 6)
	}
}

// Test: TestDeleteVisibleElement verifies that deleting a visible element updates the visible content while preserving the element as a tombstone.
func TestDeleteVisibleElement(t *testing.T) {
	document := NewDocument()
	insertText(t, document, "Hello")

	if err := document.Delete(1); err != nil {
		t.Fatalf("Delete(1) returned an unexpected error: %v", err)
	}

	if got := document.VisibleContent(); got != "Hllo" {
		t.Errorf("VisibleContent() = %q; expected %q", got, "Hllo")
	}

	if got := document.VisibleLength(); got != 4 {
		t.Errorf("VisibleLength() = %d; expected %d", got, 4)
	}

	internal := document.PrintInternal()

	if !strings.Contains(internal, "e(X)") {
		t.Errorf(
			"PrintInternal() = %q; expected it to preserve the e(X) tombstone",
			internal,
		)
	}
}

// Test: TestDeleteLastVisibleElement verifies that deleting the last visible element decreases the visible length while preserving the deleted node internally.
func TestDeleteLastVisibleElement(t *testing.T) {
	document := NewDocument()
	insertText(t, document, "Hello")

	if err := document.Delete(4); err != nil {
		t.Fatalf("Delete(4) returned an unexpected error: %v", err)
	}

	if got := document.VisibleContent(); got != "Hell" {
		t.Errorf("VisibleContent() = %q; expected %q", got, "Hell")
	}

	if got := document.VisibleLength(); got != 4 {
		t.Errorf("VisibleLength() = %d; expected %d", got, 4)
	}

	internal := document.PrintInternal()

	if !strings.Contains(internal, "o(X)") {
		t.Errorf(
			"PrintInternal() = %q; expected it to preserve the o(X) tombstone",
			internal,
		)
	}
}

// Test: TestInsertAfterTombstone verifies that a new element can be inserted after a deleted node while preserving the correct internal order.
func TestInsertAfterTombstone(t *testing.T) {
	document := NewDocument()
	insertText(t, document, "Hello")

	if err := document.Delete(1); err != nil {
		t.Fatalf("Delete(1) returned an unexpected error: %v", err)
	}

	if got := document.VisibleContent(); got != "Hllo" {
		t.Fatalf(
			"After Delete(1), VisibleContent() = %q; expected %q",
			got,
			"Hllo",
		)
	}

	if err, _ := document.InsertElement(1, 'a'); err != nil {
		t.Fatalf("InsertElement(1, 'a') returned an unexpected error: %v", err)
	}

	if got := document.VisibleContent(); got != "Hallo" {
		t.Errorf("VisibleContent() = %q; expected %q", got, "Hallo")
	}

	if got := document.VisibleLength(); got != 5 {
		t.Errorf("VisibleLength() = %d; expected %d", got, 5)
	}

	internal := document.PrintInternal()

	if !strings.Contains(internal, "e(X) -> a") {
		t.Errorf(
			"PrintInternal() = %q; expected it to contain the sequence %q",
			internal,
			"e(X) -> a",
		)
	}
}

// Test: TestInvalidIndexes verifies that invalid insert and delete indexes return errors without modifying the document state.
func TestInvalidIndexes(t *testing.T) {
	tests := []struct {
		name      string
		operation func(document *Document) error
	}{
		{
			name: "delete with a negative index",
			operation: func(document *Document) error {
				return document.Delete(-1)
			},
		},
		{
			name: "delete with an out-of-range index",
			operation: func(document *Document) error {
				return document.Delete(100)
			},
		},
		{
			name: "insert with an out-of-range index",
			operation: func(document *Document) error {
				err, _ := document.InsertElement(100, 'X')
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := NewDocument()
			insertText(t, document, "Hello")

			beforeContent := document.VisibleContent()
			beforeLength := document.VisibleLength()
			beforeInternal := document.PrintInternal()

			err := test.operation(document)
			if err == nil {
				t.Fatal("Expected an error, but the operation returned nil")
			}

			if got := document.VisibleContent(); got != beforeContent {
				t.Errorf(
					"The invalid operation changed the visible content: before=%q, after=%q",
					beforeContent,
					got,
				)
			}

			if got := document.VisibleLength(); got != beforeLength {
				t.Errorf(
					"The invalid operation changed the visible length: before=%d, after=%d",
					beforeLength,
					got,
				)
			}

			if got := document.PrintInternal(); got != beforeInternal {
				t.Errorf(
					"The invalid operation changed the internal state: before=%q, after=%q",
					beforeInternal,
					got,
				)
			}
		})
	}
}

// Concurrency Test: TestConcurrentConvergence verifies that all replicas converge to the same visible content after receiving concurrent insert operations.
func TestConcurrentConvergence(t *testing.T) {
	startID := ID{
		ClientID: "START",
		Clock:    -1,
	}

	endID := ID{
		ClientID: "END",
		Clock:    -2,
	}

	docA := NewDocument()
	docB := NewDocument()
	docC := NewDocument()

	_, idA := docA.InsertElement(0, 'X')
	_, idB := docB.InsertElement(0, 'Y')

	// Propagate A's operation to replicas B and C.
	docB.RemoteInsert(idA, startID, endID, 'X')
	docC.RemoteInsert(idA, startID, endID, 'X')

	// Propagate B's operation to replicas A and C.
	docA.RemoteInsert(idB, startID, endID, 'Y')
	docC.RemoteInsert(idB, startID, endID, 'Y')

	contentA := docA.VisibleContent()
	contentB := docB.VisibleContent()
	contentC := docC.VisibleContent()

	if contentA != contentB {
		t.Errorf(
			"Replica A content = %q; replica B content = %q",
			contentA,
			contentB,
		)
	}

	if contentB != contentC {
		t.Errorf(
			"Replica B content = %q; replica C content = %q",
			contentB,
			contentC,
		)
	}
}

// Concurrency Test: TestRemoteDelete verifies that a delete operation propagated to all replicas removes the element from their visible content.
func TestRemoteDelete(t *testing.T) {
	startID := ID{
		ClientID: "START",
		Clock:    -1,
	}

	endID := ID{
		ClientID: "END",
		Clock:    -2,
	}

	docA := NewDocument()
	docB := NewDocument()
	docC := NewDocument()

	_, idA := docA.InsertElement(0, 'X')

	// Propagate the insertion to replicas B and C.
	docB.RemoteInsert(idA, startID, endID, 'X')
	docC.RemoteInsert(idA, startID, endID, 'X')

	// Delete the element locally and propagate the deletion.
	docA.Delete(0)
	docB.RemoteDelete(idA)
	docC.RemoteDelete(idA)

	if got := docA.VisibleContent(); got != "" {
		t.Errorf("Replica A VisibleContent() = %q; expected an empty string", got)
	}

	if got := docB.VisibleContent(); got != "" {
		t.Errorf("Replica B VisibleContent() = %q; expected an empty string", got)
	}

	if got := docC.VisibleContent(); got != "" {
		t.Errorf("Replica C VisibleContent() = %q; expected an empty string", got)
	}

	if got := docA.VisibleLength(); got != 0 {
		t.Errorf("Replica A VisibleLength() = %d; expected 0", got)
	}

	if got := docB.VisibleLength(); got != 0 {
		t.Errorf("Replica B VisibleLength() = %d; expected 0", got)
	}

	if got := docC.VisibleLength(); got != 0 {
		t.Errorf("Replica C VisibleLength() = %d; expected 0", got)
	}
}

// Concurrency Test: TestRemoteOperationsAreIdempotent verifies that applying the same remote insert and delete operations multiple times does not duplicate nodes or change the final visible state.
func TestRemoteOperationsAreIdempotent(t *testing.T) {
	startID := ID{
		ClientID: "START",
		Clock:    -1,
	}

	endID := ID{
		ClientID: "END",
		Clock:    -2,
	}

	docA := NewDocument()
	docB := NewDocument()

	_, idA := docA.InsertElement(0, 'X')

	// Apply the same remote insertion twice.
	docB.RemoteInsert(idA, startID, endID, 'X')
	docB.RemoteInsert(idA, startID, endID, 'X')

	docA.Delete(0)

	// Apply the same remote deletion twice.
	docB.RemoteDelete(idA)
	docB.RemoteDelete(idA)

	if got := docB.VisibleContent(); got != "" {
		t.Errorf("Replica B VisibleContent() = %q; expected an empty string", got)
	}

	if got := docB.VisibleLength(); got != 0 {
		t.Errorf("Replica B VisibleLength() = %d; expected 0", got)
	}

	nodes := docB.Traverse()

	if got := len(nodes); got != 3 {
		t.Errorf(
			"len(docB.Traverse()) = %d; expected 3 nodes: START, X(X), and END",
			got,
		)
	}
}

// Concurrency Test: TestOriginConsistency verifies that remote insertions preserve the same visible order and internal origin structure as the source replica.
func TestOriginConsistency(t *testing.T) {
	startID := ID{
		ClientID: "START",
		Clock:    -1,
	}

	endID := ID{
		ClientID: "END",
		Clock:    -2,
	}

	docA := NewDocument()
	docB := NewDocument()

	_, idA1 := docA.InsertElement(0, 'A')
	_, idA2 := docA.InsertElement(1, 'X')
	_, idA3 := docA.InsertElement(2, 'P')
	_, idA4 := docA.InsertElement(3, 'Y')

	// Reproduce the original origin chain in replica B.
	docB.RemoteInsert(idA1, startID, endID, 'A')
	docB.RemoteInsert(idA2, idA1, endID, 'X')
	docB.RemoteInsert(idA3, idA2, endID, 'P')
	docB.RemoteInsert(idA4, idA3, endID, 'Y')

	// Insert W after A in both replicas.
	_, idA5 := docA.InsertElement(1, 'W')
	docB.RemoteInsert(idA5, idA1, idA2, 'W')

	contentA := docA.VisibleContent()
	contentB := docB.VisibleContent()

	if contentA != contentB {
		t.Errorf(
			"Visible content did not converge: replica A = %q, replica B = %q",
			contentA,
			contentB,
		)
	}

	internalA := docA.PrintInternal()
	internalB := docB.PrintInternal()

	if internalA != internalB {
		t.Errorf(
			"Internal origin structure differs:\nreplica A: %s\nreplica B: %s",
			internalA,
			internalB,
		)
	}
}

// Concurrency Test: InsertElementAndGetID inserts a local element and returns its generated ID.
func insertElementAndGetID(
	t *testing.T,
	document *Document,
	index int,
	content byte,
) ID {
	t.Helper()

	err, id := document.InsertElement(index, content)
	if err != nil {
		t.Fatalf(
			"InsertElement(%d, %q) returned an unexpected error: %v",
			index,
			content,
			err,
		)
	}

	return id
}

// Concurrency Test: RemoteInsertOrFail applies a remote insert operation and fails the test when the operation returns an unexpected error.
func remoteInsertOrFail(
	t *testing.T,
	document *Document,
	newID ID,
	originID ID,
	rightID ID,
	content byte,
) {
	t.Helper()

	if err := document.RemoteInsert(newID, originID, rightID, content); err != nil {
		t.Fatalf(
			"RemoteInsert(%v, %v, %v, %q) returned an unexpected error: %v",
			newID,
			originID,
			rightID,
			content,
			err,
		)
	}
}

// Concurrency Test: RemoteInsertPendingOrFail applies a remote insert operation that is expected to remain buffered until its dependencies arrive.
func remoteInsertPendingOrFail(
	t *testing.T,
	document *Document,
	newID ID,
	originID ID,
	rightID ID,
	content byte,
) {
	t.Helper()

	err := document.RemoteInsert(newID, originID, rightID, content)
	if err == nil {
		t.Fatalf(
			"RemoteInsert(%v, %v, %v, %q) returned nil; expected %q",
			newID,
			originID,
			rightID,
			content,
			"Pending value.",
		)
	}

	if err.Error() != "Pending value." {
		t.Fatalf(
			"RemoteInsert(%v, %v, %v, %q) returned %v; expected %q",
			newID,
			originID,
			rightID,
			content,
			err,
			"Pending value.",
		)
	}
}

// Concurrency Test: RemoteDeleteOrFail applies a remote delete operation and fails the test when the operation returns an unexpected error.
func remoteDeleteOrFail(
	t *testing.T,
	document *Document,
	elementID ID,
) {
	t.Helper()

	if err := document.RemoteDelete(elementID); err != nil {
		t.Fatalf(
			"RemoteDelete(%v) returned an unexpected error: %v",
			elementID,
			err,
		)
	}
}

// Concurrency Test: TestPendingInsertOutOfOrder verifies that an insert whose origin is missing remains pending and is automatically integrated when its origin arrives.
func TestPendingInsertOutOfOrder(t *testing.T) {
	startID := ID{
		ClientID: "START",
		Clock:    -1,
	}

	endID := ID{
		ClientID: "END",
		Clock:    -2,
	}

	source := NewDocument()
	target := NewDocument()

	idA := insertElementAndGetID(t, source, 0, 'A')
	idB := insertElementAndGetID(t, source, 1, 'B')

	// B arrives before its origin A.
	remoteInsertPendingOrFail(t, target, idB, idA, endID, 'B')

	if got := len(target.PendingInserts); got != 1 {
		t.Errorf(
			"len(PendingInserts) = %d after receiving B; expected 1",
			got,
		)
	}

	if got := target.VisibleContent(); got != "" {
		t.Errorf(
			"VisibleContent() = %q before A arrives; expected an empty string",
			got,
		)
	}

	// A arrives and automatically unlocks B.
	remoteInsertOrFail(t, target, idA, startID, endID, 'A')

	if got := target.VisibleContent(); got != "AB" {
		t.Errorf("VisibleContent() = %q; expected %q", got, "AB")
	}

	if got := len(target.PendingInserts); got != 0 {
		t.Errorf(
			"len(PendingInserts) = %d after processing A; expected 0",
			got,
		)
	}
}

// Concurrency Test: TestLongPendingInsertChain verifies that processPending repeatedly resolves a long causal chain until no pending insert operations remain.
func TestLongPendingInsertChain(t *testing.T) {
	startID := ID{
		ClientID: "START",
		Clock:    -1,
	}

	endID := ID{
		ClientID: "END",
		Clock:    -2,
	}

	source := NewDocument()
	target := NewDocument()

	idA := insertElementAndGetID(t, source, 0, 'A')
	idB := insertElementAndGetID(t, source, 1, 'B')
	idC := insertElementAndGetID(t, source, 2, 'C')

	// C depends on B, which is not available yet.
	remoteInsertPendingOrFail(t, target, idC, idB, endID, 'C')

	// B depends on A, which is also not available yet.
	remoteInsertPendingOrFail(t, target, idB, idA, endID, 'B')

	if got := len(target.PendingInserts); got != 2 {
		t.Errorf(
			"len(PendingInserts) = %d before A arrives; expected 2",
			got,
		)
	}

	// A unlocks B, and B then unlocks C.
	remoteInsertOrFail(t, target, idA, startID, endID, 'A')

	if got := target.VisibleContent(); got != "ABC" {
		t.Errorf("VisibleContent() = %q; expected %q", got, "ABC")
	}

	if got := target.VisibleLength(); got != 3 {
		t.Errorf("VisibleLength() = %d; expected 3", got)
	}

	if got := len(target.PendingInserts); got != 0 {
		t.Errorf(
			"len(PendingInserts) = %d after processing the chain; expected 0",
			got,
		)
	}
}

// Concurrency Test: TestDeleteBeforeInsert verifies that a delete received before its matching insert remains pending and is applied immediately when the element arrives.
func TestDeleteBeforeInsert(t *testing.T) {
	startID := ID{
		ClientID: "START",
		Clock:    -1,
	}

	endID := ID{
		ClientID: "END",
		Clock:    -2,
	}

	source := NewDocument()
	target := NewDocument()

	idX := insertElementAndGetID(t, source, 0, 'X')

	// The delete arrives before X exists in the target replica.
	remoteDeleteOrFail(t, target, idX)

	if got := len(target.PendingDeletes); got != 1 {
		t.Errorf(
			"len(PendingDeletes) = %d before X arrives; expected 1",
			got,
		)
	}

	// X arrives and must be tombstoned immediately.
	remoteInsertOrFail(t, target, idX, startID, endID, 'X')

	if got := target.VisibleContent(); got != "" {
		t.Errorf("VisibleContent() = %q; expected an empty string", got)
	}

	if got := target.VisibleLength(); got != 0 {
		t.Errorf("VisibleLength() = %d; expected 0", got)
	}

	if got := len(target.PendingDeletes); got != 0 {
		t.Errorf(
			"len(PendingDeletes) = %d after X arrives; expected 0",
			got,
		)
	}

	nodes := target.Traverse()

	if got := len(nodes); got != 3 {
		t.Errorf(
			"len(Traverse()) = %d; expected 3 nodes: START, X(X), and END",
			got,
		)
	}

	if internal := target.PrintInternal(); !strings.Contains(internal, "X(X)") {
		t.Errorf(
			"PrintInternal() = %q; expected it to contain the X(X) tombstone",
			internal,
		)
	}
}

// Concurrency Test: TestOutOfOrderInsertAndDelete verifies that pending inserts and pending deletes are processed correctly across a causal insertion chain.
func TestOutOfOrderInsertAndDelete(t *testing.T) {
	startID := ID{
		ClientID: "START",
		Clock:    -1,
	}

	endID := ID{
		ClientID: "END",
		Clock:    -2,
	}

	source := NewDocument()
	target := NewDocument()

	idA := insertElementAndGetID(t, source, 0, 'A')
	idB := insertElementAndGetID(t, source, 1, 'B')
	idC := insertElementAndGetID(t, source, 2, 'C')

	// Delete C before C exists.
	remoteDeleteOrFail(t, target, idC)

	// C arrives before B and must remain pending.
	remoteInsertPendingOrFail(t, target, idC, idB, endID, 'C')

	// B arrives before A and must also remain pending.
	remoteInsertPendingOrFail(t, target, idB, idA, endID, 'B')

	if got := len(target.PendingDeletes); got != 1 {
		t.Errorf(
			"len(PendingDeletes) = %d before A arrives; expected 1",
			got,
		)
	}

	if got := len(target.PendingInserts); got != 2 {
		t.Errorf(
			"len(PendingInserts) = %d before A arrives; expected 2",
			got,
		)
	}

	// A unlocks B, B unlocks C, and C receives its pending delete.
	remoteInsertOrFail(t, target, idA, startID, endID, 'A')

	if got := target.VisibleContent(); got != "AB" {
		t.Errorf("VisibleContent() = %q; expected %q", got, "AB")
	}

	if got := target.VisibleLength(); got != 2 {
		t.Errorf("VisibleLength() = %d; expected 2", got)
	}

	if got := len(target.PendingInserts); got != 0 {
		t.Errorf(
			"len(PendingInserts) = %d after processing the chain; expected 0",
			got,
		)
	}

	if got := len(target.PendingDeletes); got != 0 {
		t.Errorf(
			"len(PendingDeletes) = %d after processing the chain; expected 0",
			got,
		)
	}

	internal := target.PrintInternal()

	if !strings.Contains(internal, "A -> B -> C(X)") {
		t.Errorf(
			"PrintInternal() = %q; expected it to contain %q",
			internal,
			"A -> B -> C(X)",
		)
	}
}

// Concurrency Test: TestConvergenceWithArbitraryDelivery verifies that replicas converge to the same visible and internal state despite receiving operations in different causal delivery orders.
func TestConvergenceWithArbitraryDelivery(t *testing.T) {
	startID := ID{
		ClientID: "START",
		Clock:    -1,
	}

	endID := ID{
		ClientID: "END",
		Clock:    -2,
	}

	docA := NewDocument()
	docB := NewDocument()
	docC := NewDocument()

	idA := insertElementAndGetID(t, docA, 0, 'A')
	idB := insertElementAndGetID(t, docA, 1, 'B')
	idC := insertElementAndGetID(t, docA, 2, 'C')

	// Replica B receives C, A, and B.
	remoteInsertPendingOrFail(t, docB, idC, idB, endID, 'C')
	remoteInsertOrFail(t, docB, idA, startID, endID, 'A')
	remoteInsertOrFail(t, docB, idB, idA, endID, 'B')

	// Replica C receives B, C, and A.
	remoteInsertPendingOrFail(t, docC, idB, idA, endID, 'B')
	remoteInsertPendingOrFail(t, docC, idC, idB, endID, 'C')
	remoteInsertOrFail(t, docC, idA, startID, endID, 'A')

	contentA := docA.VisibleContent()
	contentB := docB.VisibleContent()
	contentC := docC.VisibleContent()

	if contentA != "ABC" {
		t.Errorf("Replica A VisibleContent() = %q; expected %q", contentA, "ABC")
	}

	if contentA != contentB {
		t.Errorf(
			"Visible content did not converge: replica A = %q, replica B = %q",
			contentA,
			contentB,
		)
	}

	if contentB != contentC {
		t.Errorf(
			"Visible content did not converge: replica B = %q, replica C = %q",
			contentB,
			contentC,
		)
	}

	internalA := docA.PrintInternal()
	internalB := docB.PrintInternal()
	internalC := docC.PrintInternal()

	if internalA != internalB {
		t.Errorf(
			"Internal state did not converge:\nreplica A: %s\nreplica B: %s",
			internalA,
			internalB,
		)
	}

	if internalB != internalC {
		t.Errorf(
			"Internal state did not converge:\nreplica B: %s\nreplica C: %s",
			internalB,
			internalC,
		)
	}

	if got := len(docB.PendingInserts); got != 0 {
		t.Errorf("Replica B has %d pending inserts; expected 0", got)
	}

	if got := len(docC.PendingInserts); got != 0 {
		t.Errorf("Replica C has %d pending inserts; expected 0", got)
	}

	if got := len(docB.PendingDeletes); got != 0 {
		t.Errorf("Replica B has %d pending deletes; expected 0", got)
	}

	if got := len(docC.PendingDeletes); got != 0 {
		t.Errorf("Replica C has %d pending deletes; expected 0", got)
	}
}
