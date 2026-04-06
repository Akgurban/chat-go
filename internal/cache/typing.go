package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	typingKeyPrefix = "typing:"
	typingTTL       = 3 * time.Second // Typing indicator expires after 3 seconds
)

// TypingCache handles typing indicators for chats
type TypingCache struct {
	client *redis.Client
}

// NewTypingCache creates a new typing cache
func NewTypingCache(redisClient *RedisClient) *TypingCache {
	return &TypingCache{client: redisClient.Client()}
}

// buildTypingKey creates the Redis key for typing indicators
func buildTypingKey(chatType string, chatID int) string {
	return fmt.Sprintf("%s%s:%d", typingKeyPrefix, chatType, chatID)
}

// SetTyping marks a user as typing in a chat (expires in 3 seconds)
func (t *TypingCache) SetTyping(ctx context.Context, chatType string, chatID int, userID int, username string) error {
	key := buildTypingKey(chatType, chatID)
	field := fmt.Sprintf("%d", userID)
	value := fmt.Sprintf("%s:%d", username, time.Now().Unix())

	// Use HSET with separate EXPIRE
	pipe := t.client.Pipeline()
	pipe.HSet(ctx, key, field, value)
	pipe.Expire(ctx, key, 10*time.Second) // Key expires if no activity
	_, err := pipe.Exec(ctx)
	return err
}

// ClearTyping removes a user's typing indicator
func (t *TypingCache) ClearTyping(ctx context.Context, chatType string, chatID int, userID int) error {
	key := buildTypingKey(chatType, chatID)
	field := fmt.Sprintf("%d", userID)
	return t.client.HDel(ctx, key, field).Err()
}

// TypingUser represents a user who is currently typing
type TypingUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

// GetTypingUsers returns users currently typing in a chat
func (t *TypingCache) GetTypingUsers(ctx context.Context, chatType string, chatID int) ([]TypingUser, error) {
	key := buildTypingKey(chatType, chatID)
	result, err := t.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	threshold := time.Now().Add(-typingTTL).Unix()
	var users []TypingUser
	var expiredFields []string

	for userIDStr, value := range result {
		var username string
		var timestamp int64
		_, err := fmt.Sscanf(value, "%[^:]:%d", &username, &timestamp)
		if err != nil {
			continue
		}

		// Check if typing indicator has expired
		if timestamp < threshold {
			expiredFields = append(expiredFields, userIDStr)
			continue
		}

		var userID int
		fmt.Sscanf(userIDStr, "%d", &userID)
		users = append(users, TypingUser{
			UserID:   userID,
			Username: username,
		})
	}

	// Clean up expired entries
	if len(expiredFields) > 0 {
		t.client.HDel(ctx, key, expiredFields...)
	}

	return users, nil
}
