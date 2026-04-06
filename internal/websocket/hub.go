package websocket

import (
	"sync"
)

type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Clients mapped by user ID for direct messaging
	userClients map[int]*Client

	// Clients mapped by room ID
	roomClients map[int]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to a room
	broadcast chan *BroadcastMessage

	// Direct message to a specific user
	directMessage chan *DirectMessagePayload

	mu sync.RWMutex
}

type BroadcastMessage struct {
	RoomID  int
	Message []byte
}

type DirectMessagePayload struct {
	ReceiverID int
	Message    []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		userClients:   make(map[int]*Client),
		roomClients:   make(map[int]map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan *BroadcastMessage),
		directMessage: make(chan *DirectMessagePayload),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.userClients[client.UserID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.userClients, client.UserID)

				// Remove from all rooms
				for roomID, clients := range h.roomClients {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.roomClients, roomID)
					}
				}

				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.roomClients[message.RoomID]; ok {
				for client := range clients {
					select {
					case client.send <- message.Message:
					default:
						close(client.send)
						delete(h.clients, client)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()

		case dm := <-h.directMessage:
			h.mu.RLock()
			if client, ok := h.userClients[dm.ReceiverID]; ok {
				select {
				case client.send <- dm.Message:
				default:
					// Client buffer full, skip
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) JoinRoom(client *Client, roomID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.roomClients[roomID]; !ok {
		h.roomClients[roomID] = make(map[*Client]bool)
	}
	h.roomClients[roomID][client] = true
	client.Rooms[roomID] = true
}

func (h *Hub) LeaveRoom(client *Client, roomID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.roomClients[roomID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.roomClients, roomID)
		}
	}
	delete(client.Rooms, roomID)
}

func (h *Hub) BroadcastToRoom(roomID int, message []byte) {
	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: message,
	}
}

func (h *Hub) SendDirectMessage(receiverID int, message []byte) {
	h.directMessage <- &DirectMessagePayload{
		ReceiverID: receiverID,
		Message:    message,
	}
}

func (h *Hub) IsUserOnline(userID int) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.userClients[userID]
	return ok
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// SendToUser sends a message to a specific user (implements WebSocketNotifier interface)
func (h *Hub) SendToUser(userID int, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.userClients[userID]; ok {
		select {
		case client.send <- message:
		default:
			// Client buffer full, skip
		}
	}
}

// GetOnlineUsers returns a list of online user IDs
func (h *Hub) GetOnlineUsers() []int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userIDs := make([]int, 0, len(h.userClients))
	for userID := range h.userClients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// GetRoomMembers returns online users in a specific room
func (h *Hub) GetRoomOnlineMembers(roomID int) []int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var userIDs []int
	if clients, ok := h.roomClients[roomID]; ok {
		for client := range clients {
			userIDs = append(userIDs, client.UserID)
		}
	}
	return userIDs
}
