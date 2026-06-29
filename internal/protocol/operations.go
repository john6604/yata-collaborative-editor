package protocol

import "github.com/john6604/yata-collaborative-editor/internal/identifier"

type InsertOperation struct {
	NewID    identifier.ID
	OriginID identifier.ID
	RightID  identifier.ID
	Content  byte
}

type DeleteOperation struct {
	TargetID identifier.ID
}

func NewInsertOperation(newID identifier.ID, originID identifier.ID, rightID identifier.ID, content byte) *InsertOperation {
	insert := InsertOperation{NewID: newID}
	insert.OriginID = originID
	insert.RightID = rightID
	insert.Content = content
	return &insert
}

func NewDeleteOperation(targetID identifier.ID) *DeleteOperation {
	delete := DeleteOperation{TargetID: targetID}
	return &delete
}
