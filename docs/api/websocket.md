# WebSocket API

Real-time messaging via WebSocket connection.

---

## Connection

### Endpoint

```
ws://localhost:8080/ws?token=<JWT_TOKEN>
```

Or with header:

```
ws://localhost:8080/ws
Authorization: Bearer <JWT_TOKEN>
```

### Connection Example (JavaScript)

```javascript
const token = "your-jwt-token";
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onopen = () => {
  console.log("Connected to chat server");
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log("Received:", message);
};

ws.onerror = (error) => {
  console.error("WebSocket error:", error);
};

ws.onclose = () => {
  console.log("Disconnected from chat server");
};
```

---

## Client → Server Messages

All messages follow this format:

```json
{
  "type": "message_type",
  "payload": { ... }
}
```

### Join Room

Join a chat room to receive messages.

```json
{
  "type": "join_room",
  "payload": {
    "room_id": 1
  }
}
```

| Field   | Type    | Description     |
| ------- | ------- | --------------- |
| room_id | integer | Room ID to join |

**Example:**

```javascript
ws.send(
  JSON.stringify({
    type: "join_room",
    payload: { room_id: 1 },
  }),
);
```

---

### Leave Room

Leave a chat room.

```json
{
  "type": "leave_room",
  "payload": {
    "room_id": 1
  }
}
```

| Field   | Type    | Description      |
| ------- | ------- | ---------------- |
| room_id | integer | Room ID to leave |

**Example:**

```javascript
ws.send(
  JSON.stringify({
    type: "leave_room",
    payload: { room_id: 1 },
  }),
);
```

---

### Send Chat Message

Send a message to a room (must be joined first).

```json
{
  "type": "chat_message",
  "payload": {
    "room_id": 1,
    "content": "Hello everyone!"
  }
}
```

| Field   | Type    | Description     |
| ------- | ------- | --------------- |
| room_id | integer | Target room ID  |
| content | string  | Message content |

**Example:**

```javascript
ws.send(
  JSON.stringify({
    type: "chat_message",
    payload: {
      room_id: 1,
      content: "Hello everyone!",
    },
  }),
);
```

---

### Send Direct Message

Send a private message to another user.

```json
{
  "type": "direct_message",
  "payload": {
    "receiver_id": 2,
    "content": "Hi there!"
  }
}
```

| Field       | Type    | Description       |
| ----------- | ------- | ----------------- |
| receiver_id | integer | Recipient user ID |
| content     | string  | Message content   |

**Example:**

```javascript
ws.send(
  JSON.stringify({
    type: "direct_message",
    payload: {
      receiver_id: 2,
      content: "Hi there!",
    },
  }),
);
```

---

### Typing Indicator

Notify room that user is typing.

```json
{
  "type": "typing",
  "payload": {
    "room_id": 1
  }
}
```

| Field   | Type    | Description                  |
| ------- | ------- | ---------------------------- |
| room_id | integer | Room ID where user is typing |

**Example:**

```javascript
ws.send(
  JSON.stringify({
    type: "typing",
    payload: { room_id: 1 },
  }),
);
```

---

### Edit Message

Edit a previously sent message in a room.

```json
{
  "type": "edit_message",
  "payload": {
    "message_id": 123,
    "room_id": 1,
    "content": "Updated message content"
  }
}
```

| Field      | Type    | Description                 |
| ---------- | ------- | --------------------------- |
| message_id | integer | Message ID to edit          |
| room_id    | integer | Room containing the message |
| content    | string  | New message content         |

**Example:**

```javascript
ws.send(
  JSON.stringify({
    type: "edit_message",
    payload: {
      message_id: 123,
      room_id: 1,
      content: "Updated message",
    },
  }),
);
```

---

### Delete Message

Delete a previously sent message in a room.

```json
{
  "type": "delete_message",
  "payload": {
    "message_id": 123,
    "room_id": 1
  }
}
```

| Field      | Type    | Description                 |
| ---------- | ------- | --------------------------- |
| message_id | integer | Message ID to delete        |
| room_id    | integer | Room containing the message |

**Example:**

```javascript
ws.send(
  JSON.stringify({
    type: "delete_message",
    payload: {
      message_id: 123,
      room_id: 1,
    },
  }),
);
```

---

### Mark Messages as Read

Mark messages as read in a room or DM conversation.

```json
{
  "type": "mark_read",
  "payload": {
    "room_id": 1,
    "message_id": 100
  }
}
```

For room messages:
| Field | Type | Description |
| ---------- | ------- | -------------------------------- |
| room_id | integer | Room ID |
| message_id | integer | Mark all messages up to this ID |

For direct messages:

```json
{
  "type": "mark_read",
  "payload": {
    "sender_id": 2
  }
}
```

| Field     | Type    | Description                         |
| --------- | ------- | ----------------------------------- |
| sender_id | integer | Mark all DMs from this user as read |

**Example:**

```javascript
// Mark room messages as read
ws.send(
  JSON.stringify({
    type: "mark_read",
    payload: { room_id: 1, message_id: 100 },
  }),
);

// Mark direct messages as read
ws.send(
  JSON.stringify({
    type: "mark_read",
    payload: { sender_id: 2 },
  }),
);
```

---

### Typing Indicator (Direct Message)

Notify a user that you are typing a direct message.

```json
{
  "type": "typing_dm",
  "payload": {
    "receiver_id": 2
  }
}
```

| Field       | Type    | Description           |
| ----------- | ------- | --------------------- |
| receiver_id | integer | User you're typing to |

**Example:**

```javascript
ws.send(
  JSON.stringify({
    type: "typing_dm",
    payload: { receiver_id: 2 },
  }),
);
```

---

## Server → Client Messages

### New Message

Received when a new message is sent in a joined room.

```json
{
  "type": "new_message",
  "payload": {
    "id": 123,
    "room_id": 1,
    "sender_id": 2,
    "sender_username": "jane",
    "content": "Hello!",
    "created_at": "2026-04-03T12:00:00Z"
  }
}
```

---

### Message Edited

Received when a message is edited in a joined room.

```json
{
  "type": "message_edited",
  "payload": {
    "message_id": 123,
    "room_id": 1,
    "content": "Updated content",
    "edited_at": "2026-04-03T12:05:00Z",
    "edited_by": 2,
    "editor_username": "jane"
  }
}
```

---

### Message Deleted

Received when a message is deleted in a joined room.

```json
{
  "type": "message_deleted",
  "payload": {
    "message_id": 123,
    "room_id": 1,
    "deleted_by": 2,
    "deleter_username": "jane"
  }
}
```

---

### Messages Read Receipt

Received when someone reads your direct messages.

```json
{
  "type": "messages_read",
  "payload": {
    "reader_id": 2,
    "reader_username": "jane",
    "sender_id": 1
  }
}
```

---

### Notification

Received when you get a new notification (while connected).

```json
{
  "type": "notification",
  "payload": {
    "id": 1,
    "user_id": 1,
    "type": "direct_message",
    "title": "Message from Jane",
    "body": "Hey, how are you?",
    "data": "{\"sender_id\": 2, \"message_id\": 456}",
    "is_read": false,
    "created_at": "2026-04-03T12:00:00Z"
  }
}
```

---

### User Typing (Direct Message)

Received when another user is typing a direct message to you.

```json
{
  "type": "user_typing_dm",
  "payload": {
    "user_id": 2,
    "username": "jane"
  }
}
```

---

### New Direct Message

Received when someone sends you a direct message.

```json
{
  "type": "new_direct_message",
  "payload": {
    "id": 456,
    "sender_id": 2,
    "sender_username": "jane",
    "content": "Hey, how are you?",
    "created_at": "2026-04-03T12:00:00Z"
  }
}
```

---

### User Joined

Received when a user joins a room you're in.

```json
{
  "type": "user_joined",
  "payload": {
    "room_id": 1,
    "user_id": 3,
    "username": "bob"
  }
}
```

---

### User Left

Received when a user leaves a room you're in.

```json
{
  "type": "user_left",
  "payload": {
    "room_id": 1,
    "user_id": 3,
    "username": "bob"
  }
}
```

---

### User Typing

Received when a user is typing in a room you're in.

```json
{
  "type": "user_typing",
  "payload": {
    "room_id": 1,
    "user_id": 2,
    "username": "jane"
  }
}
```

---

## Complete Client Example

```javascript
class ChatClient {
  constructor(token) {
    this.token = token;
    this.ws = null;
    this.handlers = {};
  }

  connect() {
    this.ws = new WebSocket(`ws://localhost:8080/ws?token=${this.token}`);

    this.ws.onopen = () => {
      console.log("Connected");
      this.emit("connected");
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.emit(message.type, message.payload);
    };

    this.ws.onclose = () => {
      console.log("Disconnected");
      this.emit("disconnected");
    };
  }

  on(event, handler) {
    if (!this.handlers[event]) {
      this.handlers[event] = [];
    }
    this.handlers[event].push(handler);
  }

  emit(event, data) {
    if (this.handlers[event]) {
      this.handlers[event].forEach((handler) => handler(data));
    }
  }

  send(type, payload) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, payload }));
    }
  }

  joinRoom(roomId) {
    this.send("join_room", { room_id: roomId });
  }

  leaveRoom(roomId) {
    this.send("leave_room", { room_id: roomId });
  }

  sendMessage(roomId, content) {
    this.send("chat_message", { room_id: roomId, content });
  }

  sendDirectMessage(receiverId, content) {
    this.send("direct_message", { receiver_id: receiverId, content });
  }

  sendTyping(roomId) {
    this.send("typing", { room_id: roomId });
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
    }
  }
}

// Usage
const chat = new ChatClient("your-jwt-token");

chat.on("connected", () => {
  chat.joinRoom(1);
});

chat.on("new_message", (message) => {
  console.log(`${message.sender_username}: ${message.content}`);
});

chat.on("user_joined", (data) => {
  console.log(`${data.username} joined the room`);
});

chat.on("user_typing", (data) => {
  console.log(`${data.username} is typing...`);
});

chat.connect();
```

---

## Connection Settings

| Setting          | Value      | Description                                |
| ---------------- | ---------- | ------------------------------------------ |
| Ping interval    | 54 seconds | Server sends ping to keep connection alive |
| Pong timeout     | 60 seconds | Connection closed if no pong received      |
| Max message size | 4096 bytes | Maximum size of a single message           |
| Write timeout    | 10 seconds | Timeout for writing messages               |
