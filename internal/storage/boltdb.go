package storage

import (
	"encoding/json"
	"errors"
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

func formatID(clientID document.ID) string {
	return "(" + clientID.ClientID + "," + fmt.Sprint(clientID.Clock) + ")"
}

func (s *Storage) SaveElements(document document.Document) error {

	current := document.Start.Right

	for current != document.End {

		id := formatID(current.ElementID)
		elementID := []byte(id)

		persistedElement := ToPersistedElement(current)

		data, err := json.Marshal(persistedElement)

		if err != nil {
			return err
		}

		errTransaction := s.db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("elements"))

			if bucket == nil {
				return fmt.Errorf("No bucket assigned.")
			}

			err := bucket.Put(elementID, data)

			if err != nil {
				return err
			}

			return nil
		})

		if errTransaction != nil {
			return errTransaction
		}

		current = current.Right
	}

	return nil
}

func (s *Storage) LoadElements() (map[document.ID]*PersistedElement, map[document.ID]*document.Element, error) {

	elementByIDs := make(map[document.ID]*document.Element)
	elementsPersisted := make(map[document.ID]*PersistedElement)

	errTransaction := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("elements"))

		if bucket == nil {
			return fmt.Errorf("No bucket assigned.")
		}

		err := bucket.ForEach(func(k, v []byte) error {

			var persistedElement PersistedElement

			err := json.Unmarshal(v, &persistedElement)

			if err != nil {
				return err
			}

			elementsPersisted[persistedElement.ElementID] = &persistedElement

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

func RebuildRelations(persistedElements map[document.ID]*PersistedElement, elementsBydIds map[document.ID]*document.Element) (error, map[document.ID]*document.Element) {

	for k := range elementsBydIds {
		persisted := persistedElements[k]
		if persisted == nil {
			return errors.New("No existing key."), nil
		}
		if elementsBydIds[persisted.OriginID] == nil {
			return errors.New("No existing key."), nil
		}
		elementsBydIds[k].Origin = elementsBydIds[persisted.OriginID]
		if elementsBydIds[persisted.LeftID] == nil {
			return errors.New("No existing key."), nil
		}
		elementsBydIds[k].Left = elementsBydIds[persisted.LeftID]
		if elementsBydIds[persisted.RightID] == nil {
			return errors.New("No existing key."), nil
		}
		elementsBydIds[k].Right = elementsBydIds[persisted.RightID]
	}

	return nil, elementsBydIds
}

func ConstructStartEnd() (*document.Element, *document.Element) {
	start := document.NewID("Start", -1)
	end := document.NewID("End", -2)

	startElement := document.NewElement(*start, nil, nil, nil, '\x00')
	endElement := document.NewElement(*end, nil, nil, nil, '\x00')

	return startElement, endElement
}
