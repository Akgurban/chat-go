package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"chat-go/internal/cache"
	"chat-go/internal/models"
	"chat-go/internal/repository"
)

type MessageHandler struct {
	messageRepo *repository.MessageRepository
	cache       *cache.Cache
}

func NewMessageHandler(messageRepo *repository.MessageRepository, appCache *cache.Cache) *MessageHandler {
	return &MessageHandler{
		messageRepo: messageRepo,
		cache:       appCache,
	}
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

	if messages == nil {
		messages = []models.DirectMessageWithUsers{}
	}

	respondWithJSON(w, http.StatusOK, messages)
}

// MarkDirectMessagesRead marks all direct messages from a user as read
// POST /api/dm/read/{userID}
func (h *MessageHandler) MarkDirectMessagesRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract other user ID from URL path: /api/dm/read/{userID}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	otherUserID, err := strconv.Atoi(parts[4])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Mark all messages from otherUserID to current user as read
	if err := h.messageRepo.MarkDirectMessagesAsRead(otherUserID, userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to mark messages as read")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "Messages marked as read"})
}

// ClearDirectMessageChat clears all messages in a DM conversation
// DELETE /api/dm/clear/{userID}
func (h *MessageHandler) ClearDirectMessageChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract other user ID from URL path: /api/dm/clear/{userID}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	otherUserID, err := strconv.Atoi(parts[4])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Clear all messages in the conversation
	if err := h.messageRepo.ClearDirectMessageChat(userID, otherUserID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to clear chat")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "Chat cleared"})
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

// GetChatList returns all DM chats with unread counts and optional recent messages
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

// GetChat returns a single DM chat with its messages
func (h *MessageHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract chat ID from URL path: /api/chats/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "Invalid chat path")
		return
	}

	chatID, err := strconv.Atoi(parts[3])
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

	chat, err := h.messageRepo.GetChatWithMessages(userID, chatID, messageLimit)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chat not found")
		return
	}

	respondWithJSON(w, http.StatusOK, chat)
}
