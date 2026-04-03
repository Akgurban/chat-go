package websocket

import (
	"encoding/json"
	"log"
	"time"

	"chat-go/internal/models"
	"chat-go/internal/repository"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	UserID   int
	Username string
	Rooms    map[int]bool
	send     chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn, userID int, username string) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		UserID:   userID,
		Username: username,
		Rooms:    make(map[int]bool),
		send:     make(chan []byte, 256),
	}
}

type IncomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type JoinRoomPayload struct {
	RoomID int `json:"room_id"`
}

type LeaveRoomPayload struct {
	RoomID int `json:"room_id"`
}

type ChatMessagePayload struct {
	RoomID  int    `json:"room_id"`
	Content string `json:"content"`
}

type DirectMessagePayloadIncoming struct {
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
}

func (c *Client) ReadPump(messageRepo *repository.MessageRepository) {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(message, &incoming); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		c.handleMessage(incoming, messageRepo)
	}
}

func (c *Client) handleMessage(msg IncomingMessage, messageRepo *repository.MessageRepository) {
	switch msg.Type {
	case "join_room":
		var payload JoinRoomPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		c.Hub.JoinRoom(c, payload.RoomID)

		// Notify room that user joined
		notification := models.WSMessage{
			Type: "user_joined",
			Payload: map[string]interface{}{
				"room_id":  payload.RoomID,
				"user_id":  c.UserID,
				"username": c.Username,
			},
		}
		data, _ := json.Marshal(notification)
		c.Hub.BroadcastToRoom(payload.RoomID, data)

	case "leave_room":
		var payload LeaveRoomPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		c.Hub.LeaveRoom(c, payload.RoomID)

		// Notify room that user left
		notification := models.WSMessage{
			Type: "user_left",
			Payload: map[string]interface{}{
				"room_id":  payload.RoomID,
				"user_id":  c.UserID,
				"username": c.Username,
			},
		}
		data, _ := json.Marshal(notification)
		c.Hub.BroadcastToRoom(payload.RoomID, data)

	case "chat_message":
		var payload ChatMessagePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		// Check if user is in the room
		if !c.Rooms[payload.RoomID] {
			return
		}

		// Save message to database
		dbMessage := &models.Message{
			RoomID:      payload.RoomID,
			SenderID:    &c.UserID,
			Content:     payload.Content,
			MessageType: "text",
		}
		if err := messageRepo.CreateMessage(dbMessage); err != nil {
			log.Printf("Failed to save message: %v", err)
			return
		}

		// Broadcast to room
		broadcastMsg := models.WSMessage{
			Type: "new_message",
			Payload: map[string]interface{}{
				"id":              dbMessage.ID,
				"room_id":         dbMessage.RoomID,
				"sender_id":       c.UserID,
				"sender_username": c.Username,
				"content":         dbMessage.Content,
				"created_at":      dbMessage.CreatedAt,
			},
		}
		data, _ := json.Marshal(broadcastMsg)
		c.Hub.BroadcastToRoom(payload.RoomID, data)

	case "direct_message":
		var payload DirectMessagePayloadIncoming
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		// Save direct message to database
		dm := &models.DirectMessage{
			SenderID:    &c.UserID,
			ReceiverID:  &payload.ReceiverID,
			Content:     payload.Content,
			MessageType: "text",
		}
		if err := messageRepo.CreateDirectMessage(dm); err != nil {
			log.Printf("Failed to save direct message: %v", err)
			return
		}

		// Send to receiver
		dmMsg := models.WSMessage{
			Type: "new_direct_message",
			Payload: map[string]interface{}{
				"id":              dm.ID,
				"sender_id":       c.UserID,
				"receiver_id":     payload.ReceiverID,
				"sender_username": c.Username,
				"content":         dm.Content,
				"created_at":      dm.CreatedAt,
			},
		}
		data, _ := json.Marshal(dmMsg)
		c.Hub.SendDirectMessage(payload.ReceiverID, data)

		// Also send back to sender for confirmation
		c.send <- data

	case "typing":
		var payload ChatMessagePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		typingMsg := models.WSMessage{
			Type: "user_typing",
			Payload: map[string]interface{}{
				"room_id":  payload.RoomID,
				"user_id":  c.UserID,
				"username": c.Username,
			},
		}
		data, _ := json.Marshal(typingMsg)
		c.Hub.BroadcastToRoom(payload.RoomID, data)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
