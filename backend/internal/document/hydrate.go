package document

import (
	"errors"

	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/persistence"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

type DocumentConstructor interface {
	LoadSnapshot() (*persistence.PersistedMetadata, map[identifier.ID]*persistence.PersistedElement, map[identifier.ID]*Element, map[identifier.ID]*protocol.InsertOperation, map[identifier.ID]*protocol.DeleteOperation, error)
}

func RebuildRelations(persistedElements map[identifier.ID]*persistence.PersistedElement, elementsBydIds map[identifier.ID]*Element) (error, map[identifier.ID]*Element) {

	for k := range elementsBydIds {
		persisted := persistedElements[k]
		if persisted == nil {
			return errors.New("No existing key."), nil
		}
		if k.ClientID == "START" || k.ClientID == "END" {
			elementsBydIds[k].Origin = nil
			elementsBydIds[k].Left = elementsBydIds[identifier.ID(persisted.LeftID)]
			elementsBydIds[k].Right = elementsBydIds[identifier.ID(persisted.RightID)]
		} else {
			if elementsBydIds[identifier.ID(persisted.OriginID)] == nil {
				return errors.New("No existing key."), nil
			}
			elementsBydIds[k].Origin = elementsBydIds[identifier.ID(persisted.OriginID)]
			if elementsBydIds[identifier.ID(persisted.LeftID)] == nil {
				return errors.New("No existing key."), nil
			}
			elementsBydIds[k].Left = elementsBydIds[identifier.ID(persisted.LeftID)]
			if elementsBydIds[identifier.ID(persisted.RightID)] == nil {
				return errors.New("No existing key."), nil
			}
			elementsBydIds[k].Right = elementsBydIds[identifier.ID(persisted.RightID)]
		}
	}

	return nil, elementsBydIds
}

func (d *Document) ReconstructDocument(constructor DocumentConstructor) error {

	// Collect Data
	metadata, persistedElements, elements, insertLog, deleteLog, err := constructor.LoadSnapshot()

	if err != nil {
		return err
	}

	err1, elementsByIds := RebuildRelations(persistedElements, elements)

	if err1 != nil {
		return err1
	}

	// Set data
	d.Start = elementsByIds[identifier.ID{ClientID: "START", Clock: -1}]
	d.End = elementsByIds[identifier.ID{ClientID: "END", Clock: -2}]
	d.ClientID = metadata.ClientID
	d.Clock = metadata.Clock
	d.CharacterCounter = metadata.CharacterCounter

	d.ElementsByID = elementsByIds

	d.InsertLog = insertLog
	d.DeleteLog = deleteLog

	d.ReconstructPendingInserts()
	d.ReconstructPendingDeletes()

	d.processPending()
	d.processPendingDeletes()

	if d.Start == nil || d.End == nil {
		return errors.New("Start or End are null value.")
	}

	return nil
}

func (d *Document) ReconstructPendingInserts() {

	inserts := d.InsertLog

	d.PendingInserts = make(map[identifier.ID]*PendingElement)

	for k, v := range inserts {

		_, exists := d.ElementsByID[k]

		if !exists {
			d.PendingInserts[k] = NewPending(k, v.OriginID, v.RightID, v.Content)
		}
	}
}

func (d *Document) ReconstructPendingDeletes() {

	deletes := d.DeleteLog

	d.PendingDeletes = make(map[identifier.ID]*PendingElement)

	for k := range deletes {

		_, exists := d.ElementsByID[k]

		if !exists {
			d.PendingDeletes[k] = NewPending(k, identifier.ID{}, identifier.ID{}, '\x00')
		}
	}
}
