# Project Context - Chat-Go

## Overview

A real-time chat application backend built with Go, PostgreSQL, WebSockets, and Redis.

## Tech Stack

| Component | Technology              | Version |
| --------- | ----------------------- | ------- |
| Language  | Go                      | 1.21+   |
| Database  | PostgreSQL              | 13+     |
| Cache     | Redis                   | 6+      |
| WebSocket | gorilla/websocket       | 1.5.1   |
| Auth      | JWT (golang-jwt/jwt/v5) | 5.2.0   |

## Project Structure

```
chat-go/
├── cmd/server/main.go           # Entry point
├── config/config.go             # Environment configuration
├── internal/
│   ├── cache/                   # Redis caching layer
│   │   ├── cache.go             # Combined cache manager
│   │   ├── presence.go          # User online status (TTL: 60s)
│   │   ├── pubsub.go            # Real-time messaging across servers
│   │   ├── ratelimit.go         # API rate limiting
│   │   ├── redis.go             # Redis client wrapper
│   │   ├── session.go           # JWT session management
│   │   ├── typing.go            # Typing indicators (TTL: 3s)
│   │   └── unread.go            # Unread message counts
│   ├── database/                # PostgreSQL connection & migrations
│   ├── handler/                 # HTTP & WebSocket handlers
│   ├── middleware/              # Auth, CORS, Logging, RateLimit
│   ├── models/                  # Data structures
│   ├── repository/              # Database queries (plain SQL)
│   ├── service/                 # Business logic
│   └── websocket/               # Hub & Client management
├── migrations/                  # SQL migration files
├── static/                      # Frontend files
└── docs/api/                    # API documentation
```

## Key Files to Know

### Configuration

- **`.env`** - Environment variables (copy from `.env.example`)
- **`config/config.go`** - Loads DB, Redis, JWT, Server configs

### Database

- **`internal/database/db.go`** - PostgreSQL connection
- **`internal/repository/*.go`** - All SQL queries (no ORM)
- **`migrations/*.sql`** - Database schema

### Redis Cache

- **`internal/cache/cache.go`** - Main entry point for all cache operations
- Gracefully handles Redis unavailability (app works without Redis)

### WebSocket

- **`internal/websocket/hub.go`** - Manages all connected clients
- **`internal/websocket/client.go`** - Individual client handling
- **`internal/handler/websocket_handler.go`** - WebSocket upgrade & auth

## Environment Variables

```bash
# Required
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=chat_go
JWT_SECRET=your-secret-key

# Optional
SERVER_PORT=8080
JWT_EXPIRY_HOURS=24
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## API Endpoints Summary

| Method | Endpoint             | Description                      |
| ------ | -------------------- | -------------------------------- |
| POST   | `/api/auth/register` | User registration                |
| POST   | `/api/auth/login`    | User login                       |
| GET    | `/api/users`         | List users                       |
| GET    | `/api/rooms`         | List rooms                       |
| POST   | `/api/rooms`         | Create room                      |
| GET    | `/api/chats`         | Get all chats with unread counts |
| GET    | `/api/dm/{userID}`   | Get direct messages              |
| WS     | `/ws?token=JWT`      | WebSocket connection             |

## WebSocket Message Types

| Type             | Direction     | Description             |
| ---------------- | ------------- | ----------------------- |
| `join_room`      | Client→Server | Join a chat room        |
| `leave_room`     | Client→Server | Leave a chat room       |
| `chat_message`   | Client→Server | Send room message       |
| `direct_message` | Client→Server | Send DM                 |
| `typing`         | Client→Server | Typing indicator (room) |
| `typing_dm`      | Client→Server | Typing indicator (DM)   |
| `mark_read`      | Client→Server | Mark messages as read   |
| `new_message`    | Server→Client | New message received    |
| `user_typing`    | Server→Client | Someone is typing       |

## Redis Key Patterns

| Pattern                           | Purpose             | TTL        |
| --------------------------------- | ------------------- | ---------- |
| `user:online:{userID}`            | Presence status     | 60s        |
| `unread:{userID}:{type}:{chatID}` | Unread counts       | None       |
| `typing:{type}:{chatID}`          | Typing users hash   | 10s        |
| `ratelimit:{action}:{identifier}` | Rate limit counters | Varies     |
| `session:{userID}:{tokenID}`      | Active sessions     | JWT expiry |
| `blacklist:{tokenID}`             | Revoked tokens      | JWT expiry |

## Current State

- ✅ Authentication (JWT)
- ✅ User management
- ✅ Chat rooms (public/private)
- ✅ Direct messaging
- ✅ WebSocket real-time messaging
- ✅ Message read receipts
- ✅ Message edit/delete
- ✅ Redis caching layer (optional)
- ✅ Rate limiting middleware
- ✅ Notifications system

## Running the Project

```bash
# 1. Start PostgreSQL and Redis (see redis.md)
# 2. Copy and configure .env
cp .env.example .env

# 3. Run server
go run cmd/server/main.go
```

## Testing WebSocket

```javascript
// Browser console
const ws = new WebSocket("ws://localhost:8080/ws?token=YOUR_JWT_TOKEN");
ws.onmessage = (e) => console.log(JSON.parse(e.data));
ws.send(JSON.stringify({ type: "join_room", payload: { room_id: 1 } }));
```
