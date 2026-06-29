package sync

import "github.com/john6604/yata-collaborative-editor/internal/document"

type Delta struct {
	Inserts []*document.InsertOperation
	Deletes []*document.DeleteOperation
}
