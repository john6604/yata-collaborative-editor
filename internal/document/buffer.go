package document

import "github.com/john6604/yata-collaborative-editor/internal/identifier"

type PendingElement struct {
	newID    identifier.ID
	originID identifier.ID
	rightID  identifier.ID
	content  byte
}

func NewPending(id identifier.ID, origin identifier.ID, right identifier.ID, content byte) *PendingElement {
	pending := PendingElement{newID: id}
	pending.originID = origin
	pending.rightID = right
	pending.content = content
	return &pending
}
