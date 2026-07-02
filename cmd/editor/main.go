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

	databasePath := flags.String("db", configuredDatabasePath(), "ruta del archivo BoltDB")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	absolutePath, err := filepath.Abs(*databasePath)
	if err != nil {
		fmt.Fprintf(errorOutput, "no se pudo resolver la ruta de la base: %v\n", err)
		return 1
	}

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		fmt.Fprintf(errorOutput, "no se pudo crear el directorio de la base: %v\n", err)
		return 1
	}

	store := &storage.Storage{}
	if err := store.OpenDB(absolutePath); err != nil {
		fmt.Fprintf(errorOutput, "no se pudo abrir BoltDB: %v\n", err)
		return 1
	}

	doc := document.NewDocument()
	if err := doc.ReconstructDocument(store); err != nil {
		fmt.Fprintf(output, "No se encontró un snapshot válido (%v). Se creó un documento nuevo.\n", err)
		doc = document.NewDocument()

		// Persist the newly generated ClientID immediately. Even an empty editor
		// therefore keeps the same identity after a clean restart.
		if err := store.SaveSnapshot(doc); err != nil {
			fmt.Fprintf(errorOutput, "no se pudo guardar el documento inicial: %v\n", err)
			_ = store.CloseDB()
			return 1
		}
	} else {
		fmt.Fprintf(output, "Snapshot restaurado correctamente.\n")
	}

	fmt.Fprintf(output, "Base de datos: %s\n", absolutePath)
	fmt.Fprintln(output, "Comandos: print, insert <index> <char>, delete <index>, save, state, exit")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	repl := &editor{document: doc, storage: store, output: output}
	replErr := repl.run(input, signals)

	signal.Stop(signals)

	exitCode := 0
	if replErr != nil {
		fmt.Fprintf(errorOutput, "editor finalizado con error: %v\n", replErr)
		exitCode = 1
	}

	// The final save is intentionally performed for exit, EOF, SIGINT and
	// SIGTERM. Mutating commands already save eagerly, so this is the last
	// durability barrier before closing BoltDB.
	if err := store.SaveSnapshot(doc); err != nil {
		fmt.Fprintf(errorOutput, "no se pudo guardar el snapshot final: %v\n", err)
		exitCode = 1
	}
	if err := store.CloseDB(); err != nil {
		fmt.Fprintf(errorOutput, "no se pudo cerrar BoltDB: %v\n", err)
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
			fmt.Fprintf(e.output, "\nSeñal %s recibida. Guardando y cerrando...\n", received)
			return nil

		case line, ok := <-lines:
			if !ok {
				if err := <-scanResult; err != nil {
					return fmt.Errorf("leer stdin: %w", err)
				}
				fmt.Fprintln(e.output, "\nEntrada cerrada. Guardando y cerrando...")
				return nil
			}

			exit, err := e.execute(line)
			if err != nil {
				fmt.Fprintf(e.output, "error: %v\n", err)
				continue
			}
			if exit {
				fmt.Fprintln(e.output, "Guardando y cerrando...")
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
			return false, errors.New("uso: print")
		}
		e.printDocument()
		return false, nil

	case "insert":
		if len(fields) != 3 {
			return false, errors.New("uso: insert <index> <char>")
		}

		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return false, fmt.Errorf("índice inválido %q", fields[1])
		}
		if len(fields[2]) != 1 {
			return false, errors.New("<char> debe contener exactamente un byte")
		}

		if err, _ := e.document.InsertElement(index, fields[2][0]); err != nil {
			return false, err
		}
		if err := e.storage.SaveSnapshot(e.document); err != nil {
			return false, fmt.Errorf("el insert se aplicó en memoria, pero no pudo persistirse: %w", err)
		}
		fmt.Fprintf(e.output, "Insertado %q en el índice %d.\n", fields[2][0], index)
		return false, nil

	case "delete":
		if len(fields) != 2 {
			return false, errors.New("uso: delete <index>")
		}

		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return false, fmt.Errorf("índice inválido %q", fields[1])
		}
		if err := e.document.Delete(index); err != nil {
			return false, err
		}
		if err := e.storage.SaveSnapshot(e.document); err != nil {
			return false, fmt.Errorf("el delete se aplicó en memoria, pero no pudo persistirse: %w", err)
		}
		fmt.Fprintf(e.output, "Eliminado el elemento visible del índice %d.\n", index)
		return false, nil

	case "save":
		if len(fields) != 1 {
			return false, errors.New("uso: save")
		}
		if err := e.storage.SaveSnapshot(e.document); err != nil {
			return false, fmt.Errorf("guardar snapshot: %w", err)
		}
		fmt.Fprintln(e.output, "Snapshot guardado.")
		return false, nil

	case "state":
		if len(fields) != 1 {
			return false, errors.New("uso: state")
		}
		e.printState()
		return false, nil

	case "exit":
		if len(fields) != 1 {
			return false, errors.New("uso: exit")
		}
		return true, nil

	case "help":
		fmt.Fprintln(e.output, "Comandos: print, insert <index> <char>, delete <index>, save, state, exit")
		return false, nil

	default:
		return false, fmt.Errorf("comando desconocido %q", fields[0])
	}
}

func (e *editor) printDocument() {
	fmt.Fprintf(e.output, "Visible:  %q\n", e.document.VisibleContent())
	fmt.Fprintf(e.output, "Interno:  %s\n", e.document.PrintInternal())
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
