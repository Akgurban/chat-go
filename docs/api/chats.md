# Chats API

The Chats API provides a combined view of all conversations (both direct messages and group chats) with unread counts and recent messages.

## Endpoints

### Get All Chats

Returns all chats (both DMs and rooms) sorted by last message time.

```
GET /api/chats
```

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| include_messages | boolean | false | Include recent messages for each chat |
| message_limit | integer | 10 | Number of recent messages to include (1-50) |

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |

**Response:**

```json
{
  "chats": [
    {
      "id": 2,
      "type": "direct",
      "name": "jane",
      "avatar": "https://example.com/avatar.jpg",
      "unread_count": 4,
      "last_message": {
        "id": 456,
        "content": "Hey, how are you?",
        "sender_id": 2,
        "sender_username": "jane",
        "message_type": "text",
        "is_read": false,
        "created_at": "2026-04-04T12:30:00Z"
      },
      "last_message_at": "2026-04-04T12:30:00Z",
      "is_online": true,
      "created_at": "2026-04-01T10:00:00Z"
    },
    {
      "id": 1,
      "type": "room",
      "name": "General",
      "description": "General chat room",
      "is_private": false,
      "unread_count": 2,
      "last_message": {
        "id": 123,
        "content": "Hello everyone!",
        "sender_id": 3,
        "sender_username": "bob",
        "message_type": "text",
        "created_at": "2026-04-04T12:00:00Z"
      },
      "last_message_at": "2026-04-04T12:00:00Z",
      "member_count": 5,
      "created_at": "2026-04-01T09:00:00Z"
    }
  ],
  "total_unread": 6
}
```

**With Messages:**

```
GET /api/chats?include_messages=true&message_limit=10
```

```json
{
  "chats": [
    {
      "id": 2,
      "type": "direct",
      "name": "jane",
      "unread_count": 4,
      "last_message": { ... },
      "last_message_at": "2026-04-04T12:30:00Z",
      "recent_messages": [
        {
          "id": 450,
          "sender_id": 1,
          "receiver_id": 2,
          "content": "Hi Jane!",
          "message_type": "text",
          "is_read": true,
          "created_at": "2026-04-04T12:00:00Z",
          "sender_username": "john",
          "receiver_username": "jane"
        },
        {
          "id": 456,
          "sender_id": 2,
          "receiver_id": 1,
          "content": "Hey, how are you?",
          "message_type": "text",
          "is_read": false,
          "created_at": "2026-04-04T12:30:00Z",
          "sender_username": "jane",
          "receiver_username": "john"
        }
      ],
      "is_online": true,
      "created_at": "2026-04-01T10:00:00Z"
    }
  ],
  "total_unread": 4
}
```

**Example:**

```bash
# Get all chats with unread counts
curl -X GET http://localhost:8080/api/chats \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Get all chats with last 10 messages each
curl -X GET "http://localhost:8080/api/chats?include_messages=true&message_limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

### Get Single Chat

Returns a single chat with its messages.

```
GET /api/chats/{type}/{id}
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| type | string | Chat type: `room` or `direct` |
| id | integer | Room ID (for rooms) or User ID (for DMs) |

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| message_limit | integer | 10 | Number of messages to include (1-100) |

**Response (Room):**

```json
{
  "id": 1,
  "type": "room",
  "name": "General",
  "description": "General chat room",
  "is_private": false,
  "unread_count": 2,
  "member_count": 5,
  "last_message": {
    "id": 123,
    "content": "Hello everyone!",
    "sender_id": 3,
    "sender_username": "bob",
    "message_type": "text",
    "created_at": "2026-04-04T12:00:00Z"
  },
  "last_message_at": "2026-04-04T12:00:00Z",
  "recent_messages": [
    {
      "id": 120,
      "room_id": 1,
      "sender_id": 1,
      "content": "Good morning!",
      "message_type": "text",
      "is_edited": false,
      "is_deleted": false,
      "created_at": "2026-04-04T11:00:00Z",
      "updated_at": "2026-04-04T11:00:00Z",
      "sender_username": "john",
      "sender_avatar": null
    },
    ...
  ],
  "created_at": "2026-04-01T09:00:00Z"
}
```

**Response (Direct Message):**

```json
{
  "id": 2,
  "type": "direct",
  "name": "jane",
  "avatar": "https://example.com/avatar.jpg",
  "unread_count": 4,
  "is_online": true,
  "last_message": {
    "id": 456,
    "content": "Hey, how are you?",
    "sender_id": 2,
    "sender_username": "jane",
    "message_type": "text",
    "is_read": false,
    "created_at": "2026-04-04T12:30:00Z"
  },
  "last_message_at": "2026-04-04T12:30:00Z",
  "recent_messages": [
    {
      "id": 450,
      "sender_id": 1,
      "receiver_id": 2,
      "content": "Hi Jane!",
      "message_type": "text",
      "is_read": true,
      "is_edited": false,
      "is_deleted": false,
      "created_at": "2026-04-04T12:00:00Z",
      "sender_username": "john",
      "receiver_username": "jane"
    },
    ...
  ],
  "created_at": "2026-04-01T10:00:00Z"
}
```

**Example:**

```bash
# Get room chat with 10 messages
curl -X GET "http://localhost:8080/api/chats/room/1?message_limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Get direct message chat with user ID 2
curl -X GET "http://localhost:8080/api/chats/direct/2?message_limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Chat Item Fields

### Common Fields

| Field           | Type    | Description                               |
| --------------- | ------- | ----------------------------------------- |
| id              | integer | Room ID or User ID                        |
| type            | string  | `room` or `direct`                        |
| name            | string  | Room name or username                     |
| unread_count    | integer | Number of unread messages                 |
| last_message    | object  | Last message preview                      |
| last_message_at | string  | ISO 8601 timestamp of last message        |
| recent_messages | array   | Array of recent messages (when requested) |
| created_at      | string  | ISO 8601 timestamp                        |

### Room-specific Fields

| Field        | Type    | Description               |
| ------------ | ------- | ------------------------- |
| description  | string  | Room description          |
| is_private   | boolean | Whether room is private   |
| member_count | integer | Number of members in room |

### Direct Message-specific Fields

| Field     | Type    | Description                      |
| --------- | ------- | -------------------------------- |
| avatar    | string  | User's avatar URL                |
| is_online | boolean | Whether user is currently online |

---

## Usage Tips

1. **Initial Load**: Use `GET /api/chats?include_messages=true` to load the chat list with recent messages for displaying previews.

2. **Polling**: Use `GET /api/chats` without messages to quickly check for unread counts.

3. **Opening a Chat**: Use `GET /api/chats/{type}/{id}?message_limit=50` to load more messages when user opens a chat.

4. **Real-time Updates**: Connect via WebSocket and use the `mark_read` message type to update read status.

5. **Sorting**: Chats are automatically sorted by `last_message_at` (most recent first).
