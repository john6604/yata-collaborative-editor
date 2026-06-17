package document

// Definition of structur for internal ID
type ID struct {
	ClientID string
	Clock    int
}

// Constructor function
func NewID(clientID string, clock int) *ID {
	id := ID{ClientID: clientID}
	id.Clock = clock
	return &id
}
