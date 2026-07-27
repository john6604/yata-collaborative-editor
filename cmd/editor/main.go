package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/storage"
)

const defaultDatabasePath = "../../data/yata.db"

type snapshotSaver interface {
	SaveSnapshot(*document.Document) error
}

type editor struct {
	document *document.Document
	storage  snapshotSaver
	output   io.Writer
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, input io.Reader, output, errorOutput io.Writer) int {
	flags := flag.NewFlagSet("editor", flag.ContinueOnError)
	flags.SetOutput(errorOutput)

	databasePath := flags.String("db", configuredDatabasePath(), "BoltDB file path")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	absolutePath, err := filepath.Abs(*databasePath)
	if err != nil {
		fmt.Fprintf(errorOutput, "could not resolve database path: %v\n", err)
		return 1
	}

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		fmt.Fprintf(errorOutput, "could not create database directory: %v\n", err)
		return 1
	}

	store := &storage.Storage{}
	if err := store.OpenDB(absolutePath); err != nil {
		fmt.Fprintf(errorOutput, "could not open BoltDB: %v\n", err)
		return 1
	}

	doc := document.NewDocument()
	if err := doc.ReconstructDocument(store); err != nil {
		fmt.Fprintf(output, "No valid snapshot was found (%v). A new document was created.\n", err)
		doc = document.NewDocument()

		// Persist the newly generated ClientID immediately. Even an empty editor
		// therefore keeps the same identity after a clean restart.
		if err := store.SaveSnapshot(doc); err != nil {
			fmt.Fprintf(errorOutput, "could not save initial document: %v\n", err)
			_ = store.CloseDB()
			return 1
		}
	} else {
		fmt.Fprintf(output, "Snapshot restored successfully.\n")
	}

	fmt.Fprintf(output, "Database: %s\n", absolutePath)
	fmt.Fprintln(output, "Commands: print, insert <index> <char>, delete <index>, save, state, exit")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	repl := &editor{document: doc, storage: store, output: output}
	replErr := repl.run(input, signals)

	signal.Stop(signals)

	exitCode := 0
	if replErr != nil {
		fmt.Fprintf(errorOutput, "editor finished with error: %v\n", replErr)
		exitCode = 1
	}

	// The final save is intentionally performed for exit, EOF, SIGINT and
	// SIGTERM. Mutating commands already save eagerly, so this is the last
	// durability barrier before closing BoltDB.
	if err := store.SaveSnapshot(doc); err != nil {
		fmt.Fprintf(errorOutput, "could not save final snapshot: %v\n", err)
		exitCode = 1
	}
	if err := store.CloseDB(); err != nil {
		fmt.Fprintf(errorOutput, "could not close BoltDB: %v\n", err)
		exitCode = 1
	}

	return exitCode
}

func configuredDatabasePath() string {
	if path := strings.TrimSpace(os.Getenv("YATA_DB_PATH")); path != "" {
		return path
	}
	return defaultDatabasePath
}

func (e *editor) run(input io.Reader, signals <-chan os.Signal) error {
	lines := make(chan string)
	scanResult := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		scanResult <- scanner.Err()
		close(lines)
	}()

	for {
		fmt.Fprint(e.output, "> ")

		select {
		case received := <-signals:
			fmt.Fprintf(e.output, "\nSignal %s received. Saving and closing...\n", received)
			return nil

		case line, ok := <-lines:
			if !ok {
				if err := <-scanResult; err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				fmt.Fprintln(e.output, "\nInput closed. Saving and closing...")
				return nil
			}

			exit, err := e.execute(line)
			if err != nil {
				fmt.Fprintf(e.output, "error: %v\n", err)
				continue
			}
			if exit {
				fmt.Fprintln(e.output, "Saving and closing...")
				return nil
			}
		}
	}
}

func (e *editor) execute(line string) (bool, error) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) == 0 {
		return false, nil
	}

	switch strings.ToLower(fields[0]) {
	case "print":
		if len(fields) != 1 {
			return false, errors.New("usage: print")
		}
		e.printDocument()
		return false, nil

	case "insert":
		if len(fields) != 3 {
			return false, errors.New("usage: insert <index> <char>")
		}

		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return false, fmt.Errorf("invalid index %q", fields[1])
		}
		characters := []rune(fields[2])
		if len(characters) != 1 {
			return false, errors.New("<char> must contain exactly one character")
		}
		character := characters[0]

		if err, _ := e.document.InsertElement(index, character); err != nil {
			return false, err
		}
		if err := e.storage.SaveSnapshot(e.document); err != nil {
			return false, fmt.Errorf("insert was applied in memory, but could not be persisted: %w", err)
		}
		fmt.Fprintf(e.output, "Inserted %q at index %d.\n", character, index)
		return false, nil

	case "delete":
		if len(fields) != 2 {
			return false, errors.New("usage: delete <index>")
		}

		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return false, fmt.Errorf("invalid index %q", fields[1])
		}
		if err, _ := e.document.Delete(index); err != nil {
			return false, err
		}
		if err := e.storage.SaveSnapshot(e.document); err != nil {
			return false, fmt.Errorf("delete was applied in memory, but could not be persisted: %w", err)
		}
		fmt.Fprintf(e.output, "Deleted the visible element at index %d.\n", index)
		return false, nil

	case "save":
		if len(fields) != 1 {
			return false, errors.New("usage: save")
		}
		if err := e.storage.SaveSnapshot(e.document); err != nil {
			return false, fmt.Errorf("save snapshot: %w", err)
		}
		fmt.Fprintln(e.output, "Snapshot saved.")
		return false, nil

	case "state":
		if len(fields) != 1 {
			return false, errors.New("usage: state")
		}
		e.printState()
		return false, nil

	case "exit":
		if len(fields) != 1 {
			return false, errors.New("usage: exit")
		}
		return true, nil

	case "help":
		fmt.Fprintln(e.output, "Commands: print, insert <index> <char>, delete <index>, save, state, exit")
		return false, nil

	default:
		return false, fmt.Errorf("unknown command %q", fields[0])
	}
}

func (e *editor) printDocument() {
	fmt.Fprintf(e.output, "Visible:  %q\n", e.document.VisibleContent())
	fmt.Fprintf(e.output, "Internal: %s\n", e.document.PrintInternal())
}

func (e *editor) printState() {
	fmt.Fprintf(e.output, "ClientID: %s\n", e.document.ClientID)
	fmt.Fprintf(e.output, "Clock: %d\n", e.document.Clock)
	fmt.Fprintf(e.output, "VisibleLength: %d\n", e.document.VisibleLength())

	insertIDs := sortedIDsFromInsertLog(e.document)
	fmt.Fprintf(e.output, "InsertLog (%d):\n", len(insertIDs))
	for _, id := range insertIDs {
		operation := e.document.InsertLog[id]
		fmt.Fprintf(
			e.output,
			"  %s origin=%s right=%s content=%q\n",
			formatID(operation.NewID),
			formatID(operation.OriginID),
			formatID(operation.RightID),
			operation.Content,
		)
	}

	deleteIDs := sortedIDsFromDeleteLog(e.document)
	fmt.Fprintf(e.output, "DeleteLog (%d):\n", len(deleteIDs))
	for _, id := range deleteIDs {
		fmt.Fprintf(e.output, "  %s\n", formatID(id))
	}

	pendingInsertIDs := sortedIDKeys(e.document.PendingInserts)
	fmt.Fprintf(e.output, "PendingInserts (%d): %s\n", len(pendingInsertIDs), formatIDList(pendingInsertIDs))

	pendingDeleteIDs := sortedIDKeys(e.document.PendingDeletes)
	fmt.Fprintf(e.output, "PendingDeletes (%d): %s\n", len(pendingDeleteIDs), formatIDList(pendingDeleteIDs))
}

func sortedIDsFromInsertLog(doc *document.Document) []identifier.ID {
	ids := make([]identifier.ID, 0, len(doc.InsertLog))
	for id := range doc.InsertLog {
		ids = append(ids, id)
	}
	sortIDs(ids)
	return ids
}

func sortedIDsFromDeleteLog(doc *document.Document) []identifier.ID {
	ids := make([]identifier.ID, 0, len(doc.DeleteLog))
	for id := range doc.DeleteLog {
		ids = append(ids, id)
	}
	sortIDs(ids)
	return ids
}

func sortedIDKeys[T any](values map[identifier.ID]T) []identifier.ID {
	ids := make([]identifier.ID, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sortIDs(ids)
	return ids
}

func sortIDs(ids []identifier.ID) {
	sort.Slice(ids, func(i, j int) bool {
		if ids[i].ClientID == ids[j].ClientID {
			return ids[i].Clock < ids[j].Clock
		}
		return ids[i].ClientID < ids[j].ClientID
	})
}

func formatID(id identifier.ID) string {
	return fmt.Sprintf("(%s,%d)", id.ClientID, id.Clock)
}

func formatIDList(ids []identifier.ID) string {
	if len(ids) == 0 {
		return "[]"
	}

	formatted := make([]string, 0, len(ids))
	for _, id := range ids {
		formatted = append(formatted, formatID(id))
	}
	return "[" + strings.Join(formatted, ", ") + "]"
}
