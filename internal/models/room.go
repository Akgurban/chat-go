package models

import "time"

type Room struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsPrivate   bool      `json:"is_private"`
	CreatedBy   *int      `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoomMember struct {
	ID       int       `json:"id"`
	RoomID   int       `json:"room_id"`
	UserID   int       `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type CreateRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsPrivate   bool   `json:"is_private"`
}

type RoomWithMembers struct {
	Room
	Members []UserResponse `json:"members"`
}
