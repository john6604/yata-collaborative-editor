package document

import (
	"errors"

	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/persistence"
)

func RebuildRelations(persistedElements map[identifier.ID]*persistence.PersistedElement, elementsBydIds map[identifier.ID]*Element) (error, map[identifier.ID]*Element) {

	for k := range elementsBydIds {
		persisted := persistedElements[k]
		if persisted == nil {
			return errors.New("No existing key."), nil
		}
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

	return nil, elementsBydIds
}

func ConstructStartEnd() (*Element, *Element) {
	start := identifier.NewID("START", -1)
	end := identifier.NewID("END", -2)

	startElement := NewElement(*start, nil, nil, nil, '\x00')
	endElement := NewElement(*end, nil, nil, nil, '\x00')

	return startElement, endElement
}
