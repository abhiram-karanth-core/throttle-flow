# Throttle-Flow

Throttle-Flow is a **standalone rate-limiter service** written in Go.  
It exposes a simple HTTP API that other services can call to determine whether a request should be allowed based on a rate-limit policy.

---

## What this is

- Redis-backed **fixed window** rate limiter
- Deployed as a **separate service**
- Stateless at the HTTP layer
- Uses clean abstractions via Go interfaces

This project is infrastructure code, not an application library.

---

## High-level architecture

```
Client Service
      |
      | POST /check
      v
Throttle-Flow
      |
      v
    Redis
```

---

## Core ideas

- The server depends on a `Limiter` interface
- Rate-limiting logic is isolated from HTTP
- Redis usage is fully encapsulated inside limiter implementations
- Limiter algorithms can be swapped without changing server code

---

## API

### `POST /check`

Checks whether a request should be allowed under a given rate-limit policy.

#### Request

```json
{
  "key": "user:123",
  "limit": 5,
  "window_ms": 60000            
}
```

| Field | Description |
|---|---|
| `key` | Identity to rate limit (user, API key, IP, etc.) |
| `limit` | Maximum allowed requests |
| `window_ms` | Fixed window duration in milliseconds |

#### Response

```json
{
  "allowed": true,
  "remaining": 2
}
```

| Field | Description |
|---|---|
| `allowed` | Whether the request is allowed |
| `remaining` | Remaining requests in the current window |

---

---

## Rate-limiting algorithm

Throttle-Flow currently uses a fixed window counter implemented with Redis:

- `INCR` for atomic counting
- `EXPIRE` to enforce window boundaries
- Window key based on truncated timestamps

This approach is simple, deterministic, and suitable as a baseline distributed limiter.

---

## Planned improvements

- Token bucket limiter
- Sliding window limiter
- No-op limiter for testing
- Client-side middleware
- Metrics and observability
- Explicit failure semantics (fail-open / fail-closed)

---

## Design philosophy

> Stable interfaces, replaceable implementations.

The server never depends on a specific rate-limiting algorithm — only on the behavior it provides.