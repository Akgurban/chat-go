# Messages API

Base URL: `http://localhost:8080/api`

> ⚠️ All endpoints require authentication. Include JWT token in Authorization header.

---

## Get Room Messages

Retrieve messages from a chat room.

### Request

```
GET /api/rooms/{id}/messages
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| id | integer | Room ID |

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| limit | integer | 50 | Max messages to return (1-100) |
| offset | integer | 0 | Number of messages to skip |

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |

### Response

**Success (200 OK):**

```json
[
  {
    "id": 1,
    "room_id": 1,
    "sender_id": 1,
    "content": "Hello everyone!",
    "message_type": "text",
    "created_at": "2026-04-03T12:00:00Z",
    "updated_at": "2026-04-03T12:00:00Z",
    "sender_username": "john",
    "sender_avatar": null
  },
  {
    "id": 2,
    "room_id": 1,
    "sender_id": 2,
    "content": "Hey John!",
    "message_type": "text",
    "created_at": "2026-04-03T12:01:00Z",
    "updated_at": "2026-04-03T12:01:00Z",
    "sender_username": "jane",
    "sender_avatar": "https://example.com/avatar.jpg"
  }
]
```

**Errors:**

| Status | Response                       |
| ------ | ------------------------------ |
| 400    | `{"error": "Invalid room ID"}` |
| 403    | `{"error": "Access denied"}`   |
| 404    | `{"error": "Room not found"}`  |

### Example

```bash
# Get latest 50 messages
curl -X GET http://localhost:8080/api/rooms/1/messages \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# With pagination
curl -X GET "http://localhost:8080/api/rooms/1/messages?limit=20&offset=0" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Send Room Message

Send a message to a chat room.

### Request

```
POST /api/rooms/{id}/messages
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| id | integer | Room ID |

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |
| Content-Type | application/json |

**Body:**

```json
{
  "content": "string",
  "message_type": "string"
}
```

| Field        | Type   | Required | Description                                    |
| ------------ | ------ | -------- | ---------------------------------------------- |
| content      | string | Yes      | Message content                                |
| message_type | string | No       | Type: `text`, `image`, `file`. Default: `text` |

### Response

**Success (201 Created):**

```json
{
  "id": 3,
  "room_id": 1,
  "sender_id": 1,
  "content": "Hello world!",
  "message_type": "text",
  "created_at": "2026-04-03T14:00:00Z",
  "updated_at": "2026-04-03T14:00:00Z"
}
```

**Errors:**

| Status | Response                                             |
| ------ | ---------------------------------------------------- |
| 400    | `{"error": "Invalid room ID"}`                       |
| 400    | `{"error": "Message content is required"}`           |
| 403    | `{"error": "You must be a member to send messages"}` |

### Example

```bash
curl -X POST http://localhost:8080/api/rooms/1/messages \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hello everyone!",
    "message_type": "text"
  }'
```

---

## Get Direct Messages

Retrieve direct messages with another user.

### Request

```
GET /api/dm/{userId}
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| userId | integer | Other user's ID |

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| limit | integer | 50 | Max messages to return (1-100) |
| offset | integer | 0 | Number of messages to skip |

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |

### Response

**Success (200 OK):**

```json
[
  {
    "id": 1,
    "sender_id": 1,
    "receiver_id": 2,
    "content": "Hi Jane!",
    "message_type": "text",
    "is_read": true,
    "created_at": "2026-04-03T12:00:00Z",
    "sender_username": "john",
    "receiver_username": "jane"
  },
  {
    "id": 2,
    "sender_id": 2,
    "receiver_id": 1,
    "content": "Hey John!",
    "message_type": "text",
    "is_read": true,
    "created_at": "2026-04-03T12:01:00Z",
    "sender_username": "jane",
    "receiver_username": "john"
  }
]
```

> Note: Messages are automatically marked as read when fetched.

**Errors:**

| Status | Response                       |
| ------ | ------------------------------ |
| 400    | `{"error": "Invalid user ID"}` |

### Example

```bash
# Get conversation with user ID 2
curl -X GET http://localhost:8080/api/dm/2 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# With pagination
curl -X GET "http://localhost:8080/api/dm/2?limit=20&offset=0" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Send Direct Message

Send a direct message to another user.

### Request

```
POST /api/dm/{userId}
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| userId | integer | Receiver's user ID |

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |
| Content-Type | application/json |

**Body:**

```json
{
  "content": "string",
  "message_type": "string"
}
```

| Field        | Type   | Required | Description                                    |
| ------------ | ------ | -------- | ---------------------------------------------- |
| content      | string | Yes      | Message content                                |
| message_type | string | No       | Type: `text`, `image`, `file`. Default: `text` |

### Response

**Success (201 Created):**

```json
{
  "id": 3,
  "sender_id": 1,
  "receiver_id": 2,
  "content": "How are you?",
  "message_type": "text",
  "is_read": false,
  "created_at": "2026-04-03T14:00:00Z"
}
```

**Errors:**

| Status | Response                                   |
| ------ | ------------------------------------------ |
| 400    | `{"error": "Invalid user ID"}`             |
| 400    | `{"error": "Message content is required"}` |

### Example

```bash
curl -X POST http://localhost:8080/api/dm/2 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hey, how are you?",
    "message_type": "text"
  }'
```

---

## Message Objects

### Room Message

| Field           | Type            | Description                 |
| --------------- | --------------- | --------------------------- |
| id              | integer         | Unique message identifier   |
| room_id         | integer         | Room the message belongs to |
| sender_id       | integer \| null | Sender's user ID            |
| content         | string          | Message content             |
| message_type    | string          | `text`, `image`, `file`     |
| created_at      | string          | ISO 8601 timestamp          |
| updated_at      | string          | ISO 8601 timestamp          |
| sender_username | string          | Sender's username           |
| sender_avatar   | string \| null  | Sender's avatar URL         |

### Direct Message

| Field             | Type            | Description                   |
| ----------------- | --------------- | ----------------------------- |
| id                | integer         | Unique message identifier     |
| sender_id         | integer \| null | Sender's user ID              |
| receiver_id       | integer \| null | Receiver's user ID            |
| content           | string          | Message content               |
| message_type      | string          | `text`, `image`, `file`       |
| is_read           | boolean         | Whether message has been read |
| is_edited         | boolean         | Whether message was edited    |
| edited_at         | string \| null  | ISO 8601 timestamp of edit    |
| is_deleted        | boolean         | Whether message was deleted   |
| created_at        | string          | ISO 8601 timestamp            |
| read_at           | string \| null  | ISO 8601 timestamp when read  |
| sender_username   | string          | Sender's username             |
| receiver_username | string          | Receiver's username           |

---

## Edit Room Message

Edit a previously sent room message. Only the sender can edit their own messages.

### Request

```
PUT /api/messages/{id}
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| id | integer | Message ID |

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |
| Content-Type | application/json |

**Body:**

```json
{
  "content": "Updated message content"
}
```

### Response

**Success (200 OK):**

```json
{
  "id": 1,
  "room_id": 1,
  "sender_id": 1,
  "content": "Updated message content",
  "message_type": "text",
  "is_edited": true,
  "edited_at": "2026-04-04T12:30:00Z",
  "is_deleted": false,
  "created_at": "2026-04-04T12:00:00Z",
  "updated_at": "2026-04-04T12:30:00Z"
}
```

**Errors:**
| Status | Response |
|--------|----------|
| 400 | `{"error": "Invalid message ID"}` |
| 400 | `{"error": "Message content is required"}` |
| 404 | `{"error": "Message not found or you don't have permission to edit it"}` |

### Example

```bash
curl -X PUT http://localhost:8080/api/messages/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "Updated message"}'
```

---

## Delete Room Message

Delete a previously sent room message (soft delete). Only the sender can delete their own messages.

### Request

```
DELETE /api/messages/{id}
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| id | integer | Message ID |

### Response

**Success (200 OK):**

```json
{
  "message": "Message deleted"
}
```

**Errors:**
| Status | Response |
|--------|----------|
| 404 | `{"error": "Message not found or you don't have permission to delete it"}` |

### Example

```bash
curl -X DELETE http://localhost:8080/api/messages/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Edit Direct Message

Edit a previously sent direct message. Only the sender can edit their own messages.

### Request

```
PUT /api/dm/messages/{id}
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| id | integer | Message ID |

**Body:**

```json
{
  "content": "Updated message content"
}
```

### Response

**Success (200 OK):**

```json
{
  "id": 1,
  "sender_id": 1,
  "receiver_id": 2,
  "content": "Updated message content",
  "message_type": "text",
  "is_read": true,
  "is_edited": true,
  "edited_at": "2026-04-04T12:30:00Z",
  "is_deleted": false,
  "created_at": "2026-04-04T12:00:00Z",
  "read_at": "2026-04-04T12:15:00Z"
}
```

### Example

```bash
curl -X PUT http://localhost:8080/api/dm/messages/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "Updated DM"}'
```

---

## Delete Direct Message

Delete a previously sent direct message (soft delete).

### Request

```
DELETE /api/dm/messages/{id}
```

### Response

**Success (200 OK):**

```json
{
  "message": "Message deleted"
}
```

### Example

```bash
curl -X DELETE http://localhost:8080/api/dm/messages/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Mark Room Messages as Read

Mark messages in a room as read up to a specific message.

### Request

```
POST /api/rooms/{id}/messages/read
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| id | integer | Room ID |

**Body (optional):**

```json
{
  "up_to_message_id": 100
}
```

If `up_to_message_id` is not provided, all messages will be marked as read.

### Response

**Success (200 OK):**

```json
{
  "message": "Messages marked as read"
}
```

### Example

```bash
curl -X POST http://localhost:8080/api/rooms/1/messages/read \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"up_to_message_id": 100}'
```

---

## Get Unread Room Messages Count

Get the count of unread messages in a specific room.

### Request

```
GET /api/rooms/{id}/messages/unread
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| id | integer | Room ID |

### Response

**Success (200 OK):**

```json
{
  "unread_count": 5
}
```

### Example

```bash
curl -X GET http://localhost:8080/api/rooms/1/messages/unread \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Get Unread Direct Messages Count

Get the total count of unread direct messages.

### Request

```
GET /api/dm/unread
```

### Response

**Success (200 OK):**

```json
{
  "unread_count": 3
}
```

### Example

```bash
curl -X GET http://localhost:8080/api/dm/unread \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```
