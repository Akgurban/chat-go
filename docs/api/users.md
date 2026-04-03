# Users API

Base URL: `http://localhost:8080/api`

> ⚠️ All endpoints require authentication. Include JWT token in Authorization header.

---

## Get Current User

Get the authenticated user's profile.

### Request

```
GET /api/me
```

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |

### Response

**Success (200 OK):**

```json
{
  "id": 1,
  "username": "john",
  "email": "john@example.com",
  "avatar_url": null,
  "status": "online",
  "created_at": "2026-04-03T12:00:00Z"
}
```

**Errors:**

| Status | Response                                     |
| ------ | -------------------------------------------- |
| 401    | `{"error": "Authorization header required"}` |
| 401    | `{"error": "Invalid or expired token"}`      |

### Example

```bash
curl -X GET http://localhost:8080/api/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Get All Users

Retrieve a list of all registered users.

### Request

```
GET /api/users
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
    "avatar_url": "https://example.com/avatar.jpg",
    "status": "offline",
    "created_at": "2026-04-03T13:00:00Z"
  }
]
```

### Example

```bash
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Get User by ID

Retrieve a specific user's profile.

### Request

```
GET /api/users/{id}
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| id | integer | User ID |

**Headers:**
| Header | Value |
|--------|-------|
| Authorization | Bearer \<token\> |

### Response

**Success (200 OK):**

```json
{
  "id": 2,
  "username": "jane",
  "email": "jane@example.com",
  "avatar_url": "https://example.com/avatar.jpg",
  "status": "offline",
  "created_at": "2026-04-03T13:00:00Z"
}
```

**Errors:**

| Status | Response                       |
| ------ | ------------------------------ |
| 400    | `{"error": "Invalid user ID"}` |
| 404    | `{"error": "User not found"}`  |

### Example

```bash
curl -X GET http://localhost:8080/api/users/2 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## User Object

| Field      | Type           | Description                              |
| ---------- | -------------- | ---------------------------------------- |
| id         | integer        | Unique user identifier                   |
| username   | string         | User's display name                      |
| email      | string         | User's email address                     |
| avatar_url | string \| null | URL to user's avatar image               |
| status     | string         | User status: `online`, `offline`, `away` |
| created_at | string         | ISO 8601 timestamp of account creation   |
