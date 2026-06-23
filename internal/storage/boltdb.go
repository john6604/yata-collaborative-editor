package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/john6604/yata-collaborative-editor/internal/document"
	bolt "go.etcd.io/bbolt"
)

type Storage struct {
	db *bolt.DB
}

func (s *Storage) OpenDB() error {
	db, err := bolt.Open("..\\..\\internal\\storage\\yata.db", 0o600, &bolt.Options{Timeout: 2 * time.Second})

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

func (s *Storage) LoadMetadata() (*PersistedMetadata, error) {

	var persistedMetadata PersistedMetadata

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
