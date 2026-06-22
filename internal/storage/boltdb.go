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
	db, err := bolt.Open("yata.db", 0o600, &bolt.Options{Timeout: 2 * time.Second})

	if err != nil {
		return err
	}

	s.db = db

	s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("metadata"))

		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	})

	s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("elements"))

		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	})

	defer s.db.Close()

	return nil
}
