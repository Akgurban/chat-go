# Chat-Go - Real-time Chat Application Backend

A real-time chat application backend built with Go, PostgreSQL (plain SQL), and WebSockets.

## Features

- 🔐 **Authentication**: JWT-based user registration and login
- 💬 **Real-time Messaging**: WebSocket-based chat functionality
- 🏠 **Chat Rooms**: Create and join public/private rooms
- 📨 **Direct Messages**: One-on-one private messaging
- 🔒 **Plain SQL**: No ORM, using raw PostgreSQL queries

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
│   ├── database/
│   │   ├── db.go                # Database connection
│   │   └── migrations.go        # Migration runner
│   ├── handler/                 # HTTP handlers
│   ├── middleware/              # Auth, CORS, Logging
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

## Getting Started

### 1. Setup PostgreSQL Database

```bash
# Create database
psql -U postgres -c "CREATE DATABASE chat_go;"
```

### 2. Configure Environment

```bash
# Copy example env file
cp .env.example .env

# Edit .env with your database credentials
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run the Server

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
