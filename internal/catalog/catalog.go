package catalog

import (
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

type DocumentMetadata struct {
	DocumentID string
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Store struct {
	db *sql.DB
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
			updated_at,
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

	return documents, nil
}

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
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
