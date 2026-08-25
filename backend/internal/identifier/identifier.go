package identifier

// Definition of structur for internal ID
type ID struct {
	ClientID string `json:"client_id"`
	Clock    int    `json:"clock"`
}

// Constructor function
func NewID(clientID string, clock int) *ID {
	id := ID{ClientID: clientID}
	id.Clock = clock
	return &id
}
