package persistence

import "github.com/john6604/yata-collaborative-editor/internal/identifier"

type PersistedElement struct {
	ElementID identifier.ID
	OriginID  identifier.ID
	LeftID    identifier.ID
	RightID   identifier.ID
	IsDeleted bool
	Content   byte
}

type PersistedMetadata struct {
	ClientID         string
	Clock            int
	CharacterCounter int
}
