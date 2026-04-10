package models

import "time"

type DirectMessage struct {
	ID          int        `json:"id"`
	SenderID    *int       `json:"sender_id,omitempty"`
	ReceiverID  *int       `json:"receiver_id,omitempty"`
	Content     string     `json:"content"`
	MessageType string     `json:"message_type"`
	IsRead      bool       `json:"is_read"`
	IsEdited    bool       `json:"is_edited"`
	EditedAt    *time.Time `json:"edited_at,omitempty"`
	IsDeleted   bool       `json:"is_deleted"`
	CreatedAt   time.Time  `json:"created_at"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"` // Single tick - message delivered to server
	ReadAt      *time.Time `json:"read_at,omitempty"`      // Double tick - message read by recipient
}

type SendMessageRequest struct {
	Content     string `json:"content"`
	MessageType string `json:"message_type,omitempty"`
}

type EditMessageRequest struct {
	Content string `json:"content"`
}

type DirectMessageWithUsers struct {
	DirectMessage
	SenderUsername   string `json:"sender_username"`
	ReceiverUsername string `json:"receiver_username"`
}

// WebSocket message types
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// ChatListItem represents a single chat in the chat list (DM)
type ChatListItem struct {
	ID             int              `json:"id"`                        // User ID (for DM)
	Type           string           `json:"type"`                      // "direct"
	Name           string           `json:"name"`                      // Username
	Avatar         *string          `json:"avatar,omitempty"`          // User avatar
	UnreadCount    int              `json:"unread_count"`              // Number of unread messages
	LastMessage    *ChatLastMessage `json:"last_message,omitempty"`    // Last message preview
	LastMessageAt  *time.Time       `json:"last_message_at,omitempty"` // For sorting
	RecentMessages []interface{}    `json:"recent_messages,omitempty"` // Last messages
	IsOnline       bool             `json:"is_online,omitempty"`       // User online status
	LastSeenAt     *time.Time       `json:"last_seen_at,omitempty"`    // When user was last online
	CreatedAt      time.Time        `json:"created_at"`
}

// ChatLastMessage represents the last message preview in a chat
type ChatLastMessage struct {
	ID             int       `json:"id"`
	Content        string    `json:"content"`
	SenderID       *int      `json:"sender_id,omitempty"`
	SenderUsername string    `json:"sender_username"`
	MessageType    string    `json:"message_type"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

// ChatListResponse is the response for getting all chats
type ChatListResponse struct {
	Chats       []ChatListItem `json:"chats"`
	TotalUnread int            `json:"total_unread"`
}

// PaginatedMessagesResponse is the response for paginated direct messages
type PaginatedMessagesResponse struct {
	Messages    []DirectMessageWithUsers `json:"messages"`
	CurrentPage int                      `json:"current_page"`
	TotalPages  int                      `json:"total_pages"`
	TotalCount  int                      `json:"total_count"`
	Limit       int                      `json:"limit"`
	HasMore     bool                     `json:"has_more"`
}
