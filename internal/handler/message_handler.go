package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chat-go/internal/cache"
	"chat-go/internal/models"
	"chat-go/internal/repository"
)

type MessageHandler struct {
	messageRepo *repository.MessageRepository
	roomRepo    *repository.RoomRepository
	cache       *cache.Cache
}

func NewMessageHandler(messageRepo *repository.MessageRepository, roomRepo *repository.RoomRepository, appCache *cache.Cache) *MessageHandler {
	return &MessageHandler{
		messageRepo: messageRepo,
		roomRepo:    roomRepo,
		cache:       appCache,
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

	// Get query params
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Check for 'after' parameter (get messages after this ID)
	afterID := 0
	if a := r.URL.Query().Get("after"); a != "" {
		if parsed, err := strconv.Atoi(a); err == nil && parsed > 0 {
			afterID = parsed
		}
	}

	// Check for 'unread_only' parameter
	unreadOnly := r.URL.Query().Get("unread_only") == "true"

	var messages []models.DirectMessageWithUsers

	// Use filtered query if after or unread_only is specified
	if afterID > 0 || unreadOnly {
		messages, err = h.messageRepo.GetDirectMessagesFiltered(userID, otherUserID, limit, afterID, unreadOnly)
	} else {
		// Use standard pagination query
		offset := 0
		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}
		messages, err = h.messageRepo.GetDirectMessages(userID, otherUserID, limit, offset)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	// Mark messages as read (only if not fetching unread_only, as that's likely a poll)
	if !unreadOnly {
		h.messageRepo.MarkDirectMessagesAsRead(otherUserID, userID)
	}

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

// EditRoomMessage edits a room message
func (h *MessageHandler) EditRoomMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract message ID from URL path: /api/messages/{messageID}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	messageID, err := strconv.Atoi(parts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	var req models.EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Message content is required")
		return
	}

	message, err := h.messageRepo.EditMessage(messageID, userID, req.Content)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Message not found or you don't have permission to edit it")
		return
	}

	respondWithJSON(w, http.StatusOK, message)
}

// DeleteRoomMessage deletes a room message (soft delete)
func (h *MessageHandler) DeleteRoomMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract message ID from URL path: /api/messages/{messageID}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	messageID, err := strconv.Atoi(parts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	if err := h.messageRepo.DeleteMessage(messageID, userID); err != nil {
		respondWithError(w, http.StatusNotFound, "Message not found or you don't have permission to delete it")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Message deleted"})
}

// EditDirectMessage edits a direct message
func (h *MessageHandler) EditDirectMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract message ID from URL path: /api/dm/messages/{messageID}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	messageID, err := strconv.Atoi(parts[4])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	var req models.EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Message content is required")
		return
	}

	dm, err := h.messageRepo.EditDirectMessage(messageID, userID, req.Content)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Message not found or you don't have permission to edit it")
		return
	}

	respondWithJSON(w, http.StatusOK, dm)
}

// DeleteDirectMessage deletes a direct message (soft delete)
func (h *MessageHandler) DeleteDirectMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract message ID from URL path: /api/dm/messages/{messageID}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	messageID, err := strconv.Atoi(parts[4])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	if err := h.messageRepo.DeleteDirectMessage(messageID, userID); err != nil {
		respondWithError(w, http.StatusNotFound, "Message not found or you don't have permission to delete it")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Message deleted"})
}

// MarkRoomMessagesAsRead marks messages in a room as read
func (h *MessageHandler) MarkRoomMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract room ID from URL path: /api/rooms/{roomID}/messages/read
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	roomID, err := strconv.Atoi(parts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	// Get the message ID to mark up to (optional, defaults to latest)
	var req struct {
		UpToMessageID int `json:"up_to_message_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// If no message ID provided, mark all messages
	if req.UpToMessageID == 0 {
		// Get the latest message ID in the room
		messages, err := h.messageRepo.GetRoomMessages(roomID, 1, 0)
		if err != nil || len(messages) == 0 {
			respondWithJSON(w, http.StatusOK, map[string]string{"message": "No messages to mark as read"})
			return
		}
		req.UpToMessageID = messages[0].ID
	}

	if err := h.messageRepo.MarkRoomMessagesAsRead(userID, roomID, req.UpToMessageID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to mark messages as read")
		return
	}

	// Clear unread count in Redis cache
	if h.cache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.cache.Unread.MarkAsRead(ctx, userID, "room", roomID)
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Messages marked as read"})
}

// GetUnreadRoomMessagesCount returns count of unread messages in a room
func (h *MessageHandler) GetUnreadRoomMessagesCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract room ID from URL path: /api/rooms/{roomID}/messages/unread
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	roomID, err := strconv.Atoi(parts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	count, err := h.messageRepo.GetUnreadRoomMessagesCount(userID, roomID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get unread count")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

// GetUnreadDirectMessagesCount returns count of unread direct messages
func (h *MessageHandler) GetUnreadDirectMessagesCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	count, err := h.messageRepo.GetUnreadCount(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get unread count")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

// GetChatList returns all chats (both DMs and rooms) with unread counts and optional recent messages
func (h *MessageHandler) GetChatList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Check if messages should be included
	includeMessages := r.URL.Query().Get("include_messages") == "true"

	// Get message limit (default 10)
	messageLimit := 10
	if l := r.URL.Query().Get("message_limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			messageLimit = parsed
		}
	}

	chatList, err := h.messageRepo.GetUserChatList(userID, includeMessages, messageLimit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get chat list")
		return
	}

	respondWithJSON(w, http.StatusOK, chatList)
}

// GetChat returns a single chat with its messages
func (h *MessageHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract chat ID and type from URL path: /api/chats/{type}/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "Invalid chat path")
		return
	}

	chatType := parts[3] // "room" or "direct"
	if chatType != "room" && chatType != "direct" {
		respondWithError(w, http.StatusBadRequest, "Chat type must be 'room' or 'direct'")
		return
	}

	chatID, err := strconv.Atoi(parts[4])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chat ID")
		return
	}

	// Get message limit (default 10)
	messageLimit := 10
	if l := r.URL.Query().Get("message_limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			messageLimit = parsed
		}
	}

	// Verify access for rooms
	if chatType == "room" {
		isMember, err := h.roomRepo.IsMember(chatID, userID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to check membership")
			return
		}
		room, err := h.roomRepo.GetByID(chatID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Room not found")
			return
		}
		if room.IsPrivate && !isMember {
			respondWithError(w, http.StatusForbidden, "Access denied")
			return
		}
	}

	chat, err := h.messageRepo.GetChatWithMessages(userID, chatID, chatType, messageLimit)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chat not found")
		return
	}

	respondWithJSON(w, http.StatusOK, chat)
}
