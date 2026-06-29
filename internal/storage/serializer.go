package storage

import (
	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/persistence"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

func ToPersistedElement(element *document.Element) persistence.PersistedElement {
	persistedElement := persistence.PersistedElement{ElementID: identifier.ID(element.ElementID)}
	persistedElement.Content = element.Content
	persistedElement.IsDeleted = element.IsDeleted
	if element.Origin != nil {
		persistedElement.OriginID = identifier.ID(element.Origin.ElementID)
	} else {
		persistedElement.OriginID = identifier.ID{}
	}

	if element.Left != nil {
		persistedElement.LeftID = identifier.ID(element.Left.ElementID)
	} else {
		persistedElement.LeftID = identifier.ID{}
	}

	if element.Right != nil {
		persistedElement.RightID = identifier.ID(element.Right.ElementID)
	} else {
		persistedElement.RightID = identifier.ID{}
	}

	return persistedElement
}

func ToElement(persistedElement persistence.PersistedElement) document.Element {
	element := document.NewElement(identifier.ID(persistedElement.ElementID), nil, nil, nil, persistedElement.Content)
	element.IsDeleted = persistedElement.IsDeleted

	return *element
}

func ToPersistedMetadata(document *document.Document) persistence.PersistedMetadata {
	persistedMetadata := persistence.PersistedMetadata{ClientID: document.ClientID}
	persistedMetadata.Clock = document.Clock
	persistedMetadata.CharacterCounter = document.CharacterCounter

	return persistedMetadata
}

func ToMetadata(persistedMetadata persistence.PersistedMetadata) document.Document {
	document := document.NewDocument()
	document.ClientID = persistedMetadata.ClientID
	document.CharacterCounter = persistedMetadata.CharacterCounter
	document.Clock = persistedMetadata.Clock

	return *document
}

func ToPersistedInsertOperation(insertOperation *protocol.InsertOperation) persistence.PersistedInsertOperation {
	persistedInsertOperation := persistence.PersistedInsertOperation{NewID: insertOperation.NewID}
	persistedInsertOperation.OriginID = insertOperation.OriginID
	persistedInsertOperation.RightID = insertOperation.RightID
	persistedInsertOperation.Content = insertOperation.Content
	return persistedInsertOperation
}

func ToInsertOperation(persistedInsert persistence.PersistedInsertOperation) protocol.InsertOperation {
	insert := protocol.NewInsertOperation(persistedInsert.NewID, persistedInsert.OriginID, persistedInsert.RightID, persistedInsert.Content)
	return *insert
}

func ToPersistedDeleteOperation(deleteOperation *protocol.DeleteOperation) persistence.PersistedDeleteOperation {
	persistedDeleteOperation := persistence.PersistedDeleteOperation{TargetID: deleteOperation.TargetID}
	return persistedDeleteOperation
}

func ToDeleteOperation(persistedDelete persistence.PersistedDeleteOperation) protocol.DeleteOperation {
	delete := protocol.NewDeleteOperation(persistedDelete.TargetID)
	return *delete
}
