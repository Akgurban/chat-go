package repository

import (
	"database/sql"
	"fmt"

	"chat-go/internal/models"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Direct messages
func (r *MessageRepository) CreateDirectMessage(msg *models.DirectMessage) error {
	query := `
		INSERT INTO direct_messages (sender_id, receiver_id, content, message_type, delivered_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		RETURNING id, created_at, delivered_at`

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
	).Scan(&msg.ID, &msg.CreatedAt, &msg.DeliveredAt)
}

func (r *MessageRepository) GetDirectMessages(userID1, userID2 int, limit, offset int) ([]models.DirectMessageWithUsers, error) {
	query := `
		SELECT dm.id, dm.sender_id, dm.receiver_id, dm.content, dm.message_type, dm.is_read, dm.created_at, dm.delivered_at, dm.read_at,
			   COALESCE(s.username, 'Deleted User') as sender_username,
			   COALESCE(r.username, 'Deleted User') as receiver_username
		FROM direct_messages dm
		LEFT JOIN users s ON dm.sender_id = s.id
		LEFT JOIN users r ON dm.receiver_id = r.id
		LEFT JOIN chat_cleared cc ON cc.user_id = $1 AND cc.other_user_id = $2
		WHERE ((dm.sender_id = $1 AND dm.receiver_id = $2)
		   OR (dm.sender_id = $2 AND dm.receiver_id = $1))
		   AND (cc.cleared_at IS NULL OR dm.created_at > cc.cleared_at)
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
			&msg.DeliveredAt,
			&msg.ReadAt,
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

// GetDirectMessagesFiltered returns direct messages with optional filters (after message ID, unread only)
func (r *MessageRepository) GetDirectMessagesFiltered(userID1, userID2 int, limit int, afterID int, unreadOnly bool) ([]models.DirectMessageWithUsers, error) {
	query := `
		SELECT dm.id, dm.sender_id, dm.receiver_id, dm.content, dm.message_type, dm.is_read, dm.created_at, dm.delivered_at, dm.read_at,
			   COALESCE(s.username, 'Deleted User') as sender_username,
			   COALESCE(r.username, 'Deleted User') as receiver_username
		FROM direct_messages dm
		LEFT JOIN users s ON dm.sender_id = s.id
		LEFT JOIN users r ON dm.receiver_id = r.id
		LEFT JOIN chat_cleared cc ON cc.user_id = $1 AND cc.other_user_id = $2
		WHERE ((dm.sender_id = $1 AND dm.receiver_id = $2) OR (dm.sender_id = $2 AND dm.receiver_id = $1))
		  AND (cc.cleared_at IS NULL OR dm.created_at > cc.cleared_at)`

	args := []interface{}{userID1, userID2}
	argIndex := 3

	// Filter by after message ID
	if afterID > 0 {
		query += fmt.Sprintf(" AND dm.id > $%d", argIndex)
		args = append(args, afterID)
		argIndex++
	}

	// Filter unread messages only (messages sent TO userID1 that are unread)
	if unreadOnly {
		query += fmt.Sprintf(" AND dm.is_read = false AND dm.receiver_id = $%d", argIndex)
		args = append(args, userID1)
		argIndex++
	}

	query += " ORDER BY dm.created_at ASC"
	query += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := r.db.Query(query, args...)
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
			&msg.DeliveredAt,
			&msg.ReadAt,
			&msg.SenderUsername,
			&msg.ReceiverUsername,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *MessageRepository) MarkDirectMessagesAsRead(senderID, receiverID int) error {
	query := `
		UPDATE direct_messages 
		SET is_read = true, read_at = CURRENT_TIMESTAMP 
		WHERE sender_id = $1 AND receiver_id = $2 AND is_read = false`

	_, err := r.db.Exec(query, senderID, receiverID)
	return err
}

// ClearDirectMessageChat marks the chat as cleared for the user (messages before this time won't show)
func (r *MessageRepository) ClearDirectMessageChat(userID, otherUserID int) error {
	query := `
		INSERT INTO chat_cleared (user_id, other_user_id, cleared_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, other_user_id) 
		DO UPDATE SET cleared_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query, userID, otherUserID)
	return err
}

func (r *MessageRepository) GetUnreadCount(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM direct_messages WHERE receiver_id = $1 AND is_read = false`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// ============================================
// MESSAGE EDIT FUNCTIONS
// ============================================

func (r *MessageRepository) EditDirectMessage(messageID, senderID int, content string) (*models.DirectMessage, error) {
	query := `
		UPDATE direct_messages 
		SET content = $1, is_edited = true, edited_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND sender_id = $3 AND is_deleted = false
		RETURNING id, sender_id, receiver_id, content, message_type, is_read, is_edited, edited_at, is_deleted, created_at, read_at`

	var dm models.DirectMessage
	var readAt sql.NullTime
	err := r.db.QueryRow(query, content, messageID, senderID).Scan(
		&dm.ID,
		&dm.SenderID,
		&dm.ReceiverID,
		&dm.Content,
		&dm.MessageType,
		&dm.IsRead,
		&dm.IsEdited,
		&dm.EditedAt,
		&dm.IsDeleted,
		&dm.CreatedAt,
		&readAt,
	)
	if err != nil {
		return nil, err
	}
	if readAt.Valid {
		dm.ReadAt = &readAt.Time
	}
	return &dm, nil
}

func (r *MessageRepository) DeleteDirectMessage(messageID, senderID int) error {
	query := `
		DELETE FROM direct_messages 
		WHERE id = $1 AND sender_id = $2`

	result, err := r.db.Exec(query, messageID, senderID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MessageRepository) GetDirectMessageByID(messageID int) (*models.DirectMessage, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, message_type, is_read, is_edited, edited_at, is_deleted, created_at, read_at
		FROM direct_messages WHERE id = $1`

	var dm models.DirectMessage
	var readAt sql.NullTime
	err := r.db.QueryRow(query, messageID).Scan(
		&dm.ID,
		&dm.SenderID,
		&dm.ReceiverID,
		&dm.Content,
		&dm.MessageType,
		&dm.IsRead,
		&dm.IsEdited,
		&dm.EditedAt,
		&dm.IsDeleted,
		&dm.CreatedAt,
		&readAt,
	)
	if err != nil {
		return nil, err
	}
	if readAt.Valid {
		dm.ReadAt = &readAt.Time
	}
	return &dm, nil
}

// ============================================
// CHAT LIST - DM chats only
// ============================================

// GetUserChatList returns all DM chats for a user with unread counts and recent messages
func (r *MessageRepository) GetUserChatList(userID int, includeMessages bool, messageLimit int) (*models.ChatListResponse, error) {
	response := &models.ChatListResponse{
		Chats:       []models.ChatListItem{},
		TotalUnread: 0,
	}

	// Get direct message chats
	dmChats, err := r.getDirectMessageChats(userID, includeMessages, messageLimit)
	if err != nil {
		return nil, err
	}

	// Calculate total unread
	for _, chat := range dmChats {
		response.TotalUnread += chat.UnreadCount
	}

	response.Chats = dmChats
	return response, nil
}

// getDirectMessageChats returns all DM conversations for a user
func (r *MessageRepository) getDirectMessageChats(userID int, includeMessages bool, messageLimit int) ([]models.ChatListItem, error) {
	// Get all users the current user has had DM conversations with
	query := `
		SELECT DISTINCT
			u.id,
			u.username,
			u.avatar_url,
			u.status,
			COALESCE((
				SELECT COUNT(*)
				FROM direct_messages dm
				LEFT JOIN chat_cleared cc ON cc.user_id = $1 AND cc.other_user_id = u.id
				WHERE dm.sender_id = u.id AND dm.receiver_id = $1 AND dm.is_read = false
				  AND (cc.cleared_at IS NULL OR dm.created_at > cc.cleared_at)
			), 0) as unread_count,
			(
				SELECT dm.id FROM direct_messages dm 
				LEFT JOIN chat_cleared cc ON cc.user_id = $1 AND cc.other_user_id = u.id
				WHERE ((dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1))
				  AND (cc.cleared_at IS NULL OR dm.created_at > cc.cleared_at)
				ORDER BY dm.created_at DESC LIMIT 1
			) as last_msg_id,
			(
				SELECT dm.content FROM direct_messages dm 
				LEFT JOIN chat_cleared cc ON cc.user_id = $1 AND cc.other_user_id = u.id
				WHERE ((dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1))
				  AND (cc.cleared_at IS NULL OR dm.created_at > cc.cleared_at)
				ORDER BY dm.created_at DESC LIMIT 1
			) as last_msg_content,
			(
				SELECT dm.sender_id FROM direct_messages dm 
				LEFT JOIN chat_cleared cc ON cc.user_id = $1 AND cc.other_user_id = u.id
				WHERE ((dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1))
				  AND (cc.cleared_at IS NULL OR dm.created_at > cc.cleared_at)
				ORDER BY dm.created_at DESC LIMIT 1
			) as last_msg_sender_id,
			(
				SELECT dm.message_type FROM direct_messages dm 
				LEFT JOIN chat_cleared cc ON cc.user_id = $1 AND cc.other_user_id = u.id
				WHERE ((dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1))
				  AND (cc.cleared_at IS NULL OR dm.created_at > cc.cleared_at)
				ORDER BY dm.created_at DESC LIMIT 1
			) as last_msg_type,
			(
				SELECT dm.is_read FROM direct_messages dm 
				LEFT JOIN chat_cleared cc ON cc.user_id = $1 AND cc.other_user_id = u.id
				WHERE ((dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1))
				  AND (cc.cleared_at IS NULL OR dm.created_at > cc.cleared_at)
				ORDER BY dm.created_at DESC LIMIT 1
			) as last_msg_is_read,
			(
				SELECT dm.created_at FROM direct_messages dm 
				LEFT JOIN chat_cleared cc ON cc.user_id = $1 AND cc.other_user_id = u.id
				WHERE ((dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1))
				  AND (cc.cleared_at IS NULL OR dm.created_at > cc.cleared_at)
				ORDER BY dm.created_at DESC LIMIT 1
			) as last_msg_at,
			u.created_at
		FROM users u
		WHERE u.id != $1 AND (
			EXISTS (SELECT 1 FROM direct_messages dm WHERE dm.sender_id = $1 AND dm.receiver_id = u.id)
			OR EXISTS (SELECT 1 FROM direct_messages dm WHERE dm.sender_id = u.id AND dm.receiver_id = $1)
		)
		ORDER BY last_msg_at DESC NULLS LAST`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []models.ChatListItem
	for rows.Next() {
		var chat models.ChatListItem
		var avatarURL sql.NullString
		var status sql.NullString
		var lastMsgID, lastMsgSenderID sql.NullInt64
		var lastMsgContent, lastMsgType sql.NullString
		var lastMsgIsRead sql.NullBool
		var lastMsgAt sql.NullTime

		err := rows.Scan(
			&chat.ID,
			&chat.Name,
			&avatarURL,
			&status,
			&chat.UnreadCount,
			&lastMsgID,
			&lastMsgContent,
			&lastMsgSenderID,
			&lastMsgType,
			&lastMsgIsRead,
			&lastMsgAt,
			&chat.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		chat.Type = "direct"
		if avatarURL.Valid {
			chat.Avatar = &avatarURL.String
		}
		chat.IsOnline = status.Valid && status.String == "online"

		// Set last message if exists
		if lastMsgID.Valid {
			senderID := int(lastMsgSenderID.Int64)
			senderUsername := chat.Name
			if senderID == userID {
				senderUsername = "You"
			}
			chat.LastMessage = &models.ChatLastMessage{
				ID:             int(lastMsgID.Int64),
				Content:        lastMsgContent.String,
				SenderID:       &senderID,
				SenderUsername: senderUsername,
				MessageType:    lastMsgType.String,
				IsRead:         lastMsgIsRead.Bool,
				CreatedAt:      lastMsgAt.Time,
			}
			chat.LastMessageAt = &lastMsgAt.Time
		}

		// Get recent messages if requested
		if includeMessages && messageLimit > 0 {
			messages, _ := r.GetDirectMessages(userID, chat.ID, messageLimit, 0)
			chat.RecentMessages = make([]interface{}, len(messages))
			for i, m := range messages {
				chat.RecentMessages[i] = m
			}
		}

		chats = append(chats, chat)
	}

	return chats, nil
}

// GetChatWithMessages returns a single DM chat with its messages
func (r *MessageRepository) GetChatWithMessages(userID, otherUserID int, messageLimit int) (*models.ChatListItem, error) {
	return r.getDirectChatWithMessages(userID, otherUserID, messageLimit)
}

func (r *MessageRepository) getDirectChatWithMessages(userID, otherUserID, messageLimit int) (*models.ChatListItem, error) {
	query := `
		SELECT 
			u.id,
			u.username,
			u.avatar_url,
			u.status,
			u.created_at,
			COALESCE((
				SELECT COUNT(*)
				FROM direct_messages dm
				LEFT JOIN chat_cleared cc ON cc.user_id = $1 AND cc.other_user_id = u.id
				WHERE dm.sender_id = u.id AND dm.receiver_id = $1 AND dm.is_read = false
				  AND (cc.cleared_at IS NULL OR dm.created_at > cc.cleared_at)
			), 0) as unread_count
		FROM users u
		WHERE u.id = $2`

	var chat models.ChatListItem
	var avatarURL, status sql.NullString

	err := r.db.QueryRow(query, userID, otherUserID).Scan(
		&chat.ID,
		&chat.Name,
		&avatarURL,
		&status,
		&chat.CreatedAt,
		&chat.UnreadCount,
	)
	if err != nil {
		return nil, err
	}

	chat.Type = "direct"
	if avatarURL.Valid {
		chat.Avatar = &avatarURL.String
	}
	chat.IsOnline = status.Valid && status.String == "online"

	// Get recent messages
	if messageLimit > 0 {
		messages, _ := r.GetDirectMessages(userID, otherUserID, messageLimit, 0)
		chat.RecentMessages = make([]interface{}, len(messages))
		for i, m := range messages {
			chat.RecentMessages[i] = m
		}

		// Set last message from recent messages
		if len(messages) > 0 {
			lastMsg := messages[len(messages)-1]
			senderUsername := chat.Name
			if lastMsg.SenderID != nil && *lastMsg.SenderID == userID {
				senderUsername = "You"
			}
			chat.LastMessage = &models.ChatLastMessage{
				ID:             lastMsg.ID,
				Content:        lastMsg.Content,
				SenderID:       lastMsg.SenderID,
				SenderUsername: senderUsername,
				MessageType:    lastMsg.MessageType,
				IsRead:         lastMsg.IsRead,
				CreatedAt:      lastMsg.CreatedAt,
			}
			chat.LastMessageAt = &lastMsg.CreatedAt
		}
	}

	return &chat, nil
}
