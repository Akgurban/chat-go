package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	presenceKeyPrefix = "user:online:"
	presenceTTL       = 60 * time.Second // User must heartbeat every 60 seconds
)

// PresenceCache handles user online/offline status
type PresenceCache struct {
	client *redis.Client
}

// NewPresenceCache creates a new presence cache
func NewPresenceCache(redisClient *RedisClient) *PresenceCache {
	return &PresenceCache{client: redisClient.Client()}
}

// SetOnline marks a user as online with auto-expiration
// This should be called on WebSocket connect and periodically as heartbeat
func (p *PresenceCache) SetOnline(ctx context.Context, userID int) error {
	key := fmt.Sprintf("%s%d", presenceKeyPrefix, userID)
	return p.client.Set(ctx, key, time.Now().Unix(), presenceTTL).Err()
}

// SetOffline removes a user's online status
// This should be called on WebSocket disconnect
func (p *PresenceCache) SetOffline(ctx context.Context, userID int) error {
	key := fmt.Sprintf("%s%d", presenceKeyPrefix, userID)
	return p.client.Del(ctx, key).Err()
}

// IsOnline checks if a user is currently online
func (p *PresenceCache) IsOnline(ctx context.Context, userID int) (bool, error) {
	key := fmt.Sprintf("%s%d", presenceKeyPrefix, userID)
	exists, err := p.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// GetOnlineUsers returns a map of user IDs to their online status
func (p *PresenceCache) GetOnlineUsers(ctx context.Context, userIDs []int) (map[int]bool, error) {
	if len(userIDs) == 0 {
		return make(map[int]bool), nil
	}

	pipe := p.client.Pipeline()
	cmds := make(map[int]*redis.IntCmd)

	for _, id := range userIDs {
		key := fmt.Sprintf("%s%d", presenceKeyPrefix, id)
		cmds[id] = pipe.Exists(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	result := make(map[int]bool)
	for id, cmd := range cmds {
		result[id] = cmd.Val() == 1
	}
	return result, nil
}

// GetAllOnlineUserIDs returns all online user IDs
// Note: Use with caution in large systems, consider using SCAN instead
func (p *PresenceCache) GetAllOnlineUserIDs(ctx context.Context) ([]int, error) {
	pattern := fmt.Sprintf("%s*", presenceKeyPrefix)
	keys, err := p.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	userIDs := make([]int, 0, len(keys))
	prefixLen := len(presenceKeyPrefix)

	for _, key := range keys {
		if len(key) > prefixLen {
			if id, err := strconv.Atoi(key[prefixLen:]); err == nil {
				userIDs = append(userIDs, id)
			}
		}
	}
	return userIDs, nil
}

// RefreshPresence extends the TTL for a user's online status (heartbeat)
func (p *PresenceCache) RefreshPresence(ctx context.Context, userID int) error {
	key := fmt.Sprintf("%s%d", presenceKeyPrefix, userID)
	return p.client.Expire(ctx, key, presenceTTL).Err()
}
