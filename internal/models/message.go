package models

import "time"

type Message struct {
	ID          int        `json:"id"`
	RoomID      int        `json:"room_id"`
	SenderID    *int       `json:"sender_id,omitempty"`
	Content     string     `json:"content"`
	MessageType string     `json:"message_type"`
	IsEdited    bool       `json:"is_edited"`
	EditedAt    *time.Time `json:"edited_at,omitempty"`
	IsDeleted   bool       `json:"is_deleted"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

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
	ReadAt      *time.Time `json:"read_at,omitempty"`
}

type SendMessageRequest struct {
	Content     string `json:"content"`
	MessageType string `json:"message_type,omitempty"`
}

type EditMessageRequest struct {
	Content string `json:"content"`
}

type MessageWithSender struct {
	Message
	SenderUsername string  `json:"sender_username"`
	SenderAvatar   *string `json:"sender_avatar,omitempty"`
	ReadByCount    int     `json:"read_by_count,omitempty"`
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

type WSChatMessage struct {
	RoomID  int    `json:"room_id,omitempty"`
	UserID  int    `json:"user_id,omitempty"`
	Content string `json:"content"`
}

// ChatListItem represents a single chat in the combined chat list (both DM and group)
type ChatListItem struct {
	ID             int              `json:"id"`                        // Room ID or User ID (for DM)
	Type           string           `json:"type"`                      // "room" or "direct"
	Name           string           `json:"name"`                      // Room name or username
	Avatar         *string          `json:"avatar,omitempty"`          // Room avatar or user avatar
	Description    *string          `json:"description,omitempty"`     // Room description
	IsPrivate      bool             `json:"is_private,omitempty"`      // For rooms
	UnreadCount    int              `json:"unread_count"`              // Number of unread messages
	LastMessage    *ChatLastMessage `json:"last_message,omitempty"`    // Last message preview
	LastMessageAt  *time.Time       `json:"last_message_at,omitempty"` // For sorting
	RecentMessages []interface{}    `json:"recent_messages,omitempty"` // Last 10 messages
	MemberCount    int              `json:"member_count,omitempty"`    // For rooms
	IsOnline       bool             `json:"is_online,omitempty"`       // For DM users
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
