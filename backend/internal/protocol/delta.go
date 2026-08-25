package protocol

type Delta struct {
	Inserts []*InsertOperation `json:"inserts"`
	Deletes []*DeleteOperation `json:"deletes"`
}
