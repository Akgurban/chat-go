package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"chat-go/internal/cache"
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
	cache       *cache.Cache
}

func NewWebSocketHandler(
	hub *ws.Hub,
	authService *service.AuthService,
	userRepo *repository.UserRepository,
	messageRepo *repository.MessageRepository,
	appCache *cache.Cache,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		authService: authService,
		userRepo:    userRepo,
		messageRepo: messageRepo,
		cache:       appCache,
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

	// Update user status to online in database
	h.userRepo.UpdateStatus(claims.UserID, "online")

	// Broadcast user_online to all clients
	onlineMsg := map[string]interface{}{
		"type": "user_online",
		"payload": map[string]interface{}{
			"user_id":  claims.UserID,
			"username": claims.Username,
		},
	}
	if data, err := json.Marshal(onlineMsg); err == nil {
		h.hub.BroadcastAll(data)
	}

	// Set user online in Redis cache (for cross-server presence)
	if h.cache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.cache.Presence.SetOnline(ctx, claims.UserID); err != nil {
			log.Printf("Failed to set user online in Redis: %v", err)
		}
	}

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump(h.messageRepo, h.userRepo, h.cache)
}
