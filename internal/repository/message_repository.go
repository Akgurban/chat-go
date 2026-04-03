package repository

import (
	"database/sql"

	"chat-go/internal/models"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Room messages
func (r *MessageRepository) CreateMessage(msg *models.Message) error {
	query := `
		INSERT INTO messages (room_id, sender_id, content, message_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	messageType := msg.MessageType
	if messageType == "" {
		messageType = "text"
	}

	return r.db.QueryRow(
		query,
		msg.RoomID,
		msg.SenderID,
		msg.Content,
		messageType,
	).Scan(&msg.ID, &msg.CreatedAt, &msg.UpdatedAt)
}

func (r *MessageRepository) GetRoomMessages(roomID int, limit, offset int) ([]models.MessageWithSender, error) {
	query := `
		SELECT m.id, m.room_id, m.sender_id, m.content, m.message_type, m.created_at, m.updated_at,
			   COALESCE(u.username, 'Deleted User') as sender_username, u.avatar_url
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.room_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.MessageWithSender
	for rows.Next() {
		var msg models.MessageWithSender
		err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.SenderID,
			&msg.Content,
			&msg.MessageType,
			&msg.CreatedAt,
			&msg.UpdatedAt,
			&msg.SenderUsername,
			&msg.SenderAvatar,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// Direct messages
func (r *MessageRepository) CreateDirectMessage(msg *models.DirectMessage) error {
	query := `
		INSERT INTO direct_messages (sender_id, receiver_id, content, message_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	messageType := msg.MessageType
	if messageType == "" {
		messageType = "text"
	}

	return r.db.QueryRow(
		query,
		msg.SenderID,
		msg.ReceiverID,
		msg.Content,
		messageType,
	).Scan(&msg.ID, &msg.CreatedAt)
}

func (r *MessageRepository) GetDirectMessages(userID1, userID2 int, limit, offset int) ([]models.DirectMessageWithUsers, error) {
	query := `
		SELECT dm.id, dm.sender_id, dm.receiver_id, dm.content, dm.message_type, dm.is_read, dm.created_at,
			   COALESCE(s.username, 'Deleted User') as sender_username,
			   COALESCE(r.username, 'Deleted User') as receiver_username
		FROM direct_messages dm
		LEFT JOIN users s ON dm.sender_id = s.id
		LEFT JOIN users r ON dm.receiver_id = r.id
		WHERE (dm.sender_id = $1 AND dm.receiver_id = $2)
		   OR (dm.sender_id = $2 AND dm.receiver_id = $1)
		ORDER BY dm.created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(query, userID1, userID2, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.DirectMessageWithUsers
	for rows.Next() {
		var msg models.DirectMessageWithUsers
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.MessageType,
			&msg.IsRead,
			&msg.CreatedAt,
			&msg.SenderUsername,
			&msg.ReceiverUsername,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *MessageRepository) MarkDirectMessagesAsRead(senderID, receiverID int) error {
	query := `
		UPDATE direct_messages 
		SET is_read = true 
		WHERE sender_id = $1 AND receiver_id = $2 AND is_read = false`

	_, err := r.db.Exec(query, senderID, receiverID)
	return err
}

func (r *MessageRepository) GetUnreadCount(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM direct_messages WHERE receiver_id = $1 AND is_read = false`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}
