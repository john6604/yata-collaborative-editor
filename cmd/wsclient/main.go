package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/client"
	"github.com/john6604/yata-collaborative-editor/internal/document"
	internalSync "github.com/john6604/yata-collaborative-editor/internal/sync"
)

type wsEditor struct {
	document *document.Document
	conn     *websocket.Conn
	mutex    *sync.Mutex
	output   io.Writer
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, input io.Reader, output, errorOutput io.Writer) int {
	flags := flag.NewFlagSet("wsclient", flag.ContinueOnError)
	flags.SetOutput(errorOutput)

	doc := document.NewDocument()

	var mutexDoc sync.Mutex

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8181/ws", nil)

	if err != nil {
		fmt.Fprintf(errorOutput, "wsclient finished with error: %v\n", err)
		return 1
	}

	fmt.Fprintf(output, "Connected to relay server...\n")

	defer conn.Close()

	jsonStr := `{"version":1,"type":"join","payload":{"room":"room-1","client_id":"client-B"}}`
	data := []byte(jsonStr)

	errSend := conn.WriteMessage(websocket.TextMessage, data)

	if errSend != nil {
		fmt.Fprintf(errorOutput, "wsclient finished with error: %v\n", errSend)
		return 1
	}

	_, _, errMessage := conn.ReadMessage()

	if errMessage != nil {
		fmt.Fprintf(errorOutput, "wsclient finished with error: %v\n", errMessage)
		return 1
	}

	fmt.Fprintf(output, "Joined room-1 as client-B.\n")
	fmt.Fprintln(output, "Commands: help, print, insert <index> <char>, exit")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	repl := &wsEditor{document: doc, conn: conn, mutex: &mutexDoc, output: output}
	go client.RemoteMessageLoop(doc, conn, &mutexDoc)
	errREPL := repl.run(input, signals)

	signal.Stop(signals)

	exitCode := 0
	if errREPL != nil {
		fmt.Fprintf(errorOutput, "wsclient finished with error: %v\n", errREPL)
		exitCode = 1
	}

	return exitCode
}

func (editor *wsEditor) run(input io.Reader, signals <-chan os.Signal) error {
	lines := make(chan string)
	scannResult := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			lines <- scanner.Text()
		}

		scannResult <- scanner.Err()
		close(lines)
	}()

	for {
		fmt.Fprintf(editor.output, "ws> ")

		select {
		case received := <-signals:
			fmt.Fprintf(editor.output, "\nSignal %s received. Saving and closing...\n", received)
			return nil
		case line, ok := <-lines:
			if !ok {
				if err := <-scannResult; err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				fmt.Fprintln(editor.output, "\nInput closed. Saving and closing...")
				return nil
			}

			exit, err := editor.execute(line)

			if err != nil {
				fmt.Fprintf(editor.output, "error: %v\n", err)
				continue
			}

			if exit {
				fmt.Fprintln(editor.output, "Saving and closing...")
				return nil
			}
		}
	}
}

func (editor *wsEditor) execute(line string) (bool, error) {
	fields := strings.Fields(strings.TrimSpace(line))

	if len(fields) == 0 {
		return false, nil
	}

	switch strings.ToLower(fields[0]) {
	case "help":
		if len(fields) != 1 {
			return false, errors.New("usage: help")
		}
		fmt.Fprintln(editor.output, "Commands: help, print, insert <index> <char>, exit")
		return false, nil
	case "print":
		if len(fields) != 1 {
			return false, errors.New("usage: print")
		}
		editor.printDocument()
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

		editor.mutex.Lock()

		err, operationID := editor.document.InsertElement(index, character)
		if err != nil {
			editor.mutex.Unlock()
			return false, err
		}

		operation := editor.document.InsertLog[operationID]

		if operation == nil {
			editor.mutex.Unlock()
			return false, errors.New("operation not found")
		}

		fmt.Fprintf(editor.output, "Inserted %q at index %d.\n", character, index)

		operationCopy := *operation
		currentDocument := editor.document.String()

		editor.mutex.Unlock()

		encodedOperation, errEncode := internalSync.EncodeInsertOperation(operationCopy)

		if errEncode != nil {
			return false, errors.New("operation failed to encode")
		}

		errSend := editor.conn.WriteMessage(websocket.TextMessage, encodedOperation)

		if errSend != nil {
			return false, errors.New("operation failed to send through websocket")
		}

		fmt.Fprintf(editor.output, "Local document: %s.\n", currentDocument)

		return false, nil
	case "delete":
		if len(fields) != 2 {
			return false, errors.New("usage: delete <index>")
		}

		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return false, fmt.Errorf("invalid index %q", fields[1])
		}

		editor.mutex.Lock()

		err, operationID := editor.document.Delete(index)

		if err != nil {
			editor.mutex.Unlock()
			return false, err
		}

		operation := editor.document.DeleteLog[operationID]
		if operation == nil {
			editor.mutex.Unlock()
			return false, errors.New("operation not found")
		}

		fmt.Fprintf(editor.output, "Deleted at index %d.\n", index)

		operationCopy := *operation
		currentDocument := editor.document.String()

		editor.mutex.Unlock()

		encodedOperation, errEncode := internalSync.EncodeDeleteOperation(operationCopy)

		if errEncode != nil {
			return false, errors.New("operation failed to encode")
		}

		errSend := editor.conn.WriteMessage(websocket.TextMessage, encodedOperation)

		if errSend != nil {
			return false, errors.New("operation failed to send through websocket")
		}

		fmt.Fprintf(editor.output, "Local document: %s.\n", currentDocument)

		return false, nil
	case "exit":
		if len(fields) != 1 {
			return false, errors.New("usage: exit")
		}
		return true, nil
	default:
		return false, fmt.Errorf("unknown command %q", fields[0])
	}
}

func (editor *wsEditor) printDocument() {
	editor.mutex.Lock()
	visible := editor.document.VisibleContent()
	internal := editor.document.PrintInternal()
	editor.mutex.Unlock()

	fmt.Fprintf(editor.output, "Visible: %s.\n", visible)
	fmt.Fprintf(editor.output, "Internal: %s.\n", internal)
}
