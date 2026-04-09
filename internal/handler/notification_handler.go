package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"chat-go/internal/models"
	"chat-go/internal/service"
)

type NotificationHandler struct {
	notifService *service.NotificationService
}

func NewNotificationHandler(notifService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notifService: notifService,
	}
}

// GetNotifications retrieves user notifications
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Get pagination params
	limit := 20
	offset := 0
	unreadOnly := false

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

	if u := r.URL.Query().Get("unread"); u == "true" {
		unreadOnly = true
	}

	notifications, err := h.notifService.GetNotifications(userID, limit, offset, unreadOnly)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get notifications")
		return
	}

	if notifications == nil {
		notifications = []models.Notification{}
	}

	respondWithJSON(w, http.StatusOK, notifications)
}

// GetUnreadCount returns the count of unread notifications
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	count, err := h.notifService.GetUnreadCount(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get unread count")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]int{"count": count})
}

// GetUnreadCounts returns all unread counts (notifications, DMs)
func (h *NotificationHandler) GetUnreadCounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	counts, err := h.notifService.GetUnreadCounts(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get unread counts")
		return
	}

	respondWithJSON(w, http.StatusOK, counts)
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract notification ID from URL: /api/notifications/{id}/read
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	notifID, err := strconv.Atoi(parts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	if err := h.notifService.MarkAsRead(notifID, userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to mark notification as read")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Notification marked as read"})
}

// MarkAllAsRead marks all notifications as read
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	if err := h.notifService.MarkAllAsRead(userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to mark all notifications as read")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "All notifications marked as read"})
}

// DeleteNotification deletes a notification
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract notification ID from URL: /api/notifications/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	notifID, err := strconv.Atoi(parts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	if err := h.notifService.DeleteNotification(notifID, userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete notification")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Notification deleted"})
}

// GetPreferences returns user notification preferences
func (h *NotificationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	prefs, err := h.notifService.GetPreferences(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get preferences")
		return
	}

	respondWithJSON(w, http.StatusOK, prefs)
}

// UpdatePreferences updates user notification preferences
func (h *NotificationHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	var req models.NotificationPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	prefs, err := h.notifService.UpdatePreferences(userID, &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update preferences")
		return
	}

	respondWithJSON(w, http.StatusOK, prefs)
}

// RegisterPushSubscription registers a push notification subscription
func (h *NotificationHandler) RegisterPushSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	var req models.PushSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Endpoint == "" || req.P256dh == "" || req.Auth == "" {
		respondWithError(w, http.StatusBadRequest, "Endpoint, p256dh, and auth are required")
		return
	}

	sub, err := h.notifService.RegisterPushSubscription(userID, &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to register push subscription")
		return
	}

	respondWithJSON(w, http.StatusCreated, sub)
}

// UnregisterPushSubscription removes a push notification subscription
func (h *NotificationHandler) UnregisterPushSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		respondWithError(w, http.StatusBadRequest, "Endpoint is required")
		return
	}

	if err := h.notifService.UnregisterPushSubscription(userID, endpoint); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to unregister push subscription")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Push subscription removed"})
}
