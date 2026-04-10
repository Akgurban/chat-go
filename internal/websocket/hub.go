package websocket

import (
	"sync"
)

type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Clients mapped by user ID for direct messaging
	userClients map[int]*Client

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Direct message to a specific user
	directMessage chan *DirectMessagePayload

	mu sync.RWMutex
}

type DirectMessagePayload struct {
	ReceiverID int
	Message    []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		userClients:   make(map[int]*Client),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
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
				close(client.send)
			}
			h.mu.Unlock()

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

// BroadcastAll sends a message to all connected clients
func (h *Hub) BroadcastAll(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			// Client buffer full, skip
		}
	}
}
