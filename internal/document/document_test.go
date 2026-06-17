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

		if err := document.InsertElement(index, character); err != nil {
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

	if err := document.InsertElement(0, 'X'); err != nil {
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

	if err := document.InsertElement(2, 'Y'); err != nil {
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

	if err := document.InsertElement(1, 'a'); err != nil {
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
				return document.InsertElement(100, 'X')
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
