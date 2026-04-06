package models

import "time"

// Notification types
const (
	NotificationTypeMessage       = "message"
	NotificationTypeDirectMessage = "direct_message"
	NotificationTypeMention       = "mention"
	NotificationTypeRoomInvite    = "room_invite"
	NotificationTypeRoomJoin      = "room_join"
	NotificationTypeSystem        = "system"
)

// Notification represents a user notification
type Notification struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	Data        string     `json:"data,omitempty"` // JSON string for additional data
	IsRead      bool       `json:"is_read"`
	IsPushed    bool       `json:"is_pushed"`              // Whether push notification was sent
	ReferenceID *int       `json:"reference_id,omitempty"` // ID of related message/room
	CreatedAt   time.Time  `json:"created_at"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
}

// PushSubscription stores user push notification subscriptions
type PushSubscription struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Endpoint  string    `json:"endpoint"`
	P256dh    string    `json:"p256dh"` // Public key for encryption
	Auth      string    `json:"auth"`   // Authentication secret
	UserAgent string    `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NotificationPreferences stores user notification settings
type NotificationPreferences struct {
	ID                  int       `json:"id"`
	UserID              int       `json:"user_id"`
	EmailNotifications  bool      `json:"email_notifications"`
	PushNotifications   bool      `json:"push_notifications"`
	DirectMessageNotify bool      `json:"direct_message_notify"`
	MentionNotify       bool      `json:"mention_notify"`
	RoomMessageNotify   bool      `json:"room_message_notify"`
	MuteAll             bool      `json:"mute_all"`
	QuietHoursEnabled   bool      `json:"quiet_hours_enabled"`
	QuietHoursStart     *string   `json:"quiet_hours_start,omitempty"` // Time format: "HH:MM"
	QuietHoursEnd       *string   `json:"quiet_hours_end,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// WebSocket notification message
type WSNotification struct {
	Type    string       `json:"type"`
	Payload Notification `json:"payload"`
}

// CreateNotificationRequest represents the request to create a notification
type CreateNotificationRequest struct {
	UserID      int    `json:"user_id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Data        string `json:"data,omitempty"`
	ReferenceID *int   `json:"reference_id,omitempty"`
}

// PushSubscriptionRequest represents the request to register a push subscription
type PushSubscriptionRequest struct {
	Endpoint  string `json:"endpoint"`
	P256dh    string `json:"p256dh"`
	Auth      string `json:"auth"`
	UserAgent string `json:"user_agent,omitempty"`
}

// NotificationPreferencesRequest represents the request to update preferences
type NotificationPreferencesRequest struct {
	EmailNotifications  *bool   `json:"email_notifications,omitempty"`
	PushNotifications   *bool   `json:"push_notifications,omitempty"`
	DirectMessageNotify *bool   `json:"direct_message_notify,omitempty"`
	MentionNotify       *bool   `json:"mention_notify,omitempty"`
	RoomMessageNotify   *bool   `json:"room_message_notify,omitempty"`
	MuteAll             *bool   `json:"mute_all,omitempty"`
	QuietHoursEnabled   *bool   `json:"quiet_hours_enabled,omitempty"`
	QuietHoursStart     *string `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd       *string `json:"quiet_hours_end,omitempty"`
}

// UnreadCount represents unread message counts
type UnreadCount struct {
	TotalUnread         int         `json:"total_unread"`
	DirectMessageUnread int         `json:"direct_message_unread"`
	RoomUnread          map[int]int `json:"room_unread"` // room_id -> count
	NotificationUnread  int         `json:"notification_unread"`
}

// MessageReadStatus tracks which messages a user has read
type MessageReadStatus struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	MessageID int       `json:"message_id"`
	RoomID    int       `json:"room_id"`
	ReadAt    time.Time `json:"read_at"`
}
