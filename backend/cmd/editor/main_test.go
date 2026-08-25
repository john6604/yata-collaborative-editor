package main

import (
	"bytes"
	"testing"

	clientS "github.com/john6604/yata-collaborative-editor/internal/client"
)

func newTestEditor(t *testing.T) (*editor, *clientS.CollaborativeClient, *bytes.Buffer) {
	t.Helper()

	client, err := clientS.NewCollaborativeClient("ws://localhost:8181/ws", "room-1", "client-A")
	if err != nil {
		t.Fatalf("NewCollaborativeClient() returned an unexpected error: %v", err)
	}

	output := &bytes.Buffer{}
	repl := &editor{
		client: client,
		output: output,
	}

	return repl, client, output
}

func TestInsertAndDeleteUpdateDocumentAndQueueOperations(t *testing.T) {
	repl, client, _ := newTestEditor(t)

	exit, err := repl.execute("insert 0 A")
	if err != nil {
		t.Fatalf("insert returned an unexpected error: %v", err)
	}
	if exit {
		t.Fatal("insert unexpectedly requested exit")
	}
	if got, want := client.Doc.VisibleContent(), "A"; got != want {
		t.Fatalf("VisibleContent after insert = %q; expected %q", got, want)
	}
	if got, want := client.OfflineQueueLength(), 1; got != want {
		t.Fatalf("OfflineQueue length after insert = %d; expected %d", got, want)
	}

	exit, err = repl.execute("delete 0")
	if err != nil {
		t.Fatalf("delete returned an unexpected error: %v", err)
	}
	if exit {
		t.Fatal("delete unexpectedly requested exit")
	}
	if got, want := client.Doc.VisibleContent(), ""; got != want {
		t.Fatalf("VisibleContent after delete = %q; expected %q", got, want)
	}
	if got, want := client.OfflineQueueLength(), 2; got != want {
		t.Fatalf("OfflineQueue length after delete = %d; expected %d", got, want)
	}
	if got, want := client.Doc.PrintInternal(), "START -> A(X) -> END"; got != want {
		t.Fatalf("PrintInternal after delete = %q; expected %q", got, want)
	}
}

func TestInsertUnicodeRune(t *testing.T) {
	repl, client, _ := newTestEditor(t)

	exit, err := repl.execute("insert 0 \U0001F600")
	if err != nil {
		t.Fatalf("insert returned an unexpected error: %v", err)
	}
	if exit {
		t.Fatal("insert unexpectedly requested exit")
	}
	if got, want := client.Doc.VisibleContent(), "\U0001F600"; got != want {
		t.Fatalf("VisibleContent after insert = %q; expected %q", got, want)
	}
	if got, want := client.Doc.VisibleLength(), 1; got != want {
		t.Fatalf("VisibleLength after insert = %d; expected %d", got, want)
	}
	if got, want := client.OfflineQueueLength(), 1; got != want {
		t.Fatalf("OfflineQueue length after insert = %d; expected %d", got, want)
	}
}

func TestInsertRejectsMultipleRunes(t *testing.T) {
	repl, client, _ := newTestEditor(t)

	_, err := repl.execute("insert 0 ab")
	if err == nil {
		t.Fatal("insert returned nil; expected an error")
	}
	if got := client.OfflineQueueLength(); got != 0 {
		t.Fatalf("OfflineQueue length after rejected insert = %d; expected 0", got)
	}
	if got := client.Doc.VisibleContent(); got != "" {
		t.Fatalf("VisibleContent after rejected insert = %q; expected an empty string", got)
	}
}

func TestReadOnlyCommandsDoNotQueueOperations(t *testing.T) {
	repl, client, _ := newTestEditor(t)

	for _, command := range []string{"print", "state", "help"} {
		if _, err := repl.execute(command); err != nil {
			t.Fatalf("%s returned an unexpected error: %v", command, err)
		}
	}

	if got := client.OfflineQueueLength(); got != 0 {
		t.Fatalf("read-only commands queued %d messages; expected 0", got)
	}
}
