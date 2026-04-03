package handler

import (
	"log"
	"net/http"
	"strings"

	"chat-go/internal/repository"
	"chat-go/internal/service"
	ws "chat-go/internal/websocket"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// In production, restrict this to your frontend domain
		return true
	},
}

type WebSocketHandler struct {
	hub         *ws.Hub
	authService *service.AuthService
	userRepo    *repository.UserRepository
	messageRepo *repository.MessageRepository
}

func NewWebSocketHandler(
	hub *ws.Hub,
	authService *service.AuthService,
	userRepo *repository.UserRepository,
	messageRepo *repository.MessageRepository,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		authService: authService,
		userRepo:    userRepo,
		messageRepo: messageRepo,
	}
}

func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Get token from query params or header
	token := r.URL.Query().Get("token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate token
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}

	// Create client
	client := ws.NewClient(h.hub, conn, claims.UserID, claims.Username)

	// Register client
	h.hub.RegisterClient(client)

	// Update user status to online
	h.userRepo.UpdateStatus(claims.UserID, "online")

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump(h.messageRepo)
}
