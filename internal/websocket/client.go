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
	send     chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn, userID int, username string) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		UserID:   userID,
		Username: username,
		send:     make(chan []byte, 256),
	}
}

type IncomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type DirectMessagePayloadIncoming struct {
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
}

type EditDirectMessagePayload struct {
	MessageID  int    `json:"message_id"`
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
}

type DeleteDirectMessagePayload struct {
	MessageID  int `json:"message_id"`
	ReceiverID int `json:"receiver_id"`
}

type MarkReadPayload struct {
	SenderID int `json:"sender_id"` // For DMs - who sent the messages
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

		// Send same message back to sender so it appears in their chat
		c.send <- data

	case "edit_direct_message":
		var payload EditDirectMessagePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		// Edit message in database
		editedMsg, err := messageRepo.EditDirectMessage(payload.MessageID, c.UserID, payload.Content)
		if err != nil {
			log.Printf("Failed to edit direct message: %v", err)
			return
		}

		// Send edit notification to receiver
		editNotif := models.WSMessage{
			Type: "direct_message_edited",
			Payload: map[string]interface{}{
				"message_id":      editedMsg.ID,
				"sender_id":       c.UserID,
				"receiver_id":     payload.ReceiverID,
				"content":         editedMsg.Content,
				"edited_at":       editedMsg.EditedAt,
				"edited_by":       c.UserID,
				"editor_username": c.Username,
			},
		}
		data, _ := json.Marshal(editNotif)
		c.Hub.SendDirectMessage(payload.ReceiverID, data)

		// Send same notification back to sender
		c.send <- data

	case "delete_direct_message":
		var payload DeleteDirectMessagePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		// Delete message in database
		if err := messageRepo.DeleteDirectMessage(payload.MessageID, c.UserID); err != nil {
			log.Printf("Failed to delete direct message: %v", err)
			return
		}

		// Send delete notification to receiver
		deleteNotif := models.WSMessage{
			Type: "direct_message_deleted",
			Payload: map[string]interface{}{
				"message_id":       payload.MessageID,
				"sender_id":        c.UserID,
				"receiver_id":      payload.ReceiverID,
				"deleted_by":       c.UserID,
				"deleter_username": c.Username,
			},
		}
		data, _ := json.Marshal(deleteNotif)
		c.Hub.SendDirectMessage(payload.ReceiverID, data)

		// Send same notification back to sender
		c.send <- data

	case "mark_read":
		var payload MarkReadPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		if payload.SenderID > 0 {
			// Mark direct messages as read
			messageRepo.MarkDirectMessagesAsRead(payload.SenderID, c.UserID)

			// Notify sender that messages were read
			readReceipt := models.WSMessage{
				Type: "messages_read",
				Payload: map[string]interface{}{
					"reader_id":       c.UserID,
					"reader_username": c.Username,
					"sender_id":       payload.SenderID,
				},
			}
			data, _ := json.Marshal(readReceipt)
			c.Hub.SendDirectMessage(payload.SenderID, data)
		}

	case "typing_dm":
		var payload DirectMessagePayloadIncoming
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		typingMsg := models.WSMessage{
			Type: "user_typing_dm",
			Payload: map[string]interface{}{
				"user_id":  c.UserID,
				"username": c.Username,
			},
		}
		data, _ := json.Marshal(typingMsg)
		c.Hub.SendDirectMessage(payload.ReceiverID, data)
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
