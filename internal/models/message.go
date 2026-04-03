package models

import "time"

type Message struct {
	ID          int       `json:"id"`
	RoomID      int       `json:"room_id"`
	SenderID    *int      `json:"sender_id,omitempty"`
	Content     string    `json:"content"`
	MessageType string    `json:"message_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DirectMessage struct {
	ID          int       `json:"id"`
	SenderID    *int      `json:"sender_id,omitempty"`
	ReceiverID  *int      `json:"receiver_id,omitempty"`
	Content     string    `json:"content"`
	MessageType string    `json:"message_type"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

type SendMessageRequest struct {
	Content     string `json:"content"`
	MessageType string `json:"message_type,omitempty"`
}

type MessageWithSender struct {
	Message
	SenderUsername string  `json:"sender_username"`
	SenderAvatar   *string `json:"sender_avatar,omitempty"`
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
