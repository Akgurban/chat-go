package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"chat-go/internal/models"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// ============================================
// NOTIFICATIONS
// ============================================

func (r *NotificationRepository) CreateNotification(notification *models.Notification) error {
	query := `
		INSERT INTO notifications (user_id, type, title, body, data, reference_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.db.QueryRow(
		query,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Body,
		notification.Data,
		notification.ReferenceID,
	).Scan(&notification.ID, &notification.CreatedAt)
}

func (r *NotificationRepository) GetNotifications(userID int, limit, offset int, unreadOnly bool) ([]models.Notification, error) {
	query := `
		SELECT id, user_id, type, title, body, data, is_read, is_pushed, reference_id, created_at, read_at
		FROM notifications
		WHERE user_id = $1`

	if unreadOnly {
		query += ` AND is_read = false`
	}

	query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		var data sql.NullString
		var referenceID sql.NullInt64
		var readAt sql.NullTime

		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Body,
			&data,
			&n.IsRead,
			&n.IsPushed,
			&referenceID,
			&n.CreatedAt,
			&readAt,
		)
		if err != nil {
			return nil, err
		}

		if data.Valid {
			n.Data = data.String
		}
		if referenceID.Valid {
			refID := int(referenceID.Int64)
			n.ReferenceID = &refID
		}
		if readAt.Valid {
			n.ReadAt = &readAt.Time
		}

		notifications = append(notifications, n)
	}

	return notifications, nil
}

func (r *NotificationRepository) MarkAsRead(notificationID, userID int) error {
	query := `
		UPDATE notifications 
		SET is_read = true, read_at = $1 
		WHERE id = $2 AND user_id = $3`

	_, err := r.db.Exec(query, time.Now(), notificationID, userID)
	return err
}

func (r *NotificationRepository) MarkAllAsRead(userID int) error {
	query := `
		UPDATE notifications 
		SET is_read = true, read_at = $1 
		WHERE user_id = $2 AND is_read = false`

	_, err := r.db.Exec(query, time.Now(), userID)
	return err
}

func (r *NotificationRepository) GetUnreadCount(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

func (r *NotificationRepository) DeleteNotification(notificationID, userID int) error {
	query := `DELETE FROM notifications WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(query, notificationID, userID)
	return err
}

func (r *NotificationRepository) MarkAsPushed(notificationID int) error {
	query := `UPDATE notifications SET is_pushed = true WHERE id = $1`
	_, err := r.db.Exec(query, notificationID)
	return err
}

// ============================================
// PUSH SUBSCRIPTIONS
// ============================================

func (r *NotificationRepository) CreatePushSubscription(sub *models.PushSubscription) error {
	// First, try to update existing subscription with same endpoint
	query := `
		INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth, user_agent)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, endpoint) 
		DO UPDATE SET p256dh = EXCLUDED.p256dh, auth = EXCLUDED.auth, 
		              user_agent = EXCLUDED.user_agent, updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(
		query,
		sub.UserID,
		sub.Endpoint,
		sub.P256dh,
		sub.Auth,
		sub.UserAgent,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
}

func (r *NotificationRepository) GetPushSubscriptions(userID int) ([]models.PushSubscription, error) {
	query := `
		SELECT id, user_id, endpoint, p256dh, auth, user_agent, created_at, updated_at
		FROM push_subscriptions
		WHERE user_id = $1`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []models.PushSubscription
	for rows.Next() {
		var sub models.PushSubscription
		var userAgent sql.NullString

		err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.Endpoint,
			&sub.P256dh,
			&sub.Auth,
			&userAgent,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if userAgent.Valid {
			sub.UserAgent = userAgent.String
		}

		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *NotificationRepository) DeletePushSubscription(endpoint string, userID int) error {
	query := `DELETE FROM push_subscriptions WHERE endpoint = $1 AND user_id = $2`
	_, err := r.db.Exec(query, endpoint, userID)
	return err
}

func (r *NotificationRepository) DeleteAllPushSubscriptions(userID int) error {
	query := `DELETE FROM push_subscriptions WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

// ============================================
// NOTIFICATION PREFERENCES
// ============================================

func (r *NotificationRepository) GetPreferences(userID int) (*models.NotificationPreferences, error) {
	query := `
		SELECT id, user_id, email_notifications, push_notifications, direct_message_notify,
		       mention_notify, room_message_notify, mute_all, quiet_hours_enabled,
		       quiet_hours_start, quiet_hours_end, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1`

	var prefs models.NotificationPreferences
	var quietStart, quietEnd sql.NullString

	err := r.db.QueryRow(query, userID).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.EmailNotifications,
		&prefs.PushNotifications,
		&prefs.DirectMessageNotify,
		&prefs.MentionNotify,
		&prefs.RoomMessageNotify,
		&prefs.MuteAll,
		&prefs.QuietHoursEnabled,
		&quietStart,
		&quietEnd,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create default preferences
		return r.CreateDefaultPreferences(userID)
	}
	if err != nil {
		return nil, err
	}

	if quietStart.Valid {
		prefs.QuietHoursStart = &quietStart.String
	}
	if quietEnd.Valid {
		prefs.QuietHoursEnd = &quietEnd.String
	}

	return &prefs, nil
}

func (r *NotificationRepository) CreateDefaultPreferences(userID int) (*models.NotificationPreferences, error) {
	query := `
		INSERT INTO notification_preferences (user_id)
		VALUES ($1)
		RETURNING id, user_id, email_notifications, push_notifications, direct_message_notify,
		          mention_notify, room_message_notify, mute_all, quiet_hours_enabled,
		          quiet_hours_start, quiet_hours_end, created_at, updated_at`

	var prefs models.NotificationPreferences
	var quietStart, quietEnd sql.NullString

	err := r.db.QueryRow(query, userID).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.EmailNotifications,
		&prefs.PushNotifications,
		&prefs.DirectMessageNotify,
		&prefs.MentionNotify,
		&prefs.RoomMessageNotify,
		&prefs.MuteAll,
		&prefs.QuietHoursEnabled,
		&quietStart,
		&quietEnd,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if quietStart.Valid {
		prefs.QuietHoursStart = &quietStart.String
	}
	if quietEnd.Valid {
		prefs.QuietHoursEnd = &quietEnd.String
	}

	return &prefs, nil
}

func (r *NotificationRepository) UpdatePreferences(userID int, req *models.NotificationPreferencesRequest) (*models.NotificationPreferences, error) {
	// Build dynamic update query
	query := `UPDATE notification_preferences SET updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{}
	argIndex := 1

	if req.EmailNotifications != nil {
		query += fmt.Sprintf(", email_notifications = $%d", argIndex)
		args = append(args, *req.EmailNotifications)
		argIndex++
	}
	if req.PushNotifications != nil {
		query += fmt.Sprintf(", push_notifications = $%d", argIndex)
		args = append(args, *req.PushNotifications)
		argIndex++
	}
	if req.DirectMessageNotify != nil {
		query += fmt.Sprintf(", direct_message_notify = $%d", argIndex)
		args = append(args, *req.DirectMessageNotify)
		argIndex++
	}
	if req.MentionNotify != nil {
		query += fmt.Sprintf(", mention_notify = $%d", argIndex)
		args = append(args, *req.MentionNotify)
		argIndex++
	}
	if req.RoomMessageNotify != nil {
		query += fmt.Sprintf(", room_message_notify = $%d", argIndex)
		args = append(args, *req.RoomMessageNotify)
		argIndex++
	}
	if req.MuteAll != nil {
		query += fmt.Sprintf(", mute_all = $%d", argIndex)
		args = append(args, *req.MuteAll)
		argIndex++
	}
	if req.QuietHoursEnabled != nil {
		query += fmt.Sprintf(", quiet_hours_enabled = $%d", argIndex)
		args = append(args, *req.QuietHoursEnabled)
		argIndex++
	}
	if req.QuietHoursStart != nil {
		query += fmt.Sprintf(", quiet_hours_start = $%d", argIndex)
		args = append(args, *req.QuietHoursStart)
		argIndex++
	}
	if req.QuietHoursEnd != nil {
		query += fmt.Sprintf(", quiet_hours_end = $%d", argIndex)
		args = append(args, *req.QuietHoursEnd)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE user_id = $%d", argIndex)
	args = append(args, userID)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return r.GetPreferences(userID)
}

// ============================================
// UNREAD COUNTS
// ============================================

func (r *NotificationRepository) GetUnreadCounts(userID int) (*models.UnreadCount, error) {
	counts := &models.UnreadCount{
		RoomUnread: make(map[int]int),
	}

	// Get notification unread count
	notifQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	r.db.QueryRow(notifQuery, userID).Scan(&counts.NotificationUnread)

	// Get direct message unread count
	dmQuery := `SELECT COUNT(*) FROM direct_messages WHERE receiver_id = $1 AND is_read = false`
	r.db.QueryRow(dmQuery, userID).Scan(&counts.DirectMessageUnread)

	// Get room message unread counts
	roomQuery := `
		SELECT m.room_id, COUNT(*)
		FROM messages m
		JOIN room_members rm ON m.room_id = rm.room_id AND rm.user_id = $1
		LEFT JOIN message_read_status mrs ON m.id = mrs.message_id AND mrs.user_id = $1
		WHERE mrs.id IS NULL AND m.sender_id != $1
		GROUP BY m.room_id`

	rows, err := r.db.Query(roomQuery, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var roomID, count int
			if err := rows.Scan(&roomID, &count); err == nil {
				counts.RoomUnread[roomID] = count
			}
		}
	}

	// Calculate total
	counts.TotalUnread = counts.NotificationUnread + counts.DirectMessageUnread
	for _, count := range counts.RoomUnread {
		counts.TotalUnread += count
	}

	return counts, nil
}

// ============================================
// BULK NOTIFICATION CREATION
// ============================================

func (r *NotificationRepository) CreateBulkNotifications(userIDs []int, notifType, title, body, data string, referenceID *int) error {
	if len(userIDs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO notifications (user_id, type, title, body, data, reference_id)
		VALUES ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, userID := range userIDs {
		_, err := stmt.Exec(userID, notifType, title, body, data, referenceID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Helper function to convert data to JSON string
func ToJSONString(data interface{}) string {
	bytes, _ := json.Marshal(data)
	return string(bytes)
}
