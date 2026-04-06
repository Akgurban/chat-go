package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"chat-go/internal/models"
	"chat-go/internal/repository"
)

// NotificationService handles all notification-related operations
type NotificationService struct {
	notifRepo   *repository.NotificationRepository
	wsNotifier  WebSocketNotifier
	pushEnabled bool
	vapidKeys   *VAPIDKeys
	mu          sync.RWMutex
}

// WebSocketNotifier interface for sending WebSocket notifications
type WebSocketNotifier interface {
	SendToUser(userID int, message []byte)
	IsUserOnline(userID int) bool
}

// VAPIDKeys for Web Push notifications
type VAPIDKeys struct {
	PublicKey  string
	PrivateKey string
	Subject    string
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	notifRepo *repository.NotificationRepository,
	wsNotifier WebSocketNotifier,
	vapidKeys *VAPIDKeys,
) *NotificationService {
	return &NotificationService{
		notifRepo:   notifRepo,
		wsNotifier:  wsNotifier,
		pushEnabled: vapidKeys != nil && vapidKeys.PublicKey != "",
		vapidKeys:   vapidKeys,
	}
}

// NotifyNewMessage sends notification for a new room message
func (s *NotificationService) NotifyNewMessage(
	senderID int,
	senderUsername string,
	roomID int,
	roomName string,
	messageID int,
	messageContent string,
	memberIDs []int,
) error {
	// Don't notify the sender
	recipientIDs := filterOutUser(memberIDs, senderID)
	if len(recipientIDs) == 0 {
		return nil
	}

	title := fmt.Sprintf("New message in %s", roomName)
	body := fmt.Sprintf("%s: %s", senderUsername, truncateString(messageContent, 100))
	data := repository.ToJSONString(map[string]interface{}{
		"room_id":    roomID,
		"message_id": messageID,
		"sender_id":  senderID,
	})

	// Create notifications for all recipients
	for _, userID := range recipientIDs {
		notification := &models.Notification{
			UserID:      userID,
			Type:        models.NotificationTypeMessage,
			Title:       title,
			Body:        body,
			Data:        data,
			ReferenceID: &messageID,
		}

		if err := s.notifRepo.CreateNotification(notification); err != nil {
			log.Printf("Failed to create notification for user %d: %v", userID, err)
			continue
		}

		// Send via WebSocket if user is online
		s.sendWSNotification(userID, notification)

		// Send push notification if user is offline
		if !s.wsNotifier.IsUserOnline(userID) {
			go s.sendPushNotification(userID, notification)
		}
	}

	return nil
}

// NotifyDirectMessage sends notification for a direct message
func (s *NotificationService) NotifyDirectMessage(
	senderID int,
	senderUsername string,
	receiverID int,
	messageID int,
	messageContent string,
) error {
	title := fmt.Sprintf("Message from %s", senderUsername)
	body := truncateString(messageContent, 100)
	data := repository.ToJSONString(map[string]interface{}{
		"sender_id":  senderID,
		"message_id": messageID,
		"type":       "direct_message",
	})

	notification := &models.Notification{
		UserID:      receiverID,
		Type:        models.NotificationTypeDirectMessage,
		Title:       title,
		Body:        body,
		Data:        data,
		ReferenceID: &messageID,
	}

	if err := s.notifRepo.CreateNotification(notification); err != nil {
		return err
	}

	// Send via WebSocket if user is online
	s.sendWSNotification(receiverID, notification)

	// Send push notification if user is offline
	if !s.wsNotifier.IsUserOnline(receiverID) {
		go s.sendPushNotification(receiverID, notification)
	}

	return nil
}

// NotifyMention sends notification when user is mentioned
func (s *NotificationService) NotifyMention(
	senderID int,
	senderUsername string,
	mentionedUserID int,
	roomID int,
	roomName string,
	messageID int,
	messageContent string,
) error {
	title := fmt.Sprintf("%s mentioned you in %s", senderUsername, roomName)
	body := truncateString(messageContent, 100)
	data := repository.ToJSONString(map[string]interface{}{
		"room_id":    roomID,
		"message_id": messageID,
		"sender_id":  senderID,
	})

	notification := &models.Notification{
		UserID:      mentionedUserID,
		Type:        models.NotificationTypeMention,
		Title:       title,
		Body:        body,
		Data:        data,
		ReferenceID: &messageID,
	}

	if err := s.notifRepo.CreateNotification(notification); err != nil {
		return err
	}

	// Send via WebSocket
	s.sendWSNotification(mentionedUserID, notification)

	// Send push notification if user is offline
	if !s.wsNotifier.IsUserOnline(mentionedUserID) {
		go s.sendPushNotification(mentionedUserID, notification)
	}

	return nil
}

// NotifyRoomInvite sends notification for room invite
func (s *NotificationService) NotifyRoomInvite(
	inviterID int,
	inviterUsername string,
	invitedUserID int,
	roomID int,
	roomName string,
) error {
	title := "Room Invitation"
	body := fmt.Sprintf("%s invited you to join %s", inviterUsername, roomName)
	data := repository.ToJSONString(map[string]interface{}{
		"room_id":    roomID,
		"inviter_id": inviterID,
	})

	notification := &models.Notification{
		UserID:      invitedUserID,
		Type:        models.NotificationTypeRoomInvite,
		Title:       title,
		Body:        body,
		Data:        data,
		ReferenceID: &roomID,
	}

	if err := s.notifRepo.CreateNotification(notification); err != nil {
		return err
	}

	// Send via WebSocket
	s.sendWSNotification(invitedUserID, notification)

	// Send push notification if user is offline
	if !s.wsNotifier.IsUserOnline(invitedUserID) {
		go s.sendPushNotification(invitedUserID, notification)
	}

	return nil
}

// NotifySystem sends a system notification to a user
func (s *NotificationService) NotifySystem(userID int, title, body string) error {
	notification := &models.Notification{
		UserID: userID,
		Type:   models.NotificationTypeSystem,
		Title:  title,
		Body:   body,
	}

	if err := s.notifRepo.CreateNotification(notification); err != nil {
		return err
	}

	// Send via WebSocket
	s.sendWSNotification(userID, notification)

	return nil
}

// BroadcastSystemNotification sends a system notification to multiple users
func (s *NotificationService) BroadcastSystemNotification(userIDs []int, title, body string) error {
	for _, userID := range userIDs {
		s.NotifySystem(userID, title, body)
	}
	return nil
}

// sendWSNotification sends a notification via WebSocket
func (s *NotificationService) sendWSNotification(userID int, notification *models.Notification) {
	wsMsg := models.WSNotification{
		Type:    "notification",
		Payload: *notification,
	}

	data, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal WebSocket notification: %v", err)
		return
	}

	s.wsNotifier.SendToUser(userID, data)
}

// sendPushNotification sends a push notification
func (s *NotificationService) sendPushNotification(userID int, notification *models.Notification) {
	if !s.pushEnabled {
		return
	}

	// Check user preferences
	prefs, err := s.notifRepo.GetPreferences(userID)
	if err != nil || !prefs.PushNotifications || prefs.MuteAll {
		return
	}

	// Check notification type preferences
	switch notification.Type {
	case models.NotificationTypeDirectMessage:
		if !prefs.DirectMessageNotify {
			return
		}
	case models.NotificationTypeMention:
		if !prefs.MentionNotify {
			return
		}
	case models.NotificationTypeMessage:
		if !prefs.RoomMessageNotify {
			return
		}
	}

	// Check quiet hours
	if prefs.QuietHoursEnabled && isQuietHours(prefs.QuietHoursStart, prefs.QuietHoursEnd) {
		return
	}

	// Get user's push subscriptions
	subscriptions, err := s.notifRepo.GetPushSubscriptions(userID)
	if err != nil || len(subscriptions) == 0 {
		return
	}

	// Send push notification to all subscriptions
	for _, sub := range subscriptions {
		go s.sendWebPush(&sub, notification)
	}

	// Mark notification as pushed
	s.notifRepo.MarkAsPushed(notification.ID)
}

// sendWebPush sends a Web Push notification
func (s *NotificationService) sendWebPush(subscription *models.PushSubscription, notification *models.Notification) {
	// Create push payload
	payload := map[string]interface{}{
		"title": notification.Title,
		"body":  notification.Body,
		"icon":  "/icon.png",
		"badge": "/badge.png",
		"tag":   fmt.Sprintf("notification-%d", notification.ID),
		"data": map[string]interface{}{
			"id":   notification.ID,
			"type": notification.Type,
			"url":  "/notifications",
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	// In a production environment, you would use a proper Web Push library
	// such as github.com/SherClockHolmes/webpush-go
	// This is a simplified example
	req, err := http.NewRequest("POST", subscription.Endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		log.Printf("Failed to create push request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("TTL", "86400")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send push notification: %v", err)
		return
	}
	defer resp.Body.Close()

	// If subscription is invalid, remove it
	if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
		s.notifRepo.DeletePushSubscription(subscription.Endpoint, subscription.UserID)
	}
}

// GetNotifications retrieves user notifications
func (s *NotificationService) GetNotifications(userID int, limit, offset int, unreadOnly bool) ([]models.Notification, error) {
	return s.notifRepo.GetNotifications(userID, limit, offset, unreadOnly)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID, userID int) error {
	return s.notifRepo.MarkAsRead(notificationID, userID)
}

// MarkAllAsRead marks all notifications as read
func (s *NotificationService) MarkAllAsRead(userID int) error {
	return s.notifRepo.MarkAllAsRead(userID)
}

// GetUnreadCount gets the unread notification count
func (s *NotificationService) GetUnreadCount(userID int) (int, error) {
	return s.notifRepo.GetUnreadCount(userID)
}

// GetUnreadCounts gets all unread counts
func (s *NotificationService) GetUnreadCounts(userID int) (*models.UnreadCount, error) {
	return s.notifRepo.GetUnreadCounts(userID)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(notificationID, userID int) error {
	return s.notifRepo.DeleteNotification(notificationID, userID)
}

// RegisterPushSubscription registers a push subscription
func (s *NotificationService) RegisterPushSubscription(userID int, req *models.PushSubscriptionRequest) (*models.PushSubscription, error) {
	sub := &models.PushSubscription{
		UserID:    userID,
		Endpoint:  req.Endpoint,
		P256dh:    req.P256dh,
		Auth:      req.Auth,
		UserAgent: req.UserAgent,
	}

	if err := s.notifRepo.CreatePushSubscription(sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// UnregisterPushSubscription removes a push subscription
func (s *NotificationService) UnregisterPushSubscription(userID int, endpoint string) error {
	return s.notifRepo.DeletePushSubscription(endpoint, userID)
}

// GetPreferences gets notification preferences
func (s *NotificationService) GetPreferences(userID int) (*models.NotificationPreferences, error) {
	return s.notifRepo.GetPreferences(userID)
}

// UpdatePreferences updates notification preferences
func (s *NotificationService) UpdatePreferences(userID int, req *models.NotificationPreferencesRequest) (*models.NotificationPreferences, error) {
	return s.notifRepo.UpdatePreferences(userID, req)
}

// Helper functions

func filterOutUser(userIDs []int, excludeID int) []int {
	result := make([]int, 0, len(userIDs))
	for _, id := range userIDs {
		if id != excludeID {
			result = append(result, id)
		}
	}
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func isQuietHours(start, end *string) bool {
	if start == nil || end == nil {
		return false
	}

	now := time.Now()
	currentTime := now.Format("15:04")

	// Simple comparison - assumes same day quiet hours
	if *start < *end {
		return currentTime >= *start && currentTime < *end
	}
	// Overnight quiet hours (e.g., 22:00 to 07:00)
	return currentTime >= *start || currentTime < *end
}
