package relay

import (
	"errors"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var ErrNotJoined = errors.New("not_joined")

type Hub struct {
	mutex sync.RWMutex // Possible change: this one only allows one read and write operation
	rooms map[string]*Room
}

func NewHub() *Hub {
	hub := Hub{rooms: make(map[string]*Room)}
	return &hub
}

func (h *Hub) Join(roomID string, clientID string, conn *websocket.Conn) (*ClientSession, error) {

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

	clientSession := NewClientSession(newClientID, newRoomID, conn)

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

func (h *Hub) GetActiveUsers(roomID string) ([]string, error) {

	formattedRoomID := strings.TrimSpace(roomID)
	if len(formattedRoomID) == 0 {
		return nil, errors.New("Room ID does not exist.")
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	room, exists := h.rooms[formattedRoomID]

	if !exists {
		return nil, errors.New("Room does not exist.")
	}

	var users []string

	for _, client := range room.clients {
		users = append(users, client.clientID)
	}

	if len(users) == 0 {
		return []string{}, nil
	}

	return users, nil
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

func (h *Hub) BroadcastToRoom(senderSession *ClientSession, message []byte) (bool, error) {

	if senderSession == nil {
		return false, ErrNotJoined
	}

	h.mutex.RLock()

	room, roomExists := h.rooms[senderSession.roomID]

	if !roomExists {
		h.mutex.RUnlock()
		return false, ErrNotJoined
	}

	storedSession, sessionExists := room.clients[senderSession.clientID]

	if !sessionExists {
		h.mutex.RUnlock()
		return false, ErrNotJoined
	}

	if storedSession != senderSession {
		h.mutex.RUnlock()
		return false, ErrNotJoined
	}

	var receivers []*ClientSession

	for _, v := range room.clients {
		if v != senderSession && v.webSocket != nil {
			receivers = append(receivers, v)
		}
	}

	h.mutex.RUnlock()

	if len(receivers) == 0 {
		return false, nil
	}

	for _, client := range receivers {

		errSend := client.Send(message)

		if errSend != nil {
			return false, errSend
		}
	}

	return true, nil
}

func (h *Hub) BroadcastToAllInRoom(roomID string, message []byte) error {

	formattedRoomID := strings.TrimSpace(roomID)
	if len(formattedRoomID) == 0 {
		return errors.New("Room ID does not exist.")
	}

	h.mutex.RLock()

	room, exists := h.rooms[formattedRoomID]
	if !exists {
		h.mutex.RUnlock()
		return errors.New("Room does not exist.")
	}

	var clients []*ClientSession

	for _, client := range room.clients {
		if client.webSocket != nil {
			clients = append(clients, client)
		}
	}

	h.mutex.RUnlock()

	if len(clients) == 0 {
		return nil
	}

	for _, c := range clients {
		errSend := c.Send(message)
		if errSend != nil {
			continue
		}
	}

	return nil
}
