package cache

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	// Key patterns for unread counts
	// Format: unread:{user_id}:{chat_type}:{chat_id}
	unreadKeyPrefix = "unread:"
)

// UnreadCache handles unread message counts
type UnreadCache struct {
	client *redis.Client
}

// NewUnreadCache creates a new unread count cache
func NewUnreadCache(redisClient *RedisClient) *UnreadCache {
	return &UnreadCache{client: redisClient.Client()}
}

// buildUnreadKey creates the Redis key for unread counts
func buildUnreadKey(userID int, chatType string, chatID int) string {
	return fmt.Sprintf("%s%d:%s:%d", unreadKeyPrefix, userID, chatType, chatID)
}

// IncrementUnread increases the unread count for a user in a specific chat
func (u *UnreadCache) IncrementUnread(ctx context.Context, userID int, chatType string, chatID int) error {
	key := buildUnreadKey(userID, chatType, chatID)
	return u.client.Incr(ctx, key).Err()
}

// IncrementUnreadForUsers increases unread count for multiple users (e.g., all room members)
func (u *UnreadCache) IncrementUnreadForUsers(ctx context.Context, userIDs []int, chatType string, chatID int, excludeUserID int) error {
	if len(userIDs) == 0 {
		return nil
	}

	pipe := u.client.Pipeline()
	for _, userID := range userIDs {
		if userID == excludeUserID {
			continue // Don't increment for the sender
		}
		key := buildUnreadKey(userID, chatType, chatID)
		pipe.Incr(ctx, key)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// MarkAsRead resets the unread count when a user reads messages
func (u *UnreadCache) MarkAsRead(ctx context.Context, userID int, chatType string, chatID int) error {
	key := buildUnreadKey(userID, chatType, chatID)
	return u.client.Del(ctx, key).Err()
}

// GetUnreadCount returns the unread count for a specific chat
func (u *UnreadCache) GetUnreadCount(ctx context.Context, userID int, chatType string, chatID int) (int64, error) {
	key := buildUnreadKey(userID, chatType, chatID)
	val, err := u.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// GetTotalUnread returns the total unread count across all chats for a user
func (u *UnreadCache) GetTotalUnread(ctx context.Context, userID int) (int64, error) {
	pattern := fmt.Sprintf("%s%d:*", unreadKeyPrefix, userID)
	keys, err := u.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, err
	}

	if len(keys) == 0 {
		return 0, nil
	}

	var total int64
	pipe := u.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, err
	}

	for _, cmd := range cmds {
		val, err := cmd.Int64()
		if err == nil {
			total += val
		}
	}

	return total, nil
}

// GetAllUnreadCounts returns unread counts for all chats of a user
func (u *UnreadCache) GetAllUnreadCounts(ctx context.Context, userID int) (map[string]map[int]int64, error) {
	pattern := fmt.Sprintf("%s%d:*", unreadKeyPrefix, userID)
	keys, err := u.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	// Result: map[chatType]map[chatID]count
	result := map[string]map[int]int64{
		"room":   make(map[int]int64),
		"direct": make(map[int]int64),
	}

	if len(keys) == 0 {
		return result, nil
	}

	pipe := u.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	prefixLen := len(fmt.Sprintf("%s%d:", unreadKeyPrefix, userID))
	for i, key := range keys {
		// Parse key: unread:{userID}:{chatType}:{chatID}
		suffix := key[prefixLen:] // chatType:chatID
		var chatType string
		var chatID int
		_, err := fmt.Sscanf(suffix, "%[^:]:%d", &chatType, &chatID)
		if err != nil {
			continue
		}

		val, err := cmds[i].Int64()
		if err == nil && val > 0 {
			result[chatType][chatID] = val
		}
	}

	return result, nil
}

// SetUnreadCount sets the unread count directly (useful for syncing from DB)
func (u *UnreadCache) SetUnreadCount(ctx context.Context, userID int, chatType string, chatID int, count int64) error {
	key := buildUnreadKey(userID, chatType, chatID)
	if count <= 0 {
		return u.client.Del(ctx, key).Err()
	}
	return u.client.Set(ctx, key, strconv.FormatInt(count, 10), 0).Err()
}
