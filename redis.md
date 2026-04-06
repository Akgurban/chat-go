# Redis Setup Guide

## Overview

Redis is used in this project for:

- **User Presence**: Track online/offline status
- **Unread Counts**: Fast message count caching
- **Typing Indicators**: Real-time "user is typing..."
- **Rate Limiting**: Prevent API abuse
- **Session Management**: Token validation & revocation
- **Pub/Sub**: Multi-server real-time messaging

> **Note**: Redis is optional. The app will work without it but with reduced functionality.

---

## Option 1: Use Existing Docker Redis

### Check if Redis Container Exists

```bash
# List all containers (including stopped)
docker ps -a | grep redis

# List running containers
docker ps | grep redis
```

### If Redis Container Exists but Stopped

```bash
# Start existing container
docker start redis

# Or start by container ID
docker start <container_id>
```

### Connect to Existing Redis

```bash
# Check which port Redis is mapped to
docker port redis

# Test connection
docker exec -it redis redis-cli ping
# Should return: PONG
```

### Find Redis Port

```bash
# See full container details
docker inspect redis | grep -A 10 "Ports"

# Or simpler
docker port redis 6379
```

---

## Option 2: Create New Redis Container

### Simple Redis (No Password)

```bash
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:alpine
```

### Redis with Password

```bash
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:alpine \
  redis-server --requirepass your_password_here
```

### Redis with Persistent Data

```bash
docker run -d \
  --name redis \
  -p 6379:6379 \
  -v redis_data:/data \
  redis:alpine \
  redis-server --appendonly yes
```

### Redis with Password + Persistence

```bash
docker run -d \
  --name redis \
  -p 6379:6379 \
  -v redis_data:/data \
  redis:alpine \
  redis-server --requirepass your_password_here --appendonly yes
```

---

## Option 3: Docker Compose (Recommended)

Create `docker-compose.yml` in project root:

```yaml
version: "3.8"

services:
  postgres:
    image: postgres:15-alpine
    container_name: chat-go-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: chat_go
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:alpine
    container_name: chat-go-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

volumes:
  postgres_data:
  redis_data:
```

### Run with Docker Compose

```bash
# Start services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs redis

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

---

## Configure .env File

After Redis is running, update your `.env`:

```bash
# Default (no password)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# With password
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_password_here
REDIS_DB=0

# Custom port
REDIS_HOST=localhost
REDIS_PORT=6380
REDIS_PASSWORD=
REDIS_DB=0
```

---

## Verify Connection

### From Terminal

```bash
# Connect to Redis CLI
docker exec -it redis redis-cli

# If using password
docker exec -it redis redis-cli -a your_password_here

# Test commands
127.0.0.1:6379> PING
PONG
127.0.0.1:6379> SET test "hello"
OK
127.0.0.1:6379> GET test
"hello"
127.0.0.1:6379> DEL test
(integer) 1
127.0.0.1:6379> exit
```

### From Go Application

Run the server and check logs:

```bash
go run cmd/server/main.go

# Success output:
# Connected to Redis successfully
# 📦 Redis caching enabled

# Failure output:
# Warning: Failed to connect to Redis: ...
# ⚠️  Redis caching disabled (connection failed)
```

---

## Common Issues

### Issue: Connection Refused

```
failed to connect to Redis: dial tcp 127.0.0.1:6379: connect: connection refused
```

**Solutions:**

```bash
# Check if container is running
docker ps | grep redis

# Start if stopped
docker start redis

# Check port mapping
docker port redis
```

### Issue: Wrong Port

```bash
# Find actual port
docker inspect redis --format='{{range $p, $conf := .NetworkSettings.Ports}}{{$p}} -> {{(index $conf 0).HostPort}}{{end}}'
```

### Issue: Authentication Failed

```
NOAUTH Authentication required
```

**Solution:** Set `REDIS_PASSWORD` in `.env`

### Issue: Container Name Already Exists

```bash
# Remove old container
docker rm redis

# Or use different name
docker run -d --name chat-redis -p 6379:6379 redis:alpine
```

---

## Redis Commands Cheat Sheet

```bash
# Connect
docker exec -it redis redis-cli

# Key operations
KEYS *                    # List all keys (use carefully in production)
KEYS user:online:*        # List keys by pattern
GET key                   # Get value
SET key value             # Set value
DEL key                   # Delete key
EXISTS key                # Check if exists
TTL key                   # Time to live
EXPIRE key seconds        # Set expiration

# Hash operations (used for typing indicators)
HSET key field value      # Set hash field
HGET key field            # Get hash field
HGETALL key               # Get all fields
HDEL key field            # Delete field

# Pub/Sub
SUBSCRIBE channel         # Subscribe to channel
PUBLISH channel message   # Publish message

# Server
INFO                      # Server info
DBSIZE                    # Number of keys
FLUSHDB                   # Clear current DB (careful!)
```

---

## Monitoring Redis

### Basic Stats

```bash
docker exec -it redis redis-cli INFO stats
```

### Memory Usage

```bash
docker exec -it redis redis-cli INFO memory
```

### Connected Clients

```bash
docker exec -it redis redis-cli CLIENT LIST
```

### Real-time Commands (Monitor)

```bash
docker exec -it redis redis-cli MONITOR
# Shows all commands in real-time (useful for debugging)
# Press Ctrl+C to exit
```

---

## Production Recommendations

1. **Always use a password** in production
2. **Enable persistence** (`--appendonly yes`)
3. **Set memory limits**: `--maxmemory 256mb --maxmemory-policy allkeys-lru`
4. **Use Redis Cluster** for high availability
5. **Regular backups** of RDB/AOF files
6. **Monitor memory usage** to prevent OOM

---

## Quick Start Summary

```bash
# 1. Start Redis
docker run -d --name redis -p 6379:6379 redis:alpine

# 2. Verify it's running
docker exec -it redis redis-cli ping

# 3. Update .env
echo "REDIS_HOST=localhost" >> .env
echo "REDIS_PORT=6379" >> .env
echo "REDIS_PASSWORD=" >> .env
echo "REDIS_DB=0" >> .env

# 4. Run the app
go run cmd/server/main.go

# Look for: "📦 Redis caching enabled"
```
