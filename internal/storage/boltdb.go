package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/persistence"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

var ErrSnapshotNotFound error = errors.New("Snapshot not found")

type Storage struct {
	db *bolt.DB
}

type RoomStorage struct {
	MainStorage *Storage
	RoomID      string
}

func NewRoomStorage(storage *Storage, room string) (*RoomStorage, error) {

	formattedRoom := strings.TrimSpace(room)
	if formattedRoom == "" {
		return &RoomStorage{}, errors.New("Invalid room")
	}

	if storage == nil {
		return &RoomStorage{}, errors.New("Invalid storage")
	}

	roomStorage := RoomStorage{
		MainStorage: storage,
		RoomID:      formattedRoom,
	}

	return &roomStorage, nil
}

func (s *Storage) OpenDB(path string) error {
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 2 * time.Second})

	if err != nil {
		return err
	}

	s.db = db

	errBuckets := s.db.Update(func(tx *bolt.Tx) error {

		_, errRoot := tx.CreateBucketIfNotExists([]byte("rooms"))
		if errRoot != nil {
			return errRoot
		}

		return nil
	})

	if errBuckets != nil {
		return errBuckets
	}

	return nil
}

func (s *Storage) createRoom(bucket *bolt.Bucket, room string) (*bolt.Bucket, error) {

	formattedRoom := strings.TrimSpace(room)

	if formattedRoom == "" {
		return nil, errors.New("empty room")
	}

	subRoom, errRoom := bucket.CreateBucketIfNotExists([]byte(formattedRoom))

	if errRoom != nil {
		return nil, errRoom
	}

	_, err1 := subRoom.CreateBucketIfNotExists([]byte("metadata"))

	if err1 != nil {
		return nil, fmt.Errorf("create bucket: %s", err1)
	}

	_, err2 := subRoom.CreateBucketIfNotExists([]byte("elements"))
	if err2 != nil {
		return nil, fmt.Errorf("create bucket: %s", err2)
	}

	_, err3 := subRoom.CreateBucketIfNotExists([]byte("insert_log"))
	if err3 != nil {
		return nil, fmt.Errorf("create bucket: %s", err3)
	}

	_, err4 := subRoom.CreateBucketIfNotExists([]byte("delete_log"))
	if err4 != nil {
		return nil, fmt.Errorf("create bucket: %s", err4)
	}

	return subRoom, nil
}

func (s *Storage) CloseDB() error {
	if err := s.db.Close(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) SaveSnapshot(document *document.Document, room string) error {

	errTransaction := s.db.Update(func(tx *bolt.Tx) error {

		root := tx.Bucket([]byte("rooms"))

		if root == nil {
			return errors.New("No bucket found")
		}

		roomBucket, errRoom := s.createRoom(root, room)

		if errRoom != nil {
			return errRoom
		}

		err1 := s.SaveMetadata(roomBucket, document)

		if err1 != nil {
			return err1
		}

		err2 := s.SaveElements(roomBucket, *document)

		if err2 != nil {
			return err2
		}

		err3 := s.SaveInsertOperations(roomBucket, *document)

		if err3 != nil {
			return err3
		}

		err4 := s.SaveDeleteOperations(roomBucket, *document)

		if err4 != nil {
			return err4
		}

		return nil
	})

	if errTransaction != nil {
		return errTransaction
	}

	return nil
}

func (rs *RoomStorage) LoadSnapshot() (*persistence.PersistedMetadata, map[identifier.ID]*persistence.PersistedElement, map[identifier.ID]*document.Element, map[identifier.ID]*protocol.InsertOperation, map[identifier.ID]*protocol.DeleteOperation, error) {
	return rs.MainStorage.LoadSnapshot(rs.RoomID)
}

func (s *Storage) LoadSnapshot(room string) (*persistence.PersistedMetadata, map[identifier.ID]*persistence.PersistedElement, map[identifier.ID]*document.Element, map[identifier.ID]*protocol.InsertOperation, map[identifier.ID]*protocol.DeleteOperation, error) {

	var metadata *persistence.PersistedMetadata
	var err1 error

	var elementsPersisted map[identifier.ID]*persistence.PersistedElement
	var elementsByID map[identifier.ID]*document.Element
	var err2 error

	var insertLog map[identifier.ID]*protocol.InsertOperation
	var err3 error
	var deleteLog map[identifier.ID]*protocol.DeleteOperation
	var err4 error

	formattedRoom := strings.TrimSpace(room)

	errTransaction := s.db.View(func(tx *bolt.Tx) error {

		root := tx.Bucket([]byte("rooms"))
		if root == nil {
			return errors.New("No bucket assigned.")
		}

		roomBucket := root.Bucket([]byte(formattedRoom))
		if roomBucket == nil {
			return ErrSnapshotNotFound
		}

		metadata, err1 = s.LoadMetadata(roomBucket)

		if err1 != nil {
			return err1
		}

		elementsPersisted, elementsByID, err2 = s.LoadElements(roomBucket)

		if err2 != nil {
			return err2
		}

		insertLog, err3 = s.LoadInsertOperations(roomBucket)

		if err3 != nil {
			return err3
		}

		deleteLog, err4 = s.LoadDeleteOperations(roomBucket)

		if err4 != nil {
			return err4
		}

		return nil
	})

	if errTransaction != nil {
		return nil, nil, nil, nil, nil, errTransaction
	}

	return metadata, elementsPersisted, elementsByID, insertLog, deleteLog, nil
}

func (s *Storage) LoadInternalSnapshot(room string) (protocol.Delta, error) {

	var delta protocol.Delta
	var err error

	formattedRoom := strings.TrimSpace(room)
	if formattedRoom == "" {
		return protocol.Delta{}, errors.New("No room found")
	}

	errView := s.db.View(func(tx *bolt.Tx) error {

		root := tx.Bucket([]byte("rooms"))
		if root == nil {
			return errors.New("No bucket found")
		}

		snapshotRoom := root.Bucket([]byte(formattedRoom))
		if snapshotRoom == nil {
			return ErrSnapshotNotFound
		}

		delta, err = s.LoadSnapshotRoom(snapshotRoom)
		if err != nil {
			return err
		}

		return nil
	})

	if errView != nil {
		return protocol.Delta{}, errView
	}

	return delta, nil
}

func (s *Storage) SaveMetadata(bucketRoom *bolt.Bucket, document *document.Document) error {

	persistedMetadata := ToPersistedMetadata(document)

	data, err := json.Marshal(persistedMetadata)

	if err != nil {
		return err
	}

	if bucketRoom == nil {
		return fmt.Errorf("No bucket asigned.")
	}

	metadata := bucketRoom.Bucket([]byte("metadata"))

	if metadata == nil {
		return fmt.Errorf("No bucket asigned.")
	}

	err1 := metadata.Put([]byte("document"), data)

	if err1 != nil {
		return fmt.Errorf("Failed to assign data.")
	}

	return nil
}

func (s *Storage) LoadMetadata(roomBucket *bolt.Bucket) (*persistence.PersistedMetadata, error) {

	var persistedMetadata persistence.PersistedMetadata

	if roomBucket == nil {
		return nil, ErrSnapshotNotFound
	}

	metadata := roomBucket.Bucket([]byte("metadata"))

	if metadata == nil {
		return nil, ErrSnapshotNotFound
	}

	data := metadata.Get([]byte("document"))

	if data == nil {
		return nil, ErrSnapshotNotFound
	}

	err := json.Unmarshal(data, &persistedMetadata)

	if err != nil {
		return nil, err
	}

	return &persistedMetadata, nil
}

func formatID(clientID identifier.ID) string {
	return "(" + clientID.ClientID + "," + fmt.Sprint(clientID.Clock) + ")"
}

func (s *Storage) SaveElements(bucketRoom *bolt.Bucket, document document.Document) error {

	if bucketRoom == nil {
		return fmt.Errorf("No bucket assigned.")
	}

	elements := bucketRoom.Bucket([]byte("elements"))

	if elements == nil {
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

		err1 := elements.Put(elementID, data)

		if err1 != nil {
			return err1
		}

		current = current.Right
	}

	return nil
}

func (s *Storage) LoadElements(roomBucket *bolt.Bucket) (map[identifier.ID]*persistence.PersistedElement, map[identifier.ID]*document.Element, error) {

	elementByIDs := make(map[identifier.ID]*document.Element)
	elementsPersisted := make(map[identifier.ID]*persistence.PersistedElement)

	if roomBucket == nil {
		return nil, nil, fmt.Errorf("No bucket assigned.")
	}

	elements := roomBucket.Bucket([]byte("elements"))

	if elements == nil {
		return nil, nil, fmt.Errorf("No bucket found.")
	}

	err := elements.ForEach(func(k, v []byte) error {

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
		return nil, nil, err
	}

	return elementsPersisted, elementByIDs, nil
}

func (s *Storage) SaveInsertOperations(bucketRoom *bolt.Bucket, document document.Document) error {

	if bucketRoom == nil {
		return fmt.Errorf("No bucket assigned.")
	}

	insertLog := bucketRoom.Bucket([]byte("insert_log"))

	if insertLog == nil {
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

		err1 := insertLog.Put(newID, data)

		if err1 != nil {
			return err1
		}
	}

	return nil
}

func (s *Storage) LoadInsertOperations(roomBucket *bolt.Bucket) (map[identifier.ID]*protocol.InsertOperation, error) {

	insertLog := make(map[identifier.ID]*protocol.InsertOperation)

	if roomBucket == nil {
		return nil, fmt.Errorf("No bucket assigned.")
	}

	insertBucket := roomBucket.Bucket([]byte("insert_log"))

	if insertBucket == nil {
		return nil, fmt.Errorf("No bucket found.")
	}

	err := insertBucket.ForEach(func(k, v []byte) error {

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
		return nil, err
	}

	return insertLog, nil
}

func (s *Storage) SaveDeleteOperations(bucketRoom *bolt.Bucket, document document.Document) error {

	if bucketRoom == nil {
		return fmt.Errorf("No bucket assigned.")
	}

	deleteLog := bucketRoom.Bucket([]byte("delete_log"))

	if deleteLog == nil {
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

		err1 := deleteLog.Put(targetID, data)

		if err1 != nil {
			return err1
		}
	}

	return nil
}

func (s *Storage) LoadDeleteOperations(roomBucket *bolt.Bucket) (map[identifier.ID]*protocol.DeleteOperation, error) {

	deleteLog := make(map[identifier.ID]*protocol.DeleteOperation)

	if roomBucket == nil {
		return nil, fmt.Errorf("No bucket assigned.")
	}

	deleteBucket := roomBucket.Bucket([]byte("delete_log"))
	if deleteBucket == nil {
		return nil, fmt.Errorf("No bucket found.")
	}

	err := deleteBucket.ForEach(func(k, v []byte) error {

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
		return nil, err
	}

	return deleteLog, nil
}

func (s *Storage) SaveSnapshotRoom(room string, delta protocol.Delta) error {

	formattedRoom := strings.TrimSpace(room)
	if formattedRoom == "" {
		return errors.New("empty room")
	}

	data, err := json.Marshal(delta)
	if err != nil {
		return err
	}

	errSaving := s.db.Update(func(tx *bolt.Tx) error {

		root := tx.Bucket([]byte("rooms"))
		if root == nil {
			return errors.New("no bucket found")
		}

		savingRoom, errSavingRoom := s.createRoom(root, formattedRoom)
		if errSavingRoom != nil {
			return errSavingRoom
		}

		errSnapshot := savingRoom.Put([]byte("snapshot"), data)
		if errSnapshot != nil {
			return errSnapshot
		}

		return nil
	})

	if errSaving != nil {
		return errSaving
	}

	return nil
}

func (s *Storage) LoadSnapshotRoom(roomBucket *bolt.Bucket) (protocol.Delta, error) {

	var delta protocol.Delta

	if roomBucket == nil {
		return protocol.Delta{}, errors.New("No bucket found")
	}

	data := roomBucket.Get([]byte("snapshot"))
	if data == nil {
		return protocol.Delta{}, ErrSnapshotNotFound
	}

	err := json.Unmarshal(data, &delta)
	if err != nil {
		return protocol.Delta{}, err
	}

	return delta, nil
}
