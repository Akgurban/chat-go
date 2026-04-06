# Chat-Go - Real-time Chat Application Backend

A real-time chat application backend bui### 3. Configure Environment

```bash
# Copy example env file
cp .env.example .env

# Edit .env with your database and Redis credentials
```

### 4. Install Dependencies, PostgreSQL (plain SQL), WebSockets, and Redis.

## Features

- 🔐 **Authentication**: JWT-based user registration and login
- 💬 **Real-time Messaging**: WebSocket-based chat functionality
- 🏠 **Chat Rooms**: Create and join public/private rooms
- 📨 **Direct Messages**: One-on-one private messaging
- 🔒 **Plain SQL**: No ORM, using raw PostgreSQL queries
- 📦 **Redis Caching**: Optional Redis support for:
  - User presence/online status
  - Unread message counts
  - Rate limiting
  - Typing indicators
  - Pub/Sub for multi-server support
  - Session management

## Project Structure

```
chat-go/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   └── config.go                # Configuration management
├── docs/
│   └── api/
│       ├── README.md            # API overview
│       ├── authentication.md    # Auth endpoints docs
│       ├── users.md             # Users endpoints docs
│       ├── rooms.md             # Rooms endpoints docs
│       ├── messages.md          # Messages endpoints docs
│       └── websocket.md         # WebSocket docs
├── internal/
│   ├── cache/                   # Redis caching layer
│   │   ├── cache.go             # Combined cache manager
│   │   ├── presence.go          # User online status
│   │   ├── pubsub.go            # Pub/Sub messaging
│   │   ├── ratelimit.go         # Rate limiting
│   │   ├── redis.go             # Redis client
│   │   ├── session.go           # Session management
│   │   ├── typing.go            # Typing indicators
│   │   └── unread.go            # Unread counts
│   ├── database/
│   │   ├── db.go                # Database connection
│   │   └── migrations.go        # Migration runner
│   ├── handler/                 # HTTP handlers
│   ├── middleware/              # Auth, CORS, Logging, Rate Limiting
│   ├── models/                  # Data models
│   ├── repository/              # Database operations (plain SQL)
│   ├── service/                 # Business logic
│   └── websocket/               # WebSocket hub & client
├── migrations/
│   └── 001_init.sql             # Initial database schema
├── .env.example
├── go.mod
└── README.md
```

## Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Redis 6+ (optional, for caching features)

## Getting Started

### 1. Setup PostgreSQL Database

```bash
# Create database
psql -U postgres -c "CREATE DATABASE chat_go;"
```

### 2. Setup Redis (Optional)

```bash
# Using Docker
docker run -d --name redis -p 6379:6379 redis:alpine

# Or install locally
# macOS: brew install redis && brew services start redis
# Ubuntu: sudo apt install redis-server && sudo systemctl start redis
```

### 3. Configure Environment

```bash
# Copy example env file
cp .env.example .env

# Edit .env with your database credentials
```

### 4. Install Dependencies

```bash
go mod download
```

### 5. Run the Server

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## API Documentation

For detailed API documentation, see the [docs/api](./docs/api/) folder:

| Documentation                                  | Description            |
| ---------------------------------------------- | ---------------------- |
| [Authentication](./docs/api/authentication.md) | Register & Login       |
| [Users](./docs/api/users.md)                   | User profiles          |
| [Rooms](./docs/api/rooms.md)                   | Chat room management   |
| [Messages](./docs/api/messages.md)             | Room & direct messages |
| [WebSocket](./docs/api/websocket.md)           | Real-time messaging    |

## Quick Start

### Register a User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "email": "john@example.com", "password": "secret123"}'
```

### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "john@example.com", "password": "secret123"}'
```

### Create a Room

```bash
curl -X POST http://localhost:8080/api/rooms \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "General", "description": "General chat room"}'
```

### Connect via WebSocket

```javascript
const ws = new WebSocket("ws://localhost:8080/ws?token=YOUR_TOKEN");
ws.onopen = () => {
  ws.send(JSON.stringify({ type: "join_room", payload: { room_id: 1 } }));
};
```

## Migrations

Migrations are stored in the `migrations/` folder and run automatically on server start.

| File           | Description                                   |
| -------------- | --------------------------------------------- |
| `001_init.sql` | Initial schema (users, rooms, messages, etc.) |

To add a new migration, create a file like `002_add_feature.sql` in the migrations folder.

## Environment Variables

| Variable           | Default   | Description        |
| ------------------ | --------- | ------------------ |
| `DB_HOST`          | localhost | PostgreSQL host    |
| `DB_PORT`          | 5432      | PostgreSQL port    |
| `DB_USER`          | postgres  | Database user      |
| `DB_PASSWORD`      | -         | Database password  |
| `DB_NAME`          | chat_go   | Database name      |
| `SERVER_PORT`      | 8080      | HTTP server port   |
| `JWT_SECRET`       | -         | JWT signing secret |
| `JWT_EXPIRY_HOURS` | 24        | Token expiry time  |

## License

MIT
