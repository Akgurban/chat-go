# Chat-Go API Documentation

Welcome to the Chat-Go API documentation. This real-time chat application provides REST APIs and WebSocket connections for building chat applications.

## Base URL

```
http://localhost:8080
```

## API Reference

| Documentation                         | Description                 |
| ------------------------------------- | --------------------------- |
| [Authentication](./authentication.md) | User registration and login |
| [Users](./users.md)                   | User profiles and listing   |
| [Rooms](./rooms.md)                   | Chat room management        |
| [Messages](./messages.md)             | Room and direct messages    |
| [WebSocket](./websocket.md)           | Real-time messaging         |

## Quick Start

### 1. Register a User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "email": "john@example.com", "password": "secret123"}'
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "john@example.com", "password": "secret123"}'
```

### 3. Create a Room

```bash
curl -X POST http://localhost:8080/api/rooms \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "General", "description": "General chat"}'
```

### 4. Connect via WebSocket

```javascript
const ws = new WebSocket("ws://localhost:8080/ws?token=YOUR_TOKEN");
ws.onopen = () =>
  ws.send(
    JSON.stringify({
      type: "join_room",
      payload: { room_id: 1 },
    }),
  );
```

## Authentication

All endpoints except `/api/auth/*` require a JWT token in the Authorization header:

```
Authorization: Bearer <token>
```

## Response Format

### Success Response

```json
{
  "id": 1,
  "name": "Example",
  ...
}
```

### Error Response

```json
{
  "error": "Error message description"
}
```

## HTTP Status Codes

| Code | Description           |
| ---- | --------------------- |
| 200  | Success               |
| 201  | Created               |
| 400  | Bad Request           |
| 401  | Unauthorized          |
| 403  | Forbidden             |
| 404  | Not Found             |
| 409  | Conflict              |
| 500  | Internal Server Error |

## Health Check

```
GET /health
```

Response:

```json
{ "status": "ok" }
```
