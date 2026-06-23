package document

import (
	"errors"

	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/persistence"
)

type ListConstructor interface {
	LoadMetadata() (*persistence.PersistedMetadata, error)
	LoadElements() (map[identifier.ID]*persistence.PersistedElement, map[identifier.ID]*Element, error)
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

func (d *Document) ReconstructDocument(constructor ListConstructor) error {

	// Collect Data
	metadata, err := constructor.LoadMetadata()

	if err != nil {
		return err
	}

	persistedElements, elements, err2 := constructor.LoadElements()

	if err2 != nil {
		return err2
	}

	err3, elementsByIds := RebuildRelations(persistedElements, elements)

	if err3 != nil {
		return err3
	}

	// Set data
	d.Start = elementsByIds[identifier.ID{ClientID: "START", Clock: -1}]
	d.End = elementsByIds[identifier.ID{ClientID: "END", Clock: -2}]
	d.ClientID = metadata.ClientID
	d.Clock = metadata.Clock
	d.CharacterCounter = metadata.CharacterCounter

	d.ElementsByID = elementsByIds

	if d.Start == nil || d.End == nil {
		return errors.New("Start or End are null value.")
	}

	return nil
}
