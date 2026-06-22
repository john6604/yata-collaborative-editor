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
