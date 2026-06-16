package document

type ID struct {
	ClientID string
	Clock    int
}

func NewID(clientID string, clock int) *ID {
	id := ID{ClientID: clientID}
	id.Clock = clock
	return &id
}
