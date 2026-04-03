# Authentication API

Base URL: `http://localhost:8080/api/auth`

## Register User

Create a new user account.

### Request

```
POST /api/auth/register
```

**Headers:**
| Header | Value |
|--------|-------|
| Content-Type | application/json |

**Body:**

```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

| Field    | Type   | Required | Description                    |
| -------- | ------ | -------- | ------------------------------ |
| username | string | Yes      | Unique username (max 50 chars) |
| email    | string | Yes      | Valid email address            |
| password | string | Yes      | Min 6 characters               |

### Response

**Success (201 Created):**

```json
{
  "user": {
    "id": 1,
    "username": "john",
    "email": "john@example.com",
    "avatar_url": null,
    "status": "offline",
    "created_at": "2026-04-03T12:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Errors:**

| Status | Response                                                 |
| ------ | -------------------------------------------------------- |
| 400    | `{"error": "Username, email and password are required"}` |
| 400    | `{"error": "Password must be at least 6 characters"}`    |
| 409    | `{"error": "User already exists"}`                       |

### Example

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "email": "john@example.com",
    "password": "secret123"
  }'
```

---

## Login

Authenticate and receive a JWT token.

### Request

```
POST /api/auth/login
```

**Headers:**
| Header | Value |
|--------|-------|
| Content-Type | application/json |

**Body:**

```json
{
  "email": "string",
  "password": "string"
}
```

| Field    | Type   | Required | Description              |
| -------- | ------ | -------- | ------------------------ |
| email    | string | Yes      | Registered email address |
| password | string | Yes      | Account password         |

### Response

**Success (200 OK):**

```json
{
  "user": {
    "id": 1,
    "username": "john",
    "email": "john@example.com",
    "avatar_url": null,
    "status": "offline",
    "created_at": "2026-04-03T12:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Errors:**

| Status | Response                                       |
| ------ | ---------------------------------------------- |
| 400    | `{"error": "Email and password are required"}` |
| 401    | `{"error": "Invalid credentials"}`             |

### Example

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "secret123"
  }'
```

---

## JWT Token Usage

After login/register, use the token in the `Authorization` header for protected routes:

```
Authorization: Bearer <token>
```

**Token Payload:**

```json
{
  "user_id": 1,
  "username": "john",
  "email": "john@example.com",
  "exp": 1712246400,
  "iat": 1712160000
}
```

Token expires after 24 hours (configurable via `JWT_EXPIRY_HOURS` env variable).
