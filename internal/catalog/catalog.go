package catalog

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DocumentMetadata struct {
	DocumentID string    `json:"document_id"`
	Name       string    `json:"document_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Store struct {
	db *sql.DB
}

func (store *Store) Close() error {
	if store == nil || store.db == nil {
		return nil
	}

	return store.db.Close()
}

func (store *Store) CreateDocument(id string, name string) (*DocumentMetadata, error) {
	document := DocumentMetadata{
		DocumentID: id,
		Name:       name,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, errCreate := store.db.Exec(
		`INSERT INTO documents(
			document_id,
			name,
			created_at,
			updated_at
		)
		VALUES(?, ?, ?, ?)`,
		document.DocumentID, document.Name, document.CreatedAt, document.UpdatedAt)

	if errCreate != nil {
		return nil, errCreate
	}

	return &document, nil
}

func (store *Store) RenameDocument(id string, name string) error {
	result, err := store.db.Exec(
		`UPDATE documents
		SET name = ?, 
		updated_at = ?
		WHERE document_id = ?`,
		name, time.Now(), id)

	if err != nil {
		return err
	}

	rows, errRows := result.RowsAffected()
	if errRows != nil {
		return errRows
	}

	if rows > 1 || rows == 0 {
		return errors.New("The operation affected a different amount of rows than expected\n")
	}

	return nil
}

func (store *Store) ListDocuments() ([]*DocumentMetadata, error) {
	var documents []*DocumentMetadata

	rows, err := store.db.Query(
		`SELECT document_id, name, created_at, updated_at
		FROM documents`,
	)

	if err != nil {
		return []*DocumentMetadata{}, err
	}

	defer rows.Close()
	for rows.Next() {
		var id string
		var name string
		var createdAt time.Time
		var updatedAt time.Time

		err := rows.Scan(&id, &name, &createdAt, &updatedAt)
		if err != nil {
			return []*DocumentMetadata{}, err
		}

		document := DocumentMetadata{
			DocumentID: id,
			Name:       name,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		}

		documents = append(documents, &document)
	}

	if errRows := rows.Err(); errRows != nil {
		return []*DocumentMetadata{}, errRows
	}

	if len(documents) == 0 {
		return []*DocumentMetadata{}, nil
	}

	return documents, nil
}

func (store *Store) GetDocumentByID(id string) (*DocumentMetadata, error) {

	row := store.db.QueryRow(
		`SELECT document_id, name, created_at, updated_at
		FROM documents
		WHERE document_id = ?`, id)

	var ID string
	var name string
	var createdAt time.Time
	var updatedAt time.Time

	errScan := row.Scan(&ID, &name, &createdAt, &updatedAt)
	if errScan != nil {
		return nil, errScan
	}

	document := DocumentMetadata{
		DocumentID: ID,
		Name:       name,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	return &document, nil
}

func NewStore(path string) (*Store, error) {
	cleanPath := filepath.Clean(path)
	if cleanPath == "." {
		return nil, errors.New("database path is empty")
	}

	info, errStat := os.Stat(cleanPath)
	if errStat == nil && info.IsDir() {
		return nil, fmt.Errorf("database path points to a directory: %s", cleanPath)
	}

	if errStat != nil && !errors.Is(errStat, os.ErrNotExist) {
		return nil, errStat
	}

	parent := filepath.Dir(cleanPath)
	if parent != "." {
		if errMkdir := os.MkdirAll(parent, 0o755); errMkdir != nil {
			return nil, errMkdir
		}
	}

	db, err := sql.Open("sqlite", cleanPath)
	if err != nil {
		return nil, err
	}

	errPing := db.Ping()
	if errPing != nil {
		db.Close()
		return nil, errPing
	}

	store := Store{
		db: db,
	}

	_, errCreate := store.createTable()
	if errCreate != nil {
		db.Close()
		return nil, errCreate
	}

	return &store, nil
}

func (db *Store) createTable() (sql.Result, error) {
	result, err := db.db.Exec(
		`CREATE TABLE IF NOT EXISTS documents(
		document_id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at DATE NOT NULL,
		updated_at DATE NOT NULL)`,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}
