package persistence

import "github.com/john6604/yata-collaborative-editor/internal/identifier"

type PersistedElement struct {
	ElementID identifier.ID
	OriginID  identifier.ID
	LeftID    identifier.ID
	RightID   identifier.ID
	IsDeleted bool
	Content   rune
}

type PersistedMetadata struct {
	ClientID         string
	Clock            int
	CharacterCounter int
}

type PersistedInsertOperation struct {
	NewID    identifier.ID
	OriginID identifier.ID
	RightID  identifier.ID
	Content  rune
}

type PersistedDeleteOperation struct {
	TargetID identifier.ID
}
