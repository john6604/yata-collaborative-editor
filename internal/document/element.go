package document

import "github.com/john6604/yata-collaborative-editor/internal/identifier"

// Definition of structure for Element
type Element struct {
	ElementID identifier.ID
	Origin    *Element
	Left      *Element
	Right     *Element
	IsDeleted bool
	Content   byte
}

// Constructor function
func NewElement(id identifier.ID, origin *Element, left *Element, right *Element, content byte) *Element {
	element := Element{ElementID: id}
	element.Origin = origin
	element.Left = left
	element.Right = right
	element.IsDeleted = false
	element.Content = content
	return &element
}
