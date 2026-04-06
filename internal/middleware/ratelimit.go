package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"chat-go/internal/cache"
)

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(appCache *cache.Cache, config cache.RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting if cache is not available
			if appCache == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Get identifier - prefer user ID from context, fallback to IP
			var identifier string
			if userID := r.Context().Value("user_id"); userID != nil {
				identifier = fmt.Sprintf("user:%d", userID.(int))
			} else {
				identifier = fmt.Sprintf("ip:%s", getClientIP(r))
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			allowed, remaining, resetDuration, err := appCache.Rate.IsAllowed(ctx, identifier, config)
			if err != nil {
				// On error, allow the request but log it
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(resetDuration).Unix()))

			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(resetDuration.Seconds())))
				http.Error(w, `{"error": "Rate limit exceeded. Please try again later."}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fallback to RemoteAddr
	return r.RemoteAddr
}

// APIRateLimitMiddleware applies default API rate limiting
func APIRateLimitMiddleware(appCache *cache.Cache) func(http.Handler) http.Handler {
	return RateLimitMiddleware(appCache, cache.RateLimitAPI)
}

// MessageRateLimitMiddleware applies message-specific rate limiting
func MessageRateLimitMiddleware(appCache *cache.Cache) func(http.Handler) http.Handler {
	return RateLimitMiddleware(appCache, cache.RateLimitMessage)
}

// LoginRateLimitMiddleware applies login-specific rate limiting
func LoginRateLimitMiddleware(appCache *cache.Cache) func(http.Handler) http.Handler {
	return RateLimitMiddleware(appCache, cache.RateLimitLogin)
}
