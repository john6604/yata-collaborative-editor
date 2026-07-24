package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/john6604/yata-collaborative-editor/internal/document"
)

type recordingSaver struct {
	calls int
	err   error
}

func (s *recordingSaver) SaveSnapshot(*document.Document) error {
	s.calls++
	return s.err
}

func TestInsertAndDeletePersistImmediately(t *testing.T) {
	saver := &recordingSaver{}
	doc := document.NewDocument()
	output := &bytes.Buffer{}
	repl := &editor{document: doc, storage: saver, output: output}

	exit, err := repl.execute("insert 0 A")
	if err != nil {
		t.Fatalf("insert returned an unexpected error: %v", err)
	}
	if exit {
		t.Fatal("insert unexpectedly requested exit")
	}
	if got, want := saver.calls, 1; got != want {
		t.Fatalf("SaveSnapshot calls after insert = %d; expected %d", got, want)
	}
	if got, want := doc.VisibleContent(), "A"; got != want {
		t.Fatalf("VisibleContent after insert = %q; expected %q", got, want)
	}

	exit, err = repl.execute("delete 0")
	if err != nil {
		t.Fatalf("delete returned an unexpected error: %v", err)
	}
	if exit {
		t.Fatal("delete unexpectedly requested exit")
	}
	if got, want := saver.calls, 2; got != want {
		t.Fatalf("SaveSnapshot calls after delete = %d; expected %d", got, want)
	}
	if got, want := doc.VisibleContent(), ""; got != want {
		t.Fatalf("VisibleContent after delete = %q; expected %q", got, want)
	}
	if got, want := doc.PrintInternal(), "START -> A(X) -> END"; got != want {
		t.Fatalf("PrintInternal after delete = %q; expected %q", got, want)
	}
}

func TestInsertUnicodeRune(t *testing.T) {
	saver := &recordingSaver{}
	doc := document.NewDocument()
	repl := &editor{document: doc, storage: saver, output: &bytes.Buffer{}}

	exit, err := repl.execute("insert 0 😀")
	if err != nil {
		t.Fatalf("insert returned an unexpected error: %v", err)
	}
	if exit {
		t.Fatal("insert unexpectedly requested exit")
	}
	if got, want := saver.calls, 1; got != want {
		t.Fatalf("SaveSnapshot calls after insert = %d; expected %d", got, want)
	}
	if got, want := doc.VisibleContent(), "😀"; got != want {
		t.Fatalf("VisibleContent after insert = %q; expected %q", got, want)
	}
	if got, want := doc.VisibleLength(), 1; got != want {
		t.Fatalf("VisibleLength after insert = %d; expected %d", got, want)
	}
}

func TestInsertRejectsMultipleRunes(t *testing.T) {
	saver := &recordingSaver{}
	doc := document.NewDocument()
	repl := &editor{document: doc, storage: saver, output: &bytes.Buffer{}}

	_, err := repl.execute("insert 0 ab")
	if err == nil {
		t.Fatal("insert returned nil; expected an error")
	}
	if got := saver.calls; got != 0 {
		t.Fatalf("SaveSnapshot calls after rejected insert = %d; expected 0", got)
	}
	if got := doc.VisibleContent(); got != "" {
		t.Fatalf("VisibleContent after rejected insert = %q; expected an empty string", got)
	}
}

func TestMutationReportsPersistenceFailure(t *testing.T) {
	saver := &recordingSaver{err: errors.New("disk full")}
	doc := document.NewDocument()
	repl := &editor{document: doc, storage: saver, output: &bytes.Buffer{}}

	_, err := repl.execute("insert 0 A")
	if err == nil {
		t.Fatal("insert returned nil; expected persistence error")
	}
	if got, want := saver.calls, 1; got != want {
		t.Fatalf("SaveSnapshot calls = %d; expected %d", got, want)
	}
	if got, want := doc.VisibleContent(), "A"; got != want {
		t.Fatalf("mutation should remain visible in memory: got %q; expected %q", got, want)
	}
}

func TestReadOnlyCommandsDoNotPersist(t *testing.T) {
	saver := &recordingSaver{}
	repl := &editor{
		document: document.NewDocument(),
		storage:  saver,
		output:   &bytes.Buffer{},
	}

	for _, command := range []string{"print", "state", "help"} {
		if _, err := repl.execute(command); err != nil {
			t.Fatalf("%s returned an unexpected error: %v", command, err)
		}
	}

	if got := saver.calls; got != 0 {
		t.Fatalf("read-only commands caused %d saves; expected 0", got)
	}
}
