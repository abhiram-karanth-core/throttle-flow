# Throttle-Flow

Throttle-Flow is a standalone distributed rate-limiter service written in Go. It exposes a simple HTTP API that other services can call to determine whether a request should be allowed based on a rate-limit policy.

## Overview

- Redis-backed fixed window rate limiter
- Deployed as a separate infrastructure service
- Stateless at the HTTP layer
- Safe under concurrency
- Uses atomic Redis Lua scripts
- Clean abstractions via Go interfaces

This project is infrastructure code, not an application library.

## Architecture
```
Client Service
    |
    | POST /check
    v
Throttle-Flow (Go)
    |
    v
  Redis
```

Multiple Throttle-Flow instances can safely share the same Redis backend.

## Core Design

- The server depends on a `Limiter` interface
- Rate-limiting logic is isolated from HTTP
- Redis usage is fully encapsulated inside limiter implementations
- Limiter algorithms can be swapped without changing server code
- All Redis mutations required for rate-limiting are atomic

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
|-------|-------------|
| `key` | Identity to rate limit (user, API key, IP, etc.) |
| `limit` | Maximum allowed requests in the window |
| `window_ms` | Fixed window duration in milliseconds |

#### Response
```json
{
  "allowed": true,
  "remaining": 2
}
```

| Field | Description |
|-------|-------------|
| `allowed` | Whether the request is allowed |
| `remaining` | Remaining allowed requests in the current window |

## Rate-Limiting Algorithm

Throttle-Flow currently uses a Redis-backed fixed window counter implemented using a Lua script for atomicity.

### Why Lua?

The rate limiter performs:

1. `INCR` to increment the request counter
2. `PEXPIRE` to attach a TTL on the first request in the window

Both operations are executed inside a single atomic Lua script.

This guarantees:

- No race condition between increment and expiry
- Expiry is always attached when the window starts
- Safe behavior under concurrent requests
- Correct behavior across multiple service instances
- No risk of permanently stuck counters

### Window Lifecycle

**First request:**
- Counter becomes 1
- TTL is attached

**Subsequent requests:**
- Counter increments
- TTL remains unchanged

**When TTL expires:**
- Redis automatically deletes the key
- A new request starts a new window

Counters are not decremented; they are reset via key expiration.

### Failure Safety

Without atomic execution, a failure between `INCR` and `EXPIRE` could leave a key without TTL, permanently blocking future requests.

Throttle-Flow avoids this by executing both operations inside Redis as a single atomic script.

### Concurrency Model

- Go handles HTTP requests concurrently (goroutines per request)
- Redis processes commands sequentially
- Lua ensures atomic counter and expiry updates
- Safe for multi-instance deployments

## Current Limitations

This implementation uses a fixed window algorithm, which means:

- Burst traffic is possible at window boundaries
- Counters increment even for blocked requests

This is acceptable as a deterministic baseline implementation.

## Planned Improvements

- Token bucket limiter
- Sliding window limiter
- No-op limiter for testing
- Client-side middleware
- Metrics and observability
- Explicit failure semantics (fail-open / fail-closed)
- Configurable burst behavior

## Design Philosophy

Stable interfaces, replaceable implementations.

The HTTP layer depends only on the `Limiter` interface, not on a specific rate-limiting algorithm. Algorithms can evolve without breaking service contracts.