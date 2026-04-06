package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	sessionKeyPrefix = "session:"
)

// SessionCache handles session/JWT token management
type SessionCache struct {
	client *redis.Client
}

// NewSessionCache creates a new session cache
func NewSessionCache(redisClient *RedisClient) *SessionCache {
	return &SessionCache{client: redisClient.Client()}
}

// StoreSession stores a session with an expiration time
func (s *SessionCache) StoreSession(ctx context.Context, userID int, tokenID string, expiration time.Duration) error {
	key := fmt.Sprintf("%s%d:%s", sessionKeyPrefix, userID, tokenID)
	return s.client.Set(ctx, key, "active", expiration).Err()
}

// IsSessionValid checks if a session exists and is valid
func (s *SessionCache) IsSessionValid(ctx context.Context, userID int, tokenID string) (bool, error) {
	key := fmt.Sprintf("%s%d:%s", sessionKeyPrefix, userID, tokenID)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// RevokeSession revokes a specific session
func (s *SessionCache) RevokeSession(ctx context.Context, userID int, tokenID string) error {
	key := fmt.Sprintf("%s%d:%s", sessionKeyPrefix, userID, tokenID)
	return s.client.Del(ctx, key).Err()
}

// RevokeAllSessions revokes all sessions for a user (logout everywhere)
func (s *SessionCache) RevokeAllSessions(ctx context.Context, userID int) error {
	pattern := fmt.Sprintf("%s%d:*", sessionKeyPrefix, userID)
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...).Err()
}

// RefreshSession extends the TTL of a session
func (s *SessionCache) RefreshSession(ctx context.Context, userID int, tokenID string, expiration time.Duration) error {
	key := fmt.Sprintf("%s%d:%s", sessionKeyPrefix, userID, tokenID)
	return s.client.Expire(ctx, key, expiration).Err()
}

// GetActiveSessions returns the number of active sessions for a user
func (s *SessionCache) GetActiveSessions(ctx context.Context, userID int) (int64, error) {
	pattern := fmt.Sprintf("%s%d:*", sessionKeyPrefix, userID)
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, err
	}
	return int64(len(keys)), nil
}

// AddToBlacklist adds a token to the blacklist (for logout before expiry)
func (s *SessionCache) AddToBlacklist(ctx context.Context, tokenID string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	return s.client.Set(ctx, key, "revoked", expiration).Err()
}

// IsBlacklisted checks if a token is blacklisted
func (s *SessionCache) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}
