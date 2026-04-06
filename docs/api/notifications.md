# Notifications API

The Notifications API allows users to manage their notifications, push subscriptions, and notification preferences.

## Endpoints

### Get Notifications

```
GET /api/notifications
```

Retrieves the user's notifications.

**Query Parameters:**

- `limit` (optional): Number of notifications to return (default: 20, max: 100)
- `offset` (optional): Pagination offset (default: 0)
- `unread` (optional): If "true", returns only unread notifications

**Response:**

```json
[
  {
    "id": 1,
    "user_id": 123,
    "type": "direct_message",
    "title": "Message from John",
    "body": "Hey, how are you?",
    "data": "{\"sender_id\": 456, \"message_id\": 789}",
    "is_read": false,
    "is_pushed": true,
    "reference_id": 789,
    "created_at": "2026-04-04T10:30:00Z",
    "read_at": null
  }
]
```

### Get Unread Count

```
GET /api/notifications/count
```

Returns the count of unread notifications.

**Response:**

```json
{
  "count": 5
}
```

### Get All Unread Counts

```
GET /api/notifications/counts
```

Returns counts of all unread items (notifications, direct messages, room messages).

**Response:**

```json
{
  "total_unread": 15,
  "direct_message_unread": 3,
  "room_unread": {
    "1": 5,
    "2": 2
  },
  "notification_unread": 5
}
```

### Mark Notification as Read

```
POST /api/notifications/{id}/read
```

Marks a specific notification as read.

**Response:**

```json
{
  "message": "Notification marked as read"
}
```

### Mark All Notifications as Read

```
POST /api/notifications/read-all
```

Marks all notifications as read.

**Response:**

```json
{
  "message": "All notifications marked as read"
}
```

### Delete Notification

```
DELETE /api/notifications/{id}
```

Deletes a specific notification.

**Response:**

```json
{
  "message": "Notification deleted"
}
```

---

## Notification Preferences

### Get Preferences

```
GET /api/notifications/preferences
```

Returns the user's notification preferences.

**Response:**

```json
{
  "id": 1,
  "user_id": 123,
  "email_notifications": true,
  "push_notifications": true,
  "direct_message_notify": true,
  "mention_notify": true,
  "room_message_notify": true,
  "mute_all": false,
  "quiet_hours_enabled": false,
  "quiet_hours_start": null,
  "quiet_hours_end": null,
  "created_at": "2026-04-04T10:00:00Z",
  "updated_at": "2026-04-04T10:00:00Z"
}
```

### Update Preferences

```
PUT /api/notifications/preferences
```

Updates the user's notification preferences. All fields are optional.

**Request Body:**

```json
{
  "email_notifications": true,
  "push_notifications": true,
  "direct_message_notify": true,
  "mention_notify": true,
  "room_message_notify": false,
  "mute_all": false,
  "quiet_hours_enabled": true,
  "quiet_hours_start": "22:00",
  "quiet_hours_end": "07:00"
}
```

**Response:** Returns the updated preferences object.

---

## Push Notifications

### Register Push Subscription

```
POST /api/notifications/push/subscribe
```

Registers a Web Push subscription for the user.

**Request Body:**

```json
{
  "endpoint": "https://fcm.googleapis.com/fcm/send/...",
  "p256dh": "BNcRdreALRFX...",
  "auth": "tBHItJ...",
  "user_agent": "Mozilla/5.0..."
}
```

**Response:**

```json
{
  "id": 1,
  "user_id": 123,
  "endpoint": "https://fcm.googleapis.com/fcm/send/...",
  "p256dh": "BNcRdreALRFX...",
  "auth": "tBHItJ...",
  "user_agent": "Mozilla/5.0...",
  "created_at": "2026-04-04T10:00:00Z",
  "updated_at": "2026-04-04T10:00:00Z"
}
```

### Unregister Push Subscription

```
DELETE /api/notifications/push/unsubscribe?endpoint={endpoint}
```

Removes a push subscription.

**Query Parameters:**

- `endpoint` (required): The push subscription endpoint to remove

**Response:**

```json
{
  "message": "Push subscription removed"
}
```

---

## Notification Types

| Type             | Description                     |
| ---------------- | ------------------------------- |
| `message`        | New message in a room           |
| `direct_message` | New direct message              |
| `mention`        | User was mentioned in a message |
| `room_invite`    | Invitation to join a room       |
| `room_join`      | Someone joined a room           |
| `system`         | System notification             |

---

## WebSocket Notifications

When connected via WebSocket, notifications are pushed in real-time:

```json
{
  "type": "notification",
  "payload": {
    "id": 1,
    "user_id": 123,
    "type": "direct_message",
    "title": "Message from John",
    "body": "Hey, how are you?",
    "data": "{\"sender_id\": 456, \"message_id\": 789}",
    "is_read": false,
    "created_at": "2026-04-04T10:30:00Z"
  }
}
```
