package document

type PendingElement struct {
	newID    ID
	originID ID
	rightID  ID
	content  byte
}

func NewPending(id ID, origin ID, right ID, content byte) *PendingElement {
	pending := PendingElement{newID: id}
	pending.originID = origin
	pending.rightID = right
	pending.content = content
	return &pending
}
