# Rooms API

Base URL: `http://localhost:8080/api`

> ⚠️ All endpoints require authentication. Include JWT token in Authorization header.

---

## Get All Public Rooms

Retrieve a list of all public chat rooms.

### Request

```
GET /api/rooms
```

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
    "name": "General",
    "description": "General chat room",
    "is_private": false,
    "created_by": 1,
    "created_at": "2026-04-03T12:00:00Z",
    "updated_at": "2026-04-03T12:00:00Z"
  },
  {
    "id": 2,
    "name": "Tech Talk",
    "description": "Discuss technology",
    "is_private": false,
    "created_by": 2,
    "created_at": "2026-04-03T13:00:00Z",
    "updated_at": "2026-04-03T13:00:00Z"
  }
]
```

### Example

```bash
curl -X GET http://localhost:8080/api/rooms \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Get My Rooms

Retrieve rooms the current user has joined.

### Request

```
GET /api/rooms/my
```

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
    "name": "General",
    "description": "General chat room",
    "is_private": false,
    "created_by": 1,
    "created_at": "2026-04-03T12:00:00Z",
    "updated_at": "2026-04-03T12:00:00Z"
  }
]
```

### Example

```bash
curl -X GET http://localhost:8080/api/rooms/my \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Get Room Details

Retrieve details of a specific room including members.

### Request

```
GET /api/rooms/{id}
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| id | integer | Room ID |

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |

### Response

**Success (200 OK):**

```json
{
  "id": 1,
  "name": "General",
  "description": "General chat room",
  "is_private": false,
  "created_by": 1,
  "created_at": "2026-04-03T12:00:00Z",
  "updated_at": "2026-04-03T12:00:00Z",
  "members": [
    {
      "id": 1,
      "username": "john",
      "email": "john@example.com",
      "avatar_url": null,
      "status": "online",
      "created_at": "2026-04-03T12:00:00Z"
    },
    {
      "id": 2,
      "username": "jane",
      "email": "jane@example.com",
      "avatar_url": null,
      "status": "offline",
      "created_at": "2026-04-03T13:00:00Z"
    }
  ]
}
```

**Errors:**

| Status | Response                       |
| ------ | ------------------------------ |
| 400    | `{"error": "Invalid room ID"}` |
| 404    | `{"error": "Room not found"}`  |

### Example

```bash
curl -X GET http://localhost:8080/api/rooms/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Create Room

Create a new chat room.

### Request

```
POST /api/rooms
```

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |
| Content-Type | application/json |

**Body:**

```json
{
  "name": "string",
  "description": "string",
  "is_private": boolean
}
```

| Field       | Type    | Required | Description               |
| ----------- | ------- | -------- | ------------------------- |
| name        | string  | Yes      | Room name (max 100 chars) |
| description | string  | No       | Room description          |
| is_private  | boolean | No       | Default: false            |

### Response

**Success (201 Created):**

```json
{
  "id": 3,
  "name": "My Room",
  "description": "A private room",
  "is_private": true,
  "created_by": 1,
  "created_at": "2026-04-03T14:00:00Z",
  "updated_at": "2026-04-03T14:00:00Z"
}
```

**Errors:**

| Status | Response                             |
| ------ | ------------------------------------ |
| 400    | `{"error": "Room name is required"}` |

### Example

```bash
curl -X POST http://localhost:8080/api/rooms \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "General",
    "description": "General chat room",
    "is_private": false
  }'
```

---

## Join Room

Join a public room.

### Request

```
POST /api/rooms/{id}/join
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| id | integer | Room ID |

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |

### Response

**Success (200 OK):**

```json
{
  "message": "Joined room successfully"
}
```

**Errors:**

| Status | Response                                |
| ------ | --------------------------------------- |
| 400    | `{"error": "Invalid room ID"}`          |
| 403    | `{"error": "Cannot join private room"}` |
| 404    | `{"error": "Room not found"}`           |

### Example

```bash
curl -X POST http://localhost:8080/api/rooms/1/join \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Leave Room

Leave a room.

### Request

```
POST /api/rooms/{id}/leave
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| id | integer | Room ID |

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |

### Response

**Success (200 OK):**

```json
{
  "message": "Left room successfully"
}
```

**Errors:**

| Status | Response                       |
| ------ | ------------------------------ |
| 400    | `{"error": "Invalid room ID"}` |

### Example

```bash
curl -X POST http://localhost:8080/api/rooms/1/leave \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Room Object

| Field       | Type            | Description             |
| ----------- | --------------- | ----------------------- |
| id          | integer         | Unique room identifier  |
| name        | string          | Room name               |
| description | string \| null  | Room description        |
| is_private  | boolean         | Whether room is private |
| created_by  | integer \| null | User ID of room creator |
| created_at  | string          | ISO 8601 timestamp      |
| updated_at  | string          | ISO 8601 timestamp      |
