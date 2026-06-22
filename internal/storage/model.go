package storage

import (
	"github.com/john6604/yata-collaborative-editor/internal/document"
)

type PersistedElement struct {
	ElementID document.ID
	OriginID  document.ID
	LeftID    document.ID
	RightID   document.ID
	IsDeleted bool
	Content   byte
}

type PersistedMetadata struct {
	ClientID         document.ID
	Clock            int
	CharacterCounter int
}
