package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/persistence"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

type Storage struct {
	db *bolt.DB
}

func (s *Storage) OpenDB(path string) error {
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 2 * time.Second})

	if err != nil {
		return err
	}

	s.db = db

	errBuckets := s.db.Update(func(tx *bolt.Tx) error {
		_, err1 := tx.CreateBucketIfNotExists([]byte("metadata"))

		if err1 != nil {
			return fmt.Errorf("create bucket: %s", err1)
		}

		_, err2 := tx.CreateBucketIfNotExists([]byte("elements"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}

		_, err3 := tx.CreateBucketIfNotExists([]byte("insert_log"))
		if err3 != nil {
			return fmt.Errorf("create bucket: %s", err3)
		}

		_, err4 := tx.CreateBucketIfNotExists([]byte("delete_log"))
		if err4 != nil {
			return fmt.Errorf("create bucket: %s", err4)
		}

		return nil
	})

	if errBuckets != nil {
		return errBuckets
	}

	return nil
}

func (s *Storage) CloseDB() error {
	if err := s.db.Close(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) SaveMetadata(document *document.Document) error {

	persistedMetadata := ToPersistedMetadata(document)

	data, err := json.Marshal(persistedMetadata)

	if err != nil {
		return err
	}

	errTransaction := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("metadata"))

		if bucket == nil {
			return fmt.Errorf("No bucket asigned.")
		}

		err := bucket.Put([]byte("document"), data)

		if err != nil {
			return fmt.Errorf("Failed to assign data.")
		}

		return nil
	})

	if errTransaction != nil {
		return errTransaction
	}

	return nil
}

func (s *Storage) LoadMetadata() (*persistence.PersistedMetadata, error) {

	var persistedMetadata persistence.PersistedMetadata

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("metadata"))

		if bucket == nil {
			return fmt.Errorf("No bucket found.")
		}

		data := bucket.Get([]byte("document"))

		if data == nil {
			return fmt.Errorf("No data associated with the key.")
		}

		err := json.Unmarshal(data, &persistedMetadata)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &persistedMetadata, nil
}

func formatID(clientID identifier.ID) string {
	return "(" + clientID.ClientID + "," + fmt.Sprint(clientID.Clock) + ")"
}

func (s *Storage) SaveElements(document document.Document) error {

	errTransaction := s.db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte("elements"))

		if bucket == nil {
			return fmt.Errorf("No bucket assigned.")
		}

		current := document.Start

		for current != nil {

			id := formatID(current.ElementID)
			elementID := []byte(id)

			persistedElement := ToPersistedElement(current)

			data, err := json.Marshal(persistedElement)

			if err != nil {
				return err
			}

			err1 := bucket.Put(elementID, data)

			if err1 != nil {
				return err1
			}

			current = current.Right
		}

		return nil
	})

	if errTransaction != nil {
		return errTransaction
	}

	return nil
}

func (s *Storage) LoadElements() (map[identifier.ID]*persistence.PersistedElement, map[identifier.ID]*document.Element, error) {

	elementByIDs := make(map[identifier.ID]*document.Element)
	elementsPersisted := make(map[identifier.ID]*persistence.PersistedElement)

	errTransaction := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("elements"))

		if bucket == nil {
			return fmt.Errorf("No bucket assigned.")
		}

		err := bucket.ForEach(func(k, v []byte) error {

			var persistedElement persistence.PersistedElement

			err := json.Unmarshal(v, &persistedElement)

			if err != nil {
				return err
			}

			elementsPersisted[identifier.ID(persistedElement.ElementID)] = &persistedElement

			element := ToElement(persistedElement)

			elementByIDs[element.ElementID] = &element

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})

	if errTransaction != nil {
		return nil, nil, errTransaction
	}

	return elementsPersisted, elementByIDs, nil

}

func (s *Storage) SaveInsertOperations(document document.Document) error {

	errTransaction := s.db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte("insert_log"))

		if bucket == nil {
			return fmt.Errorf("No bucket assigned.")
		}

		inserts := document.InsertLog

		for k, v := range inserts {

			id := formatID(k)
			newID := []byte(id)

			persistedInsert := ToPersistedInsertOperation(v)

			data, err := json.Marshal(persistedInsert)

			if err != nil {
				return err
			}

			err1 := bucket.Put(newID, data)

			if err1 != nil {
				return err1
			}
		}

		return nil
	})

	if errTransaction != nil {
		return errTransaction
	}

	return nil
}

func (s *Storage) LoadInsertOperations() (map[identifier.ID]*protocol.InsertOperation, error) {

	insertLog := make(map[identifier.ID]*protocol.InsertOperation)

	errTransaction := s.db.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte("insert_log"))

		if bucket == nil {
			return fmt.Errorf("No bucket assigned.")
		}

		err := bucket.ForEach(func(k, v []byte) error {

			var persistedInsert persistence.PersistedInsertOperation

			err := json.Unmarshal(v, &persistedInsert)

			if err != nil {
				return err
			}

			insert := ToInsertOperation(persistedInsert)

			insertLog[insert.NewID] = &insert

			return nil

		})

		if err != nil {
			return err
		}

		return nil
	})

	if errTransaction != nil {
		return nil, errTransaction
	}

	return insertLog, nil
}

func (s *Storage) SaveDeleteOperations(document document.Document) error {

	errTransaction := s.db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte("delete_log"))

		if bucket == nil {
			return fmt.Errorf("No bucket assigned.")
		}

		deletes := document.DeleteLog

		for k, v := range deletes {

			id := formatID(k)
			targetID := []byte(id)

			persistedDelete := ToPersistedDeleteOperation(v)

			data, err := json.Marshal(persistedDelete)

			if err != nil {
				return err
			}

			err1 := bucket.Put(targetID, data)

			if err1 != nil {
				return err1
			}
		}

		return nil
	})

	if errTransaction != nil {
		return errTransaction
	}

	return nil
}

func (s *Storage) LoadDeleteOperations() (map[identifier.ID]*protocol.DeleteOperation, error) {

	deleteLog := make(map[identifier.ID]*protocol.DeleteOperation)

	errTransaction := s.db.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte("delete_log"))

		if bucket == nil {
			return fmt.Errorf("No bucket assigned.")
		}

		err := bucket.ForEach(func(k, v []byte) error {

			var persistedDelete persistence.PersistedDeleteOperation

			err := json.Unmarshal(v, &persistedDelete)

			if err != nil {
				return err
			}

			delete := ToDeleteOperation(persistedDelete)

			deleteLog[delete.TargetID] = &delete

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})

	if errTransaction != nil {
		return nil, errTransaction
	}

	return deleteLog, nil
}
