package cache

import (
	"context"
)

// Cache combines all cache types for easy initialization and access
type Cache struct {
	redis    *RedisClient
	Presence *PresenceCache
	Unread   *UnreadCache
	Typing   *TypingCache
	Session  *SessionCache
	Rate     *RateLimiter
	PubSub   *PubSubManager
}

// NewCache creates a new cache instance with all cache types
func NewCache(redisClient *RedisClient) *Cache {
	return &Cache{
		redis:    redisClient,
		Presence: NewPresenceCache(redisClient),
		Unread:   NewUnreadCache(redisClient),
		Typing:   NewTypingCache(redisClient),
		Session:  NewSessionCache(redisClient),
		Rate:     NewRateLimiter(redisClient),
		PubSub:   NewPubSubManager(redisClient),
	}
}

// Close closes the Redis connection
func (c *Cache) Close() error {
	return c.redis.Close()
}

// Ping checks if Redis is available
func (c *Cache) Ping(ctx context.Context) error {
	return c.redis.Client().Ping(ctx).Err()
}

// FlushAll clears all data (use with caution, mainly for testing)
func (c *Cache) FlushAll(ctx context.Context) error {
	return c.redis.Client().FlushAll(ctx).Err()
}
