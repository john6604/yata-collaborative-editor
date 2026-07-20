package relay

import (
	"errors"
	"strings"
	"sync"
)

type Hub struct {
	mutex sync.RWMutex // Possible change: this one only allows one read and write operation
	rooms map[string]*Room
}

func NewHub() *Hub {
	hub := Hub{rooms: make(map[string]*Room)}
	return &hub
}

func (h *Hub) Join(roomID string, clientID string) (*ClientSession, error) {

	newRoomID := strings.TrimSpace(roomID)

	if len(newRoomID) == 0 {
		return nil, errors.New("Room ID is empty.")
	}

	newClientID := strings.TrimSpace(clientID)

	if len(newClientID) == 0 {
		return nil, errors.New("Client ID is empty.")
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	room, roomExists := h.rooms[newRoomID]

	if !roomExists {
		room = NewRoom(newRoomID)
		h.rooms[room.id] = room
	}

	_, clienExists := room.clients[newClientID]

	if clienExists {
		return nil, errors.New("Client already exists.")
	}

	clientSession := NewClientSession(newClientID, newRoomID)

	room.clients[newClientID] = clientSession

	return clientSession, nil
}

func (h *Hub) Leave(clientSession *ClientSession) {

	if clientSession == nil {
		return
	}

	if clientSession.clientID == "" || clientSession.roomID == "" {
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	room, roomExists := h.rooms[clientSession.roomID]

	if !roomExists {
		return
	}

	session, sessionExists := room.clients[clientSession.clientID]

	if !sessionExists {
		return
	}

	if session != clientSession {
		return
	}

	delete(room.clients, session.clientID)

	if len(room.clients) == 0 {
		delete(h.rooms, room.id)
	}

}

func (h *Hub) CountRooms() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.rooms)
}

func (h *Hub) CountClients(roomID string) (int, error) {

	formattedRoomID := strings.TrimSpace(roomID)

	if len(formattedRoomID) == 0 {
		return 0, errors.New("Room ID does not exist.")
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	room, exists := h.rooms[formattedRoomID]

	if !exists {
		return 0, errors.New("Room does not exist.")
	}

	return len(room.clients), nil
}

func (h *Hub) HasClient(roomID string, clientID string) (bool, error) {

	formattedRoomID := strings.TrimSpace(roomID)

	if len(formattedRoomID) == 0 {
		return false, errors.New("Room ID is not valid.")
	}

	formattedClientID := strings.TrimSpace(clientID)

	if len(formattedClientID) == 0 {
		return false, errors.New("Client ID is not valid.")
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	room, roomExists := h.rooms[formattedRoomID]

	if !roomExists {
		return false, errors.New("Room ID does not exist.")
	}

	_, clientExists := room.clients[formattedClientID]

	return clientExists, nil
}
