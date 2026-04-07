package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

const (
	// Channel patterns for pub/sub
	channelDirect = "chat:direct:%d:%d"
	channelUser   = "user:%d:notifications"
)

// PubSubManager handles Redis Pub/Sub for real-time messaging across server instances
type PubSubManager struct {
	client *redis.Client
}

// NewPubSubManager creates a new pub/sub manager
func NewPubSubManager(redisClient *RedisClient) *PubSubManager {
	return &PubSubManager{client: redisClient.Client()}
}

// MessageEvent represents a message published through Redis
type MessageEvent struct {
	Type       string          `json:"type"`                  // "new_message", "edit_message", "delete_message", "typing"
	ChatType   string          `json:"chat_type"`             // "direct"
	ChatID     int             `json:"chat_id"`               // conversation partner user_id
	SenderID   int             `json:"sender_id"`             // Who sent/triggered the event
	ReceiverID int             `json:"receiver_id,omitempty"` // For direct messages
	Data       json.RawMessage `json:"data"`                  // Event-specific data
}

// PresenceEvent represents a user presence change
type PresenceEvent struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Status   string `json:"status"` // "online" or "offline"
}

// NotificationEvent represents a notification to be sent to a user
type NotificationEvent struct {
	UserID  int             `json:"user_id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// PublishDirectMessage publishes a message event to a direct message channel
// Channel is sorted by lower user ID first to ensure both parties use the same channel
func (p *PubSubManager) PublishDirectMessage(ctx context.Context, userID1, userID2 int, event MessageEvent) error {
	// Ensure consistent channel naming
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}
	channel := fmt.Sprintf(channelDirect, userID1, userID2)
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, channel, data).Err()
}

// PublishUserNotification publishes a notification to a user's channel
func (p *PubSubManager) PublishUserNotification(ctx context.Context, userID int, notification NotificationEvent) error {
	channel := fmt.Sprintf(channelUser, userID)
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, channel, data).Err()
}

// PublishPresenceChange publishes a presence change event
func (p *PubSubManager) PublishPresenceChange(ctx context.Context, event PresenceEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, "presence:changes", data).Err()
}

// MessageHandler is a function that handles received messages
type MessageHandler func(channel string, payload []byte)

// SubscribeToDirectMessages subscribes to a direct message channel between two users
func (p *PubSubManager) SubscribeToDirectMessages(ctx context.Context, userID1, userID2 int, handler MessageHandler) {
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}
	channel := fmt.Sprintf(channelDirect, userID1, userID2)
	p.subscribe(ctx, channel, handler)
}

// SubscribeToUserNotifications subscribes to a user's notification channel
func (p *PubSubManager) SubscribeToUserNotifications(ctx context.Context, userID int, handler MessageHandler) {
	channel := fmt.Sprintf(channelUser, userID)
	p.subscribe(ctx, channel, handler)
}

// SubscribeToPresenceChanges subscribes to presence change events
func (p *PubSubManager) SubscribeToPresenceChanges(ctx context.Context, handler MessageHandler) {
	p.subscribe(ctx, "presence:changes", handler)
}

// subscribe is a helper to handle subscription logic
func (p *PubSubManager) subscribe(ctx context.Context, channel string, handler MessageHandler) {
	pubsub := p.client.Subscribe(ctx, channel)

	go func() {
		defer pubsub.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return // Context cancelled
					}
					log.Printf("PubSub receive error on %s: %v", channel, err)
					continue
				}
				handler(msg.Channel, []byte(msg.Payload))
			}
		}
	}()
}

// SubscribeMultiple subscribes to multiple channels
func (p *PubSubManager) SubscribeMultiple(ctx context.Context, channels []string, handler MessageHandler) *redis.PubSub {
	pubsub := p.client.Subscribe(ctx, channels...)

	go func() {
		for {
			select {
			case <-ctx.Done():
				pubsub.Close()
				return
			default:
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					log.Printf("PubSub receive error: %v", err)
					continue
				}
				handler(msg.Channel, []byte(msg.Payload))
			}
		}
	}()

	return pubsub
}
