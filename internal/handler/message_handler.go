package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"chat-go/internal/models"
	"chat-go/internal/repository"
)

type MessageHandler struct {
	messageRepo *repository.MessageRepository
	roomRepo    *repository.RoomRepository
}

func NewMessageHandler(messageRepo *repository.MessageRepository, roomRepo *repository.RoomRepository) *MessageHandler {
	return &MessageHandler{
		messageRepo: messageRepo,
		roomRepo:    roomRepo,
	}
}

func (h *MessageHandler) GetRoomMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract room ID from URL path: /api/rooms/{roomID}/messages
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	roomID, err := strconv.Atoi(parts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	// Check if user is a member of the room
	isMember, err := h.roomRepo.IsMember(roomID, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	room, err := h.roomRepo.GetByID(roomID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Room not found")
		return
	}

	// If room is private and user is not a member, deny access
	if room.IsPrivate && !isMember {
		respondWithError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Get pagination params
	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	messages, err := h.messageRepo.GetRoomMessages(roomID, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	if messages == nil {
		messages = []models.MessageWithSender{}
	}

	respondWithJSON(w, http.StatusOK, messages)
}

func (h *MessageHandler) SendRoomMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract room ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	roomID, err := strconv.Atoi(parts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	// Check if user is a member of the room
	isMember, err := h.roomRepo.IsMember(roomID, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		respondWithError(w, http.StatusForbidden, "You must be a member to send messages")
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Message content is required")
		return
	}

	message := &models.Message{
		RoomID:      roomID,
		SenderID:    &userID,
		Content:     req.Content,
		MessageType: req.MessageType,
	}

	if err := h.messageRepo.CreateMessage(message); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	respondWithJSON(w, http.StatusCreated, message)
}

func (h *MessageHandler) GetDirectMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract other user ID from URL path: /api/dm/{userID}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	otherUserID, err := strconv.Atoi(parts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get pagination params
	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	messages, err := h.messageRepo.GetDirectMessages(userID, otherUserID, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	// Mark messages as read
	h.messageRepo.MarkDirectMessagesAsRead(otherUserID, userID)

	if messages == nil {
		messages = []models.DirectMessageWithUsers{}
	}

	respondWithJSON(w, http.StatusOK, messages)
}

func (h *MessageHandler) SendDirectMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract receiver ID from URL path: /api/dm/{userID}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	receiverID, err := strconv.Atoi(parts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Message content is required")
		return
	}

	dm := &models.DirectMessage{
		SenderID:    &userID,
		ReceiverID:  &receiverID,
		Content:     req.Content,
		MessageType: req.MessageType,
	}

	if err := h.messageRepo.CreateDirectMessage(dm); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	respondWithJSON(w, http.StatusCreated, dm)
}
