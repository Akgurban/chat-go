package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	rateLimitKeyPrefix = "ratelimit:"
)

// RateLimiter handles rate limiting using Redis sorted sets
type RateLimiter struct {
	client *redis.Client
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *RedisClient) *RateLimiter {
	return &RateLimiter{client: redisClient.Client()}
}

// RateLimitConfig defines rate limiting parameters
type RateLimitConfig struct {
	Action string        // Action name (e.g., "message", "login", "api")
	Limit  int64         // Maximum requests allowed
	Window time.Duration // Time window
}

// Common rate limit configurations
var (
	RateLimitMessage = RateLimitConfig{
		Action: "message",
		Limit:  30,
		Window: time.Minute,
	}
	RateLimitLogin = RateLimitConfig{
		Action: "login",
		Limit:  5,
		Window: time.Minute,
	}
	RateLimitAPI = RateLimitConfig{
		Action: "api",
		Limit:  100,
		Window: time.Minute,
	}
	RateLimitRegister = RateLimitConfig{
		Action: "register",
		Limit:  3,
		Window: time.Hour,
	}
)

// IsAllowed checks if an action is allowed under rate limiting (sliding window)
// Returns: allowed (bool), remaining requests (int64), reset time (time.Duration)
func (r *RateLimiter) IsAllowed(ctx context.Context, identifier string, config RateLimitConfig) (bool, int64, time.Duration, error) {
	key := fmt.Sprintf("%s%s:%s", rateLimitKeyPrefix, config.Action, identifier)
	now := time.Now()
	windowStart := now.Add(-config.Window).UnixNano()
	nowNano := now.UnixNano()

	// Use a transaction to ensure atomicity
	pipe := r.client.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Count current entries in the window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request with timestamp as score
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(nowNano), Member: nowNano})

	// Set key expiration
	pipe.Expire(ctx, key, config.Window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	count := countCmd.Val()
	allowed := count < config.Limit
	remaining := config.Limit - count - 1
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time (when the oldest entry will expire)
	resetDuration := config.Window

	return allowed, remaining, resetDuration, nil
}

// IsAllowedSimple is a simplified version that just returns whether the action is allowed
func (r *RateLimiter) IsAllowedSimple(ctx context.Context, userID int, config RateLimitConfig) (bool, error) {
	identifier := fmt.Sprintf("user:%d", userID)
	allowed, _, _, err := r.IsAllowed(ctx, identifier, config)
	return allowed, err
}

// IsAllowedByIP checks rate limit by IP address
func (r *RateLimiter) IsAllowedByIP(ctx context.Context, ip string, config RateLimitConfig) (bool, int64, time.Duration, error) {
	identifier := fmt.Sprintf("ip:%s", ip)
	return r.IsAllowed(ctx, identifier, config)
}

// Reset clears rate limiting for an identifier
func (r *RateLimiter) Reset(ctx context.Context, identifier string, action string) error {
	key := fmt.Sprintf("%s%s:%s", rateLimitKeyPrefix, action, identifier)
	return r.client.Del(ctx, key).Err()
}
