package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"

	clientS "github.com/john6604/yata-collaborative-editor/internal/client"
	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/sync"
)

type editor struct {
	client *clientS.CollaborativeClient
	output io.Writer
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, input io.Reader, output, errorOutput io.Writer) int {
	flags := flag.NewFlagSet("editor", flag.ContinueOnError)
	flags.SetOutput(errorOutput)

	server := flags.String("server", "ws://localhost:8181/ws", "Server")
	room := flags.String("room", "room-1", "Room")
	clientFlag := flags.String("client", "", "Client")

	errFlag := flags.Parse(args)

	if errFlag != nil {
		return 2
	}

	if *clientFlag == "" {
		fmt.Fprintf(errorOutput, "client flag is required, usage: -client <client>\n")
		return 2
	}

	if *server == "" || *room == "" {
		return 2
	}

	clientCollaborative, errClient := clientS.NewCollaborativeClient(*server, *room, *clientFlag)
	if errClient != nil {
		fmt.Fprintf(errorOutput, "client finished with error: %v\n", errClient)
		return 1
	}

	repl := &editor{
		client: clientCollaborative,
		output: output,
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	clientCollaborative.Start()
	defer clientCollaborative.Close()

	replErr := repl.run(input, signals)
	exitCode := 0
	if replErr != nil {
		fmt.Fprintf(errorOutput, "editor finished with error: %v\n", replErr)
		exitCode = 1
	}

	return exitCode
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
			fmt.Fprintf(e.output, "\nSignal %s received. Closing...\n", received)
			return nil

		case line, ok := <-lines:
			if !ok {
				if err := <-scanResult; err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				fmt.Fprintln(e.output, "\nInput closed. Closing...")
				return nil
			}

			exit, err := e.execute(line)
			if err != nil {
				fmt.Fprintf(e.output, "error: %v\n", err)
				continue
			}
			if exit {
				fmt.Fprintln(e.output, "Closing...")
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
			return false, errors.New("usage: insert <index> <character>")
		}

		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return false, fmt.Errorf("invalid index %q", fields[1])
		}

		characters := []rune(fields[2])
		if len(characters) != 1 {
			return false, errors.New("<character> must contain exactly one character")
		}
		character := characters[0]

		e.client.DocMutex.Lock()

		err, operationID := e.client.Doc.InsertElement(index, character)
		if err != nil {
			e.client.DocMutex.Unlock()
			return false, err
		}

		operation := e.client.Doc.InsertLog[operationID]

		if operation == nil {
			e.client.DocMutex.Unlock()
			return false, errors.New("operation not found")
		}

		fmt.Fprintf(e.output, "Inserted %q at index %d.\n", character, index)

		operationCopy := *operation
		currentDocument := e.client.Doc.String()

		e.client.DocMutex.Unlock()

		encodedOperation, errEncode := sync.EncodeInsertOperation(operationCopy)

		if errEncode != nil {
			return false, errors.New("operation failed to encode")
		}

		errSend := e.client.SendOrQueue(encodedOperation)

		if errSend != nil {
			return false, errors.New("operation failed to send through websocket")
		}

		state := e.client.GetState()
		if state == clientS.StateOnline {
			errSnapshot := e.client.SendCurrentSnapshot()
			if errSnapshot != nil {
				fmt.Fprintf(e.output, "snapshot failed to send\n")
			}
		}

		fmt.Fprintf(e.output, "Local document: %s.\n", currentDocument)

		return false, nil

	case "delete":

		if len(fields) != 2 {
			return false, errors.New("usage: delete <index>")
		}

		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return false, fmt.Errorf("invalid index %q", fields[1])
		}

		e.client.DocMutex.Lock()

		err, operationID := e.client.Doc.Delete(index)

		if err != nil {
			e.client.DocMutex.Unlock()
			return false, err
		}

		operation := e.client.Doc.DeleteLog[operationID]
		if operation == nil {
			e.client.DocMutex.Unlock()
			return false, errors.New("operation not found")
		}

		fmt.Fprintf(e.output, "Deleted at index %d.\n", index)

		operationCopy := *operation
		currentDocument := e.client.Doc.String()

		e.client.DocMutex.Unlock()

		encodedOperation, errEncode := sync.EncodeDeleteOperation(operationCopy)

		if errEncode != nil {
			return false, errors.New("operation failed to encode")
		}

		errSend := e.client.SendOrQueue(encodedOperation)

		if errSend != nil {
			return false, errors.New("operation failed to send through websocket")
		}

		state := e.client.GetState()
		if state == clientS.StateOnline {
			errSnapshot := e.client.SendCurrentSnapshot()
			if errSnapshot != nil {
				fmt.Fprintf(e.output, "snapshot failed to send\n")
			}
		}

		fmt.Fprintf(e.output, "Local document: %s.\n", currentDocument)

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
		fmt.Fprintln(e.output, "Commands: print, insert <index> <char>, delete <index>, state, exit")
		return false, nil

	default:
		return false, fmt.Errorf("unknown command %q", fields[0])
	}
}

func (e *editor) printDocument() {
	e.client.DocMutex.Lock()
	visible := e.client.Doc.VisibleContent()
	internal := e.client.Doc.PrintInternal()
	e.client.DocMutex.Unlock()

	fmt.Fprintf(e.output, "Visible: %q.\n", visible)
	fmt.Fprintf(e.output, "Internal: %q.\n", internal)
}

func (e *editor) printState() {
	e.client.DocMutex.Lock()
	fmt.Fprintf(e.output, "ClientID: %s\n", e.client.Doc.ClientID)
	fmt.Fprintf(e.output, "Clock: %d\n", e.client.Doc.Clock)
	fmt.Fprintf(e.output, "VisibleLength: %d\n", e.client.Doc.VisibleLength())

	insertIDs := sortedIDsFromInsertLog(e.client.Doc)
	fmt.Fprintf(e.output, "InsertLog (%d):\n", len(insertIDs))
	for _, id := range insertIDs {
		operation := e.client.Doc.InsertLog[id]
		fmt.Fprintf(
			e.output,
			"  %s origin=%s right=%s content=%q\n",
			formatID(operation.NewID),
			formatID(operation.OriginID),
			formatID(operation.RightID),
			operation.Content,
		)
	}

	deleteIDs := sortedIDsFromDeleteLog(e.client.Doc)
	fmt.Fprintf(e.output, "DeleteLog (%d):\n", len(deleteIDs))
	for _, id := range deleteIDs {
		fmt.Fprintf(e.output, "  %s\n", formatID(id))
	}

	pendingInsertIDs := sortedIDKeys(e.client.Doc.PendingInserts)
	fmt.Fprintf(e.output, "PendingInserts (%d): %s\n", len(pendingInsertIDs), formatIDList(pendingInsertIDs))

	pendingDeleteIDs := sortedIDKeys(e.client.Doc.PendingDeletes)
	fmt.Fprintf(e.output, "PendingDeletes (%d): %s\n", len(pendingDeleteIDs), formatIDList(pendingDeleteIDs))

	e.client.DocMutex.Unlock()
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
