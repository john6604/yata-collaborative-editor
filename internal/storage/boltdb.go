package storage

import (
	"fmt"
	"time"

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
