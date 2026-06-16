package document

type Element struct {
	ElementID ID
	Origin    *Element
	Left      *Element
	Right     *Element
	IsDeleted bool
	Content   byte
}

func NewElement(id ID, origin *Element, left *Element, right *Element, content byte) *Element {
	element := Element{ElementID: id}
	element.Origin = origin
	element.Left = left
	element.Right = right
	element.IsDeleted = false
	element.Content = content
	return &element
}
