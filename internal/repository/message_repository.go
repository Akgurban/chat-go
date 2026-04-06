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

// GetDirectMessagesFiltered returns direct messages with optional filters (after message ID, unread only)
func (r *MessageRepository) GetDirectMessagesFiltered(userID1, userID2 int, limit int, afterID int, unreadOnly bool) ([]models.DirectMessageWithUsers, error) {
	query := `
		SELECT dm.id, dm.sender_id, dm.receiver_id, dm.content, dm.message_type, dm.is_read, dm.created_at,
			   COALESCE(s.username, 'Deleted User') as sender_username,
			   COALESCE(r.username, 'Deleted User') as receiver_username
		FROM direct_messages dm
		LEFT JOIN users s ON dm.sender_id = s.id
		LEFT JOIN users r ON dm.receiver_id = r.id
		WHERE ((dm.sender_id = $1 AND dm.receiver_id = $2) OR (dm.sender_id = $2 AND dm.receiver_id = $1))
		  AND dm.is_deleted = false`

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

func (r *MessageRepository) GetUnreadCount(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM direct_messages WHERE receiver_id = $1 AND is_read = false`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// ============================================
// MESSAGE EDIT FUNCTIONS
// ============================================

func (r *MessageRepository) EditMessage(messageID, senderID int, content string) (*models.Message, error) {
	query := `
		UPDATE messages 
		SET content = $1, is_edited = true, edited_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND sender_id = $3 AND is_deleted = false
		RETURNING id, room_id, sender_id, content, message_type, is_edited, edited_at, is_deleted, created_at, updated_at`

	var msg models.Message
	err := r.db.QueryRow(query, content, messageID, senderID).Scan(
		&msg.ID,
		&msg.RoomID,
		&msg.SenderID,
		&msg.Content,
		&msg.MessageType,
		&msg.IsEdited,
		&msg.EditedAt,
		&msg.IsDeleted,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

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

func (r *MessageRepository) DeleteMessage(messageID, senderID int) error {
	query := `
		UPDATE messages 
		SET is_deleted = true, content = '[Message deleted]', updated_at = CURRENT_TIMESTAMP
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

func (r *MessageRepository) DeleteDirectMessage(messageID, senderID int) error {
	query := `
		UPDATE direct_messages 
		SET is_deleted = true, content = '[Message deleted]'
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

func (r *MessageRepository) GetMessageByID(messageID int) (*models.Message, error) {
	query := `
		SELECT id, room_id, sender_id, content, message_type, is_edited, edited_at, is_deleted, created_at, updated_at
		FROM messages WHERE id = $1`

	var msg models.Message
	err := r.db.QueryRow(query, messageID).Scan(
		&msg.ID,
		&msg.RoomID,
		&msg.SenderID,
		&msg.Content,
		&msg.MessageType,
		&msg.IsEdited,
		&msg.EditedAt,
		&msg.IsDeleted,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &msg, nil
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
// MESSAGE READ STATUS (for room messages)
// ============================================

func (r *MessageRepository) MarkRoomMessageAsRead(userID, messageID, roomID int) error {
	query := `
		INSERT INTO message_read_status (user_id, message_id, room_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, message_id) DO NOTHING`

	_, err := r.db.Exec(query, userID, messageID, roomID)
	return err
}

func (r *MessageRepository) MarkRoomMessagesAsRead(userID, roomID int, upToMessageID int) error {
	query := `
		INSERT INTO message_read_status (user_id, message_id, room_id)
		SELECT $1, m.id, m.room_id
		FROM messages m
		WHERE m.room_id = $2 AND m.id <= $3 AND m.sender_id != $1
		AND NOT EXISTS (
			SELECT 1 FROM message_read_status mrs 
			WHERE mrs.user_id = $1 AND mrs.message_id = m.id
		)`

	_, err := r.db.Exec(query, userID, roomID, upToMessageID)
	return err
}

func (r *MessageRepository) GetUnreadRoomMessagesCount(userID, roomID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		LEFT JOIN message_read_status mrs ON m.id = mrs.message_id AND mrs.user_id = $1
		WHERE m.room_id = $2 AND m.sender_id != $1 AND mrs.id IS NULL AND m.is_deleted = false`

	var count int
	err := r.db.QueryRow(query, userID, roomID).Scan(&count)
	return count, err
}

func (r *MessageRepository) GetMessageReadBy(messageID int) ([]int, error) {
	query := `SELECT user_id FROM message_read_status WHERE message_id = $1`

	rows, err := r.db.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

func (r *MessageRepository) GetLastReadMessageID(userID, roomID int) (int, error) {
	query := `
		SELECT COALESCE(MAX(message_id), 0)
		FROM message_read_status
		WHERE user_id = $1 AND room_id = $2`

	var lastID int
	err := r.db.QueryRow(query, userID, roomID).Scan(&lastID)
	return lastID, err
}

// ============================================
// CHAT LIST - Combined DM and Room chats
// ============================================

// GetUserChatList returns all chats (both DMs and rooms) for a user with unread counts and recent messages
func (r *MessageRepository) GetUserChatList(userID int, includeMessages bool, messageLimit int) (*models.ChatListResponse, error) {
	response := &models.ChatListResponse{
		Chats:       []models.ChatListItem{},
		TotalUnread: 0,
	}

	// Get room chats
	roomChats, err := r.getRoomChats(userID, includeMessages, messageLimit)
	if err != nil {
		return nil, err
	}

	// Get direct message chats
	dmChats, err := r.getDirectMessageChats(userID, includeMessages, messageLimit)
	if err != nil {
		return nil, err
	}

	// Combine and sort by last message time
	allChats := append(roomChats, dmChats...)

	// Sort by last message time (most recent first)
	for i := 0; i < len(allChats); i++ {
		for j := i + 1; j < len(allChats); j++ {
			timeI := allChats[i].LastMessageAt
			timeJ := allChats[j].LastMessageAt

			if timeI == nil && timeJ != nil {
				allChats[i], allChats[j] = allChats[j], allChats[i]
			} else if timeI != nil && timeJ != nil && timeJ.After(*timeI) {
				allChats[i], allChats[j] = allChats[j], allChats[i]
			}
		}
	}

	// Calculate total unread
	for _, chat := range allChats {
		response.TotalUnread += chat.UnreadCount
	}

	response.Chats = allChats
	return response, nil
}

// getRoomChats returns all room chats for a user
func (r *MessageRepository) getRoomChats(userID int, includeMessages bool, messageLimit int) ([]models.ChatListItem, error) {
	query := `
		SELECT 
			r.id,
			r.name,
			r.description,
			r.is_private,
			r.created_at,
			(SELECT COUNT(*) FROM room_members WHERE room_id = r.id) as member_count,
			COALESCE((
				SELECT COUNT(*)
				FROM messages m
				LEFT JOIN message_read_status mrs ON m.id = mrs.message_id AND mrs.user_id = $1
				WHERE m.room_id = r.id AND m.sender_id != $1 AND mrs.id IS NULL AND m.is_deleted = false
			), 0) as unread_count,
			(
				SELECT m.id FROM messages m WHERE m.room_id = r.id AND m.is_deleted = false ORDER BY m.created_at DESC LIMIT 1
			) as last_msg_id,
			(
				SELECT m.content FROM messages m WHERE m.room_id = r.id AND m.is_deleted = false ORDER BY m.created_at DESC LIMIT 1
			) as last_msg_content,
			(
				SELECT m.sender_id FROM messages m WHERE m.room_id = r.id AND m.is_deleted = false ORDER BY m.created_at DESC LIMIT 1
			) as last_msg_sender_id,
			(
				SELECT COALESCE(u.username, 'Deleted User') FROM messages m 
				LEFT JOIN users u ON m.sender_id = u.id 
				WHERE m.room_id = r.id AND m.is_deleted = false ORDER BY m.created_at DESC LIMIT 1
			) as last_msg_sender_username,
			(
				SELECT m.message_type FROM messages m WHERE m.room_id = r.id AND m.is_deleted = false ORDER BY m.created_at DESC LIMIT 1
			) as last_msg_type,
			(
				SELECT m.created_at FROM messages m WHERE m.room_id = r.id AND m.is_deleted = false ORDER BY m.created_at DESC LIMIT 1
			) as last_msg_at
		FROM rooms r
		INNER JOIN room_members rm ON r.id = rm.room_id AND rm.user_id = $1
		ORDER BY last_msg_at DESC NULLS LAST`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []models.ChatListItem
	for rows.Next() {
		var chat models.ChatListItem
		var lastMsgID, lastMsgSenderID sql.NullInt64
		var lastMsgContent, lastMsgSenderUsername, lastMsgType sql.NullString
		var lastMsgAt sql.NullTime

		err := rows.Scan(
			&chat.ID,
			&chat.Name,
			&chat.Description,
			&chat.IsPrivate,
			&chat.CreatedAt,
			&chat.MemberCount,
			&chat.UnreadCount,
			&lastMsgID,
			&lastMsgContent,
			&lastMsgSenderID,
			&lastMsgSenderUsername,
			&lastMsgType,
			&lastMsgAt,
		)
		if err != nil {
			return nil, err
		}

		chat.Type = "room"

		// Set last message if exists
		if lastMsgID.Valid {
			senderID := int(lastMsgSenderID.Int64)
			chat.LastMessage = &models.ChatLastMessage{
				ID:             int(lastMsgID.Int64),
				Content:        lastMsgContent.String,
				SenderID:       &senderID,
				SenderUsername: lastMsgSenderUsername.String,
				MessageType:    lastMsgType.String,
				CreatedAt:      lastMsgAt.Time,
			}
			chat.LastMessageAt = &lastMsgAt.Time
		}

		// Get recent messages if requested
		if includeMessages && messageLimit > 0 {
			messages, _ := r.GetRoomMessages(chat.ID, messageLimit, 0)
			chat.RecentMessages = make([]interface{}, len(messages))
			for i, m := range messages {
				chat.RecentMessages[i] = m
			}
		}

		chats = append(chats, chat)
	}

	return chats, nil
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
				WHERE dm.sender_id = u.id AND dm.receiver_id = $1 AND dm.is_read = false AND dm.is_deleted = false
			), 0) as unread_count,
			(
				SELECT dm.id FROM direct_messages dm 
				WHERE (dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1)
				AND dm.is_deleted = false
				ORDER BY dm.created_at DESC LIMIT 1
			) as last_msg_id,
			(
				SELECT dm.content FROM direct_messages dm 
				WHERE ((dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1))
				AND dm.is_deleted = false
				ORDER BY dm.created_at DESC LIMIT 1
			) as last_msg_content,
			(
				SELECT dm.sender_id FROM direct_messages dm 
				WHERE ((dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1))
				AND dm.is_deleted = false
				ORDER BY dm.created_at DESC LIMIT 1
			) as last_msg_sender_id,
			(
				SELECT dm.message_type FROM direct_messages dm 
				WHERE ((dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1))
				AND dm.is_deleted = false
				ORDER BY dm.created_at DESC LIMIT 1
			) as last_msg_type,
			(
				SELECT dm.is_read FROM direct_messages dm 
				WHERE ((dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1))
				AND dm.is_deleted = false
				ORDER BY dm.created_at DESC LIMIT 1
			) as last_msg_is_read,
			(
				SELECT dm.created_at FROM direct_messages dm 
				WHERE ((dm.sender_id = $1 AND dm.receiver_id = u.id) OR (dm.sender_id = u.id AND dm.receiver_id = $1))
				AND dm.is_deleted = false
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

// GetChatWithMessages returns a single chat with its messages
func (r *MessageRepository) GetChatWithMessages(userID, chatID int, chatType string, messageLimit int) (*models.ChatListItem, error) {
	if chatType == "room" {
		return r.getRoomChatWithMessages(userID, chatID, messageLimit)
	}
	return r.getDirectChatWithMessages(userID, chatID, messageLimit)
}

func (r *MessageRepository) getRoomChatWithMessages(userID, roomID, messageLimit int) (*models.ChatListItem, error) {
	query := `
		SELECT 
			r.id,
			r.name,
			r.description,
			r.is_private,
			r.created_at,
			(SELECT COUNT(*) FROM room_members WHERE room_id = r.id) as member_count,
			COALESCE((
				SELECT COUNT(*)
				FROM messages m
				LEFT JOIN message_read_status mrs ON m.id = mrs.message_id AND mrs.user_id = $1
				WHERE m.room_id = r.id AND m.sender_id != $1 AND mrs.id IS NULL AND m.is_deleted = false
			), 0) as unread_count
		FROM rooms r
		WHERE r.id = $2`

	var chat models.ChatListItem
	err := r.db.QueryRow(query, userID, roomID).Scan(
		&chat.ID,
		&chat.Name,
		&chat.Description,
		&chat.IsPrivate,
		&chat.CreatedAt,
		&chat.MemberCount,
		&chat.UnreadCount,
	)
	if err != nil {
		return nil, err
	}

	chat.Type = "room"

	// Get recent messages
	if messageLimit > 0 {
		messages, _ := r.GetRoomMessages(roomID, messageLimit, 0)
		chat.RecentMessages = make([]interface{}, len(messages))
		for i, m := range messages {
			chat.RecentMessages[i] = m
		}

		// Set last message from recent messages
		if len(messages) > 0 {
			lastMsg := messages[len(messages)-1]
			chat.LastMessage = &models.ChatLastMessage{
				ID:             lastMsg.ID,
				Content:        lastMsg.Content,
				SenderID:       lastMsg.SenderID,
				SenderUsername: lastMsg.SenderUsername,
				MessageType:    lastMsg.MessageType,
				CreatedAt:      lastMsg.CreatedAt,
			}
			chat.LastMessageAt = &lastMsg.CreatedAt
		}
	}

	return &chat, nil
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
				WHERE dm.sender_id = u.id AND dm.receiver_id = $1 AND dm.is_read = false AND dm.is_deleted = false
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
