package storage

import (
	"github.com/john6604/yata-collaborative-editor/internal/document"
)

func ToPersistedElement(element *document.Element) PersistedElement {
	persistedElement := PersistedElement{ElementID: element.ElementID}
	persistedElement.Content = element.Content
	persistedElement.IsDeleted = element.IsDeleted
	if element.Origin != nil {
		persistedElement.OriginID = element.Origin.ElementID
	} else {
		persistedElement.OriginID = document.ID{}
	}

	if element.Left != nil {
		persistedElement.LeftID = element.Left.ElementID
	} else {
		persistedElement.LeftID = document.ID{}
	}

	if element.Right != nil {
		persistedElement.RightID = element.Right.ElementID
	} else {
		persistedElement.RightID = document.ID{}
	}

	return persistedElement
}

func ToElement(persistedElement PersistedElement) document.Element {
	element := document.NewElement(persistedElement.ElementID, nil, nil, nil, persistedElement.Content)
	element.IsDeleted = persistedElement.IsDeleted

	return *element
}

func ToPersistedMetadata(document *document.Document) PersistedMetadata {
	persistedMetadata := PersistedMetadata{ClientID: document.ClientID}
	persistedMetadata.Clock = document.Clock
	persistedMetadata.CharacterCounter = document.CharacterCounter

	return persistedMetadata
}

func ToMetadata(persistedMetadata PersistedMetadata) document.Document {
	document := document.NewDocument()
	document.ClientID = persistedMetadata.ClientID
	document.CharacterCounter = persistedMetadata.CharacterCounter
	document.Clock = persistedMetadata.Clock

	return *document
}
