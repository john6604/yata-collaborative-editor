package protocol

type Delta struct {
	Inserts []*InsertOperation
	Deletes []*DeleteOperation
}
