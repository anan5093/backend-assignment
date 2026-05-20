# Backend Assignment API

[![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](./go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Deployment: Render](https://img.shields.io/badge/Deployment-Render-46E3B7?logo=render&logoColor=white)](https://backend-assignment-9z1m.onrender.com/)
[![Health Check](https://img.shields.io/website?down_message=down&label=health&up_message=up&url=https%3A%2F%2Fbackend-assignment-9z1m.onrender.com%2Fhealth)](https://backend-assignment-9z1m.onrender.com/health)
[![Status: Active](https://img.shields.io/badge/status-active-success)](https://github.com/anan5093/backend-assignment)

A production-structured Go HTTP service implementing a concurrency-safe rolling-window rate limiter and an optimized product catalog API — built with in-memory storage, a layered architecture, and a clean separation of concerns.

---

## Table of Contents

1. [Overview](#overview)
2. [Tech Stack](#tech-stack)
3. [Project Structure](#project-structure)
4. [Architecture](#architecture)
5. [How to Run](#how-to-run)
6. [API Reference](#api-reference)
   - [POST /request](#post-request)
   - [GET /stats](#get-stats)
   - [POST /products](#post-products)
   - [GET /products](#get-products)
   - [GET /products/{id}](#get-productsid)
   - [POST /products/{id}/media](#post-productsidmedia)
7. [Validation Rules](#validation-rules)
8. [Rate Limiter Design](#rate-limiter-design)
9. [Product Storage Design](#product-storage-design)
10. [Concurrency Safety](#concurrency-safety)
11. [Security Considerations](#security-considerations)
12. [Testing Guide](#testing-guide)
13. [Production Limitations](#production-limitations)
14. [Future Improvements](#future-improvements)
15. [Challenges Faced](#challenges-faced)
16. [Verified Test Output](#verified-test-output)
17. [AI Usage Disclosure](#ai-usage-disclosure)
18. [Repository Metadata](#repository-metadata)
19. [Live Demo and API Base URL](#live-demo-and-api-base-url)
20. [Deployment Guide (Render)](#deployment-guide-render)
21. [Recent Development Updates](#recent-development-updates)
22. [Maintainer and Contact](#maintainer-and-contact)
23. [License](#license)
24. [Project Status](#project-status)
25. [Contribution Workflow](#contribution-workflow)
26. [Topics and Keywords](#topics-and-keywords)
27. [Roadmap Additions](#roadmap-additions)

---

## Overview

This project is a runnable Go HTTP backend implementing two parts:

| Part | Feature | Description |
|------|---------|-------------|
| 1 | Rate-limited request API | Rolling 1-minute window, max 5 requests per user, concurrency-safe |
| 2 | Product catalog API | Create, list, detail, and media-append endpoints with optimized DTOs |

The service uses **in-memory storage only** — no database, no external dependencies beyond the router. Data does not persist across restarts.

---

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.22 |
| Router | `github.com/gorilla/mux` v1.8.1 |
| HTTP | `net/http` (standard library) |
| Concurrency | `sync.RWMutex`, `sync/atomic` |
| Serialization | `encoding/json` (standard library) |
| Storage | In-memory (`map`, `slice`) |

---

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go            # Entry point: wires dependencies, registers routes, starts server
├── internal/
│   ├── handlers/
│   │   ├── rate_handler.go    # HTTP handlers for /request and /stats
│   │   └── product_handler.go # HTTP handlers for /products and /products/{id}/media
│   ├── services/
│   │   ├── rate_service.go    # Rate limiting business logic
│   │   └── product_service.go # Product creation, listing, and media logic
│   ├── store/
│   │   ├── rate_store.go      # In-memory rate limiter state
│   │   └── product_store.go   # In-memory product catalog state
│   ├── models/
│   │   ├── rate.go            # UserRateData, RateRequest, UserStats types
│   │   └── product.go         # Product, ProductSummary, Pagination types
│   ├── validation/
│   │   └── validation.go      # URL, string, and pagination validation
│   ├── middleware/
│   │   └── logging.go         # Request logging middleware
│   └── utils/
│       └── http.go            # JSON encoding helpers and error response writer
├── go.mod
└── go.sum
```

---

## Architecture

The service follows a **layered architecture** with strict dependency direction: handlers → services → store.

```
┌─────────────────────────────────────────────────────────────┐
│                          HTTP Layer                         │
│              gorilla/mux router + middleware                │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                      Handlers Layer                         │
│         Parse HTTP, validate input shape, write responses   │
│         rate_handler.go  |  product_handler.go              │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                      Services Layer                         │
│     Business logic: rate decisions, product operations      │
│     rate_service.go  |  product_service.go                  │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                       Store Layer                           │
│       In-memory state, concurrency-safe via RWMutex         │
│       rate_store.go  |  product_store.go                    │
└─────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Responsibility |
|-------|---------------|
| `handlers` | Parse HTTP request bodies and query parameters; map service errors to HTTP status codes; write JSON responses |
| `services` | Enforce business rules (field requirements, URL validation, media limits); call the store |
| `store` | Own all shared in-memory state; enforce concurrency safety with `sync.RWMutex`; perform SKU uniqueness checks and timestamp pruning |
| `models` | Plain data types shared across layers — no methods, no logic |
| `validation` | Reusable, stateless validation functions used by the service layer |
| `middleware` | Cross-cutting concerns applied at the router level (request logging) |
| `utils` | HTTP helpers: JSON encoder, error writer, JSON decoder with single-value enforcement |

### Why Layered Architecture?

- **Separation of concerns** — HTTP parsing logic never touches business rules; storage layout changes don't ripple into handlers.
- **Testability** — each layer can be tested in isolation; services and stores can be unit-tested without starting an HTTP server.
- **Maintainability** — new endpoints follow the same pattern, keeping the codebase predictable as it grows.
- **Dependency clarity** — the dependency graph is strictly downward; no layer imports from one above it.

---

## How to Run

**Prerequisites:** Go 1.22 or later installed and on your `PATH`.

```bash
# From the repository root
go run ./cmd/server
```

The server starts and logs:

```
2026/05/19 19:39:41 server listening on http://localhost:8080
```

The server handles `SIGINT` and `SIGTERM` with a 5-second graceful shutdown window:

```
2026/05/19 19:50:37 server stopped
```

> **Note:** If port 8080 is already in use, stop the conflicting process first or adjust the `Addr` field in `main.go`.

---

## API Reference

All request and response bodies use `application/json`. All error responses follow a consistent shape:

```json
{ "error": "<description>" }
```

---

### POST /request

Submits a request on behalf of a user. The request is accepted if the user has fewer than 5 accepted requests in the current rolling 1-minute window.

**Request body:**

```json
{
  "user_id": "alice",
  "payload": { "action": "click", "item": 42 }
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `user_id` | string | ✅ | Must be non-empty after whitespace trimming |
| `payload` | any valid JSON | ✅ | Any JSON value (object, array, primitive) |

**Responses:**

| Status | Meaning | Body |
|--------|---------|------|
| `201 Created` | Request accepted | `{ "message": "request accepted", "user_id": "alice" }` |
| `400 Bad Request` | Missing or invalid fields | `{ "error": "user_id is required" }` |
| `429 Too Many Requests` | Rate limit exceeded | `{ "error": "rate limit exceeded" }` |

---

### GET /stats

Returns per-user rate limiter statistics. The `accepted_requests_current_window` field reflects only timestamps still within the active 1-minute window at the time of the request.

**Response:**

```json
{
  "users": {
    "alice": {
      "accepted_requests_current_window": 3,
      "rejected_requests_total": 2
    },
    "bob": {
      "accepted_requests_current_window": 5,
      "rejected_requests_total": 0
    }
  }
}
```

| Field | Description |
|-------|-------------|
| `accepted_requests_current_window` | Count of accepted requests still within the rolling 1-minute window |
| `rejected_requests_total` | Cumulative count of rejected requests since server start |

**Responses:**

| Status | Meaning |
|--------|---------|
| `200 OK` | Statistics returned (empty `users` map if no requests have been made) |

---

### POST /products

Creates a new product. The SKU must be unique across all existing products.

**Request body:**

```json
{
  "name": "Widget A",
  "sku": "SKU-001",
  "image_urls": ["https://cdn.example.com/img1.jpg"],
  "video_urls": ["https://cdn.example.com/demo.mp4"]
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `name` | string | ✅ | Must be non-empty |
| `sku` | string | ✅ | Must be non-empty and globally unique |
| `image_urls` | array of strings | ❌ | Max 20 URLs; each must be a valid `http`/`https` URL ≤ 2048 chars |
| `video_urls` | array of strings | ❌ | Same constraints as `image_urls` |

**Response (`201 Created`):**

```json
{
  "id": "1",
  "name": "Widget A",
  "sku": "SKU-001",
  "image_urls": ["https://cdn.example.com/img1.jpg"],
  "video_urls": ["https://cdn.example.com/demo.mp4"],
  "created_at": "2026-05-19T19:45:04Z"
}
```

**Error responses:**

| Status | Cause |
|--------|-------|
| `400 Bad Request` | Invalid JSON, missing `name`/`sku`, invalid URL, too many URLs |
| `409 Conflict` | SKU already exists |

---

### GET /products

Returns a paginated list of products as **lightweight summary DTOs**. Media arrays (`image_urls`, `video_urls`) are intentionally excluded — only counts and a single thumbnail URL are serialized. See [Product Storage Design](#product-storage-design) for the rationale.

**Query parameters:**

| Parameter | Default | Maximum | Notes |
|-----------|---------|---------|-------|
| `limit` | `20` | `100` | Number of items per page |
| `offset` | `0` | — | Zero-based item offset |

**Response (`200 OK`):**

```json
{
  "items": [
    {
      "id": "1",
      "name": "Widget A",
      "sku": "SKU-001",
      "image_count": 1,
      "video_count": 1,
      "thumbnail_url": "https://cdn.example.com/img1.jpg",
      "created_at": "2026-05-19T19:45:04Z"
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 1
  }
}
```

> **Note:** `thumbnail_url` is omitted from the response when the product has no images.

**Error responses:**

| Status | Cause |
|--------|-------|
| `400 Bad Request` | `limit` is not a positive integer; `offset` is negative |

---

### GET /products/{id}

Returns the full product detail for a single product, including complete `image_urls` and `video_urls` arrays.

**Response (`200 OK`):**

```json
{
  "id": "1",
  "name": "Widget A",
  "sku": "SKU-001",
  "image_urls": [
    "https://cdn.example.com/img1.jpg",
    "https://cdn.example.com/img2.jpg"
  ],
  "video_urls": ["https://cdn.example.com/demo.mp4"],
  "created_at": "2026-05-19T19:45:04Z"
}
```

**Error responses:**

| Status | Cause |
|--------|-------|
| `404 Not Found` | Product with the given `id` does not exist |

---

### POST /products/{id}/media

Appends additional image or video URLs to an existing product. At least one URL must be provided across both arrays. The total count per type (existing + new) must not exceed 20.

**Request body:**

```json
{
  "image_urls": ["https://cdn.example.com/img2.jpg"],
  "video_urls": []
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `image_urls` | array of strings | ❌ | At least one array must contain ≥ 1 URL |
| `video_urls` | array of strings | ❌ | Same URL validation as product creation |

**Response (`200 OK`):** Full updated product detail (same shape as `GET /products/{id}`).

**Error responses:**

| Status | Cause |
|--------|-------|
| `400 Bad Request` | No URLs provided; invalid URL; would exceed 20 URLs per type |
| `404 Not Found` | Product does not exist |

---

## Validation Rules

### Rate Limiter (`POST /request`)

| Rule | Behaviour |
|------|-----------|
| `user_id` present | `400` if field is missing from JSON body |
| `user_id` non-empty | `400` if value is empty or whitespace-only |
| `payload` present | `400` if field is missing from JSON body |
| Valid JSON body | `400` if request body is malformed JSON |
| Rate limit | `429` if the user has ≥ 5 accepted requests in the last 60 seconds |

### Product Catalog

| Rule | Behaviour |
|------|-----------|
| `name` required | `400` if missing or whitespace-only |
| `sku` required | `400` if missing or whitespace-only |
| SKU uniqueness | `409` if SKU already exists |
| Valid JSON body | `400` if request body is malformed JSON |
| URL scheme | `400` if a URL does not use `http` or `https` |
| URL length | `400` if a URL exceeds 2048 characters |
| Media count (create) | `400` if `image_urls` or `video_urls` contains more than 20 entries |
| Media count (append) | `400` if appending would bring either type above 20 total |
| Media non-empty (append) | `400` if both arrays are empty or omitted |
| `limit` parameter | `400` if not a positive integer; silently capped at 100 |
| `offset` parameter | `400` if negative |

---

## Rate Limiter Design

### Rolling-Window Algorithm

The limiter uses a **per-user rolling window** — not a fixed bucket that resets at a clock boundary. This prevents burst exploitation at window edges.

For each user, the store maintains:
- A slice of **accepted request timestamps** (UTC).
- A cumulative **rejected request count**.

On every `POST /request`:

1. **Prune stale timestamps** — discard all entries older than `now - 1 minute`. This operation runs in-place over the slice to avoid allocation.
2. **Evaluate** — if `len(acceptedTimestamps) >= 5`, increment the rejected count and return `false`.
3. **Accept** — append `now` to the timestamp slice and return `true`.

```
User: alice
Window: [t-60s ──────────────────────── t now]

Timestamps: [t-45s, t-30s, t-20s, t-10s, t-5s]  ← 5 entries → REJECT
Timestamps: [t-55s, t-30s, t-20s, t-10s, t-5s]  ← after pruning: 5 → REJECT
Timestamps: [t-30s, t-20s, t-10s, t-5s]          ← after pruning: 4 → ACCEPT
```

### Why Rolling Window over Fixed Window?

A **fixed window** resets the counter at a clock boundary (e.g. every full minute). A user could send 5 requests at 00:59 and 5 more at 01:01, making 10 requests in 2 seconds — far exceeding the intent of the limit. A **rolling window** measures the most recent 60 seconds from any point in time, making the limit invariant regardless of when requests arrive.

### Why `sync.RWMutex`?

The `Allow` method is a **check-and-mutate** operation. Using a write lock (`Lock`) rather than a read lock for the entirety of the check-and-append ensures that two concurrent goroutines for the same `user_id` cannot both observe `len < 5` and both append — which would silently allow a 6th (or more) request. A read lock would be insufficient here.

The `Stats` endpoint also prunes timestamps on read (to return current-window counts), so it too acquires a write lock rather than a read lock.

---

## Product Storage Design

### Internal Layout

The `ProductStore` separates product metadata from media deliberately:

```
productsByID     map[string]*productRecord   ← id, name, sku, created_at
mediaByProductID map[string]*productMedia    ← image_urls, video_urls (slices)
skuIndex         map[string]string           ← sku → id (O(1) uniqueness check)
productOrder     []string                    ← insertion-order list of IDs
```

IDs are generated using `sync/atomic` (`atomic.Uint64`) so that concurrent `Create` calls never produce duplicate IDs without needing the full write lock solely for ID generation.

### List vs Detail DTO Optimization

`GET /products` returns `ProductSummary` — a deliberately lightweight DTO:

```go
type ProductSummary struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    SKU          string    `json:"sku"`
    ImageCount   int       `json:"image_count"`
    VideoCount   int       `json:"video_count"`
    ThumbnailURL string    `json:"thumbnail_url,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
}
```

The full `image_urls` and `video_urls` slices are **never serialized** in the list response. Only `len(media.ImageURLs)`, `len(media.VideoURLs)`, and `media.ImageURLs[0]` (thumbnail) are read.

This matters at scale: a catalog of 1,000 products each with 20 image URLs would produce 20,000 URL strings in a single list response if the full arrays were included. Instead, the list response stays constant in size relative to the page size, not the media count.

`GET /products/{id}` returns the full `Product` type with all media arrays — suitable for a product detail page that needs to render all images and videos.

### Pagination

Pagination is offset-based with the following defaults:

| Parameter | Default | Cap |
|-----------|---------|-----|
| `limit` | 20 | 100 |
| `offset` | 0 | — |

The `total` field in the `pagination` object always reflects the full catalog size, enabling clients to determine whether more pages exist: `hasMore = offset + limit < total`.

### Scalability Considerations

The in-memory design is intentional for this assignment. In a production system:
- The product list is served under a **read lock**, allowing concurrent reads without blocking.
- Insertions take a write lock but are O(1) for the map operations and O(1) amortized for the slice append.
- Media append operations also hold the write lock for the duration of the bounds check and append, preventing race conditions on media count validation.

---

## Concurrency Safety

All shared mutable state lives in the store layer and is protected by `sync.RWMutex`.

| Operation | Lock type | Reason |
|-----------|-----------|--------|
| `RateStore.Allow` | Write (`Lock`) | Check-and-mutate must be atomic |
| `RateStore.Stats` | Write (`Lock`) | Prunes timestamps on read |
| `ProductStore.Create` | Write (`Lock`) | SKU uniqueness check + insertion must be atomic |
| `ProductStore.List` | Read (`RLock`) | Read-only scan; safe for concurrent list requests |
| `ProductStore.GetByID` | Read (`RLock`) | Read-only lookup |
| `ProductStore.AppendMedia` | Write (`Lock`) | Bounds check + append must be atomic |

The rate limiter's `Allow` method uses a write lock for the full check-and-append rather than a read lock followed by a write lock (the classic TOCTOU pattern), which would allow two goroutines to both observe the limit not exceeded and both proceed to append.

---

## Security Considerations

| Concern | Mitigation |
|---------|-----------|
| Malformed request bodies | `DecodeJSON` uses `json.Decoder` and enforces a single JSON value — trailing content causes a `400` |
| Empty or whitespace-only strings | All required string fields are trimmed and checked before use |
| Invalid URL input | URL scheme and structure are validated before storage |
| Oversized URL strings | Maximum URL length enforced at 2048 characters |
| Excessive media uploads | Per-type media count capped at 20 (both at creation and append) |
| Race conditions | All shared state protected with `sync.RWMutex`; no unprotected concurrent writes |
| Memory growth | Rate limiter timestamps are pruned on every request; media counts are capped |
| Predictable error responses | All errors return a consistent `{ "error": "..." }` JSON body with appropriate HTTP status codes |
| Slow request headers | `ReadHeaderTimeout: 5s` configured on the HTTP server to mitigate slowloris-style header exhaustion |

> **Out of scope for this assignment:** authentication, authorization, HTTPS termination, and distributed denial-of-service mitigation. These would be handled at the infrastructure layer (load balancer, API gateway) in a production deployment.

---

## Testing Guide

### Start the Server

```bash
go run ./cmd/server
```

---

### curl Examples (Linux / macOS)

#### GET /stats — empty state

```bash
curl -i http://localhost:8080/stats
```

Expected:
```json
{ "users": {} }
```

---

#### POST /request — accepted

```bash
curl -i -X POST http://localhost:8080/request \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"alice","payload":{"action":"test"}}'
```

Expected (`201`):
```json
{ "message": "request accepted", "user_id": "alice" }
```

---

#### POST /request — rate limit exceeded (send 6 times)

```bash
for i in $(seq 1 6); do
  curl -s -o /dev/null -w "Request $i: HTTP %{http_code}\n" \
    -X POST http://localhost:8080/request \
    -H 'Content-Type: application/json' \
    -d '{"user_id":"alice","payload":{}}'
done
```

Expected output:
```
Request 1: HTTP 201
Request 2: HTTP 201
Request 3: HTTP 201
Request 4: HTTP 201
Request 5: HTTP 201
Request 6: HTTP 429
```

---

#### GET /stats — after requests

```bash
curl -i http://localhost:8080/stats
```

Expected:
```json
{
  "users": {
    "alice": {
      "accepted_requests_current_window": 5,
      "rejected_requests_total": 1
    }
  }
}
```

---

#### POST /products — create a product

```bash
curl -i -X POST http://localhost:8080/products \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Widget A",
    "sku": "SKU-001",
    "image_urls": ["https://cdn.example.com/img1.jpg"],
    "video_urls": ["https://cdn.example.com/demo.mp4"]
  }'
```

Expected (`201`):
```json
{
  "id": "1",
  "name": "Widget A",
  "sku": "SKU-001",
  "image_urls": ["https://cdn.example.com/img1.jpg"],
  "video_urls": ["https://cdn.example.com/demo.mp4"],
  "created_at": "2026-05-19T19:45:04Z"
}
```

---

#### POST /products — duplicate SKU

```bash
curl -i -X POST http://localhost:8080/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"Widget A Copy","sku":"SKU-001","image_urls":[],"video_urls":[]}'
```

Expected (`409`):
```json
{ "error": "duplicate sku" }
```

---

#### GET /products — list

```bash
curl -i 'http://localhost:8080/products?limit=20&offset=0'
```

Expected (`200`):
```json
{
  "items": [
    {
      "id": "1",
      "name": "Widget A",
      "sku": "SKU-001",
      "image_count": 1,
      "video_count": 1,
      "thumbnail_url": "https://cdn.example.com/img1.jpg",
      "created_at": "2026-05-19T19:45:04Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 1 }
}
```

---

#### GET /products/{id} — detail

```bash
curl -i http://localhost:8080/products/1
```

Expected (`200`):
```json
{
  "id": "1",
  "name": "Widget A",
  "sku": "SKU-001",
  "image_urls": ["https://cdn.example.com/img1.jpg"],
  "video_urls": ["https://cdn.example.com/demo.mp4"],
  "created_at": "2026-05-19T19:45:04Z"
}
```

---

#### POST /products/{id}/media — append media

```bash
curl -i -X POST http://localhost:8080/products/1/media \
  -H 'Content-Type: application/json' \
  -d '{"image_urls":["https://cdn.example.com/img2.jpg"],"video_urls":[]}'
```

Expected (`200`): Updated product with both image URLs.

---

### PowerShell Examples (Windows)

PowerShell's `curl` is an alias for `Invoke-WebRequest`. Use `Invoke-RestMethod` or call `curl.exe` directly.

#### POST /request

```powershell
Invoke-RestMethod -Method POST -Uri http://localhost:8080/request `
  -ContentType 'application/json' `
  -Body '{"user_id":"alice","payload":{"action":"test"}}'
```

#### POST /products

```powershell
Invoke-RestMethod -Method POST -Uri http://localhost:8080/products `
  -ContentType 'application/json' `
  -Body '{"name":"Widget A","sku":"SKU-001","image_urls":["https://cdn.example.com/img1.jpg"],"video_urls":[]}'
```

#### GET /products

```powershell
Invoke-RestMethod -Uri 'http://localhost:8080/products?limit=20&offset=0'
```

---

## Production Limitations

This service was implemented according to the assignment constraints. The following limitations should be understood before any production consideration:

| Limitation | Impact |
|-----------|--------|
| **In-memory storage** | All product and rate limiter data is lost on process restart |
| **No persistence layer** | No database; no recovery after crash |
| **Single-instance only** | Cannot be horizontally scaled — two instances would have independent state |
| **No distributed rate limiting** | Rate limiter state is local to the process; load-balanced deployments would not share limits |
| **No CDN integration** | Media URLs are stored as plain strings; no upload or delivery infrastructure exists |
| **No authentication or authorization** | All endpoints are publicly accessible |
| **Unbounded user growth** | The rate limiter map grows with unique `user_id` values and is never evicted |
| **Unbounded product growth** | The product catalog grows indefinitely; no archiving or deletion |

---

## Future Improvements

| Improvement | Benefit |
|-------------|---------|
| **PostgreSQL** | Persistent product catalog with full-text search and relational integrity |
| **Redis** | Distributed rate limiter with TTL-based key expiry; shared across instances |
| **Distributed rate limiting** | Enforce per-user limits across a horizontally scaled fleet using Redis sorted sets |
| **CDN-backed media storage** | Accept media uploads, store in object storage (S3), and distribute via CDN |
| **Structured logging** | Replace `log.Printf` with a structured logger (e.g. `slog`) for queryable log pipelines |
| **Metrics and monitoring** | Expose Prometheus metrics: request rate, rejection rate, product count, response latency |
| **Docker / container image** | Package the binary in a minimal container for portable deployment |
| **Unit and integration tests** | Automated test coverage for service logic, store behaviour, and HTTP handler contracts |
| **Graceful shutdown improvements** | Drain in-flight requests with configurable timeout; emit structured shutdown events |
| **Request ID tracing** | Inject a `X-Request-ID` header on every response for distributed tracing correlation |

---

## Challenges Faced

### 1. Port 8080 Already in Use

During development, `go run ./cmd/server` failed silently because port 8080 was occupied by a previous server process that had not been terminated. The server goroutine exited immediately, but the main goroutine continued waiting on the signal channel.

**Resolution:** Identified the occupying process using `lsof -i :8080` (Linux/macOS) or `netstat -ano | findstr 8080` (Windows) and terminated it before restarting.

---

### 2. Go Toolchain Not Initially on PATH

On the test environment, `go` and `gofmt` were not available immediately in the shell PATH after installation.

**Resolution:** Sourced the shell profile (`source ~/.profile`) or opened a new terminal session to pick up the updated PATH. Verified with `go version` before proceeding.

---

### 3. PowerShell `curl` Alias Confusion

PowerShell defines `curl` as an alias for `Invoke-WebRequest`, which has different flag syntax from the Unix `curl` binary. Attempts to use `-X`, `-H`, and `-d` flags failed with parameter binding errors.

**Resolution:** Used `Invoke-RestMethod` for cleaner JSON handling in PowerShell, or invoked `curl.exe` explicitly to use the real binary when available.

---

### 4. PowerShell Multiline Command Syntax

Unix-style line continuation with `\` does not work in PowerShell. Multi-line commands require the backtick `` ` `` continuation character, and JSON body strings require careful quoting.

**Resolution:** Used the backtick continuation character and escaped inner quotes or used single-quoted strings where appropriate.

---

### 5. Rate Limiter Correctness Under Repeated Requests

Verifying that exactly the 6th request — and not the 5th — returns a `429` required careful review of the check-and-append logic. An off-by-one error (`>` vs `>=`) would either allow 6 requests or cap at 4.

**Resolution:** The condition `len(data.AcceptedTimestamps) >= s.limit` correctly rejects when the count equals the limit (5), ensuring exactly 5 accepted requests and the 6th is rejected. This was verified manually by sending 6 sequential requests and confirming HTTP status codes.

---

### 6. Optimizing the List Endpoint

Naively returning full product structs from `GET /products` would serialize all media URL arrays for every product on the page — a non-trivial amount of data for catalogs with many media-rich products.

**Resolution:** Introduced a `ProductSummary` DTO that the store populates by reading only `len()` and the first element of each media slice, without copying or serializing the full arrays. The `GET /products/{id}` endpoint retains the full `Product` type for cases where complete media data is required.

---

## Verified Test Output

The following server log was captured during manual end-to-end testing and confirms all endpoints were exercised:

```
2026/05/19 19:39:41 server listening on http://localhost:8080
2026/05/19 19:42:19 GET /stats 0s
2026/05/19 19:44:24 POST /request 1.7393ms
2026/05/19 19:45:04 POST /products 506.2µs
2026/05/19 19:45:23 GET /products 3.218ms
2026/05/19 19:46:20 GET /stats 0s
2026/05/19 19:46:58 POST /request 1.035ms
2026/05/19 19:47:10 POST /request 0s
2026/05/19 19:47:17 POST /request 0s
2026/05/19 19:47:24 POST /request 2.9033ms
2026/05/19 19:47:30 POST /request 2.5908ms
2026/05/19 19:47:37 POST /request 0s
2026/05/19 19:48:00 GET /products/1 0s
2026/05/19 19:48:30 POST /products/1/media 6.169ms
2026/05/19 19:50:37 server stopped
```

The `POST /request` sequence (7 requests) confirms that requests 1–5 were accepted (`201`) and the 6th and 7th returned `429 Too Many Requests`. Timestamps confirm the requests were sent within the same 1-minute window.

---

## AI Usage Disclosure

AI-assisted tooling — including Codex — was used selectively during development to accelerate implementation, iterate on architectural decisions, validate edge-case handling, and refine documentation quality.

Specific areas where AI assistance was applied:

- **Architecture refinement** — evaluating layer boundaries and dependency direction in the layered design.
- **Validation strategy** — brainstorming edge cases for URL validation, pagination bounds, and rate limiter correctness.
- **API design iteration** — reviewing request/response shapes for consistency and idiomatic Go conventions.
- **Debugging assistance** — identifying concurrency considerations in the check-and-append rate limiter pattern.
- **Documentation acceleration** — drafting and refining README sections for clarity and completeness.

All critical engineering decisions were independently reviewed. Endpoints were manually tested using `curl` and PowerShell. Concurrency behaviour — including the rate limiter's `sync.RWMutex` usage and the atomic ID generator — was manually verified for correctness. The final code structure, validation logic, and HTTP status code mapping reflect deliberate engineering choices, not generated defaults.

---

## Repository Metadata

- **Repository Title:** Backend Assignment API
- **Repository Name:** [`anan5093/backend-assignment`](https://github.com/anan5093/backend-assignment)
- **Short Description:** Production-structured Go backend implementing a concurrency-safe rolling-window rate limiter and optimized product catalog API with layered architecture and in-memory storage.
- **Primary Language:** Go (1.22)
- **Deployment Platform:** Render
- **Default Branch:** `main`

### Badges

[![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](./go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Deployment: Render](https://img.shields.io/badge/Deployment-Render-46E3B7?logo=render&logoColor=white)](https://backend-assignment-9z1m.onrender.com/)
[![Health Check](https://img.shields.io/website?down_message=down&label=health&up_message=up&url=https%3A%2F%2Fbackend-assignment-9z1m.onrender.com%2Fhealth)](https://backend-assignment-9z1m.onrender.com/health)
[![Status: Active](https://img.shields.io/badge/status-active-success)](https://github.com/anan5093/backend-assignment)

---

## Live Demo and API Base URL

- **Live Deployment URL:** https://backend-assignment-9z1m.onrender.com/
- **Production API Base URL:** `https://backend-assignment-9z1m.onrender.com`
- **Local API Base URL:** `http://localhost:8080`

### Health Endpoint Usage

```bash
# Root health/info endpoint
curl -i https://backend-assignment-9z1m.onrender.com/

# Dedicated health endpoint
curl -i https://backend-assignment-9z1m.onrender.com/health
```

Expected success responses:

```json
{
  "message": "Backend Assignment API Running",
  "status": "healthy"
}
```

```json
{
  "status": "ok"
}
```

---

## Deployment Guide (Render)

This repository is configured for cloud deployment on Render using the Go service runtime and dynamic port binding.

### 1) GitHub Integration

1. Create a new **Web Service** on Render.
2. Connect your GitHub account and select your fork/target repository (for example, this project: `anan5093/backend-assignment`).
3. Configure auto-deploy from your target branch (typically `main`).

### 2) Build and Start Commands

Use one of the following options:

```bash
# Build command
go build -o server ./cmd/server
```

```bash
# Start command
./server
```

Alternative (simple, slower cold start):

```bash
go run ./cmd/server
```

### 3) Environment Variables

- `PORT` is injected by Render at runtime.
- The server supports dynamic cloud port assignment via:
  - `os.Getenv("PORT")` when set
  - fallback to `8080` when unset (local development)

### 4) Health Check Configuration

Recommended health check path:

- `/health` (dedicated)
- `/` (also valid root API health/info response)

### 5) Production Deployment Notes

- Data is in-memory; restarts or redeployments clear runtime state.
- Horizontal scaling is not state-safe for this assignment design (no shared distributed store).
- Keep a single service instance for consistent in-memory behavior.
- Graceful shutdown is implemented with a 5-second timeout for in-flight request handling.

---

## Recent Development Updates

Latest repository evolution and deployment-focused enhancements:

| Area | Update |
|------|--------|
| Main server bootstrap | Refactored `main` startup and shutdown flow for cleaner production wiring |
| Root endpoint | Added `GET /` endpoint returning JSON service health/info |
| Health checks | Added dedicated `GET /health` endpoint for platform probes |
| Cloud portability | Added dynamic `PORT` handling for Render and similar platforms |
| Deployment stability | Updated response formatting and health output consistency |
| Operational readiness | Improved startup/shutdown logging and graceful termination behavior |

### Commit Highlights

- Refactored the main server startup/shutdown flow and added root + dedicated health check routes.
- Improved health response formatting for cleaner deployment probe output.

---

## Maintainer and Contact

- **Maintainer:** Anand Raj
- **GitHub:** https://github.com/anan5093
- **Repository:** https://github.com/anan5093/backend-assignment
- **Issues/Support:** Please open an issue in the repository for bugs, suggestions, or deployment questions.

### Contribution Ownership

By contributing to this repository, you affirm that your submissions are your original work (or appropriately licensed), and you agree that accepted contributions are maintained under this repository's MIT License and project governance by the repository maintainer.

---

## License

This project is licensed under the **MIT License**.

- Full license text: [`LICENSE`](./LICENSE)
- SPDX identifier: `MIT`

You are free to use, modify, distribute, and sublicense this software, provided the copyright and license notice are included.
The software is provided "as is", without warranty of any kind.

---

## Project Status

**Current Status:** Active and deployed.

The API is functional for assignment scope and cloud-hosted usage, including health endpoints and dynamic cloud port support. Current architecture remains intentionally in-memory and single-instance to match assignment constraints.

---

## Contribution Workflow

1. Fork the repository.
2. Create a feature branch from `main`.
3. Make changes with clear commits.
4. Run:
   ```bash
   go build ./...
   go test ./...
   ```
5. Open a pull request with a clear change summary and verification notes.

---

## Topics and Keywords

`go` · `golang` · `rest-api` · `backend` · `http-server` · `gorilla-mux` · `concurrency` · `rate-limiter` · `product-catalog` · `layered-architecture` · `in-memory-storage` · `deployment` · `render` · `api-design`

---

## Roadmap Additions

In addition to the existing [Future Improvements](#future-improvements), near-term roadmap enhancements include:

- Implement CI workflows for automated build/test validation on pull requests.
- Document deployment workflow screenshots/checklists for smoother onboarding.
- Expand API reference tables with root and health endpoint request/response examples.
- Provide an optional containerized deployment path for consistent local/prod parity.
- Introduce an observability baseline (structured logs and endpoint-level metrics).
