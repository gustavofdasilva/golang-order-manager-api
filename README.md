# Golang Order Manager API

A production-inspired REST API built with Go, designed around a complete order management workflow with authentication, session management, inventory control, and scalable layered architecture.

## Features

- JWT authentication with refresh token rotation
- Multi-device session management and revocation
- Order lifecycle management with enforced status transitions
- Automatic inventory control (stock reservation and restoration)
- Soft delete for users, products, and orders
- Pagination and filtering on list endpoints
- Per-IP rate limiting
- Structured JSON logging with request tracing
- Swagger documentation
- PostgreSQL with connection pooling and transactional guarantees

## Tech Stack

- **Language:** Go 1.24
- **Web Framework:** Echo v4
- **Database:** PostgreSQL
- **Auth:** JWT (`golang-jwt/jwt/v5`) + bcrypt
- **Migrations:** golang-migrate
- **Docs:** Swagger (swaggo)
- **Config:** godotenv

## Project Structure

```
.
├── cmd/api/main.go               # Entry point
├── internal/
│   ├── config/config.go          # Environment variable loading
│   ├── dto/                      # Request/response DTOs
│   ├── errors/errors.go          # Centralized error definitions
│   ├── handlers/                 # HTTP handlers (auth, user, product, order, health)
│   ├── middleware/               # Auth, logger, rate limiter
│   ├── models/                   # Domain models and business logic (e.g. status transitions)
│   ├── repositories/             # Data access layer (supports *sql.DB and *sql.Tx)
│   ├── responses/response.go     # Standard response envelope structs
│   ├── router/                   # Route definitions
│   ├── security/                 # Password hashing (bcrypt) and JWT generation
│   └── services/                 # Business logic layer
├── pkg/
│   ├── database/database.go      # PostgreSQL connection and pool setup
│   └── logger/logger.go          # Structured slog setup
├── migrations/                   # SQL migration files (golang-migrate)
├── docs/                         # Auto-generated Swagger docs
├── Makefile
├── .env.example
└── go.mod
```

## Environment Variables

Copy `.env.example` to `.env` and fill in the values.

### API

| Variable | Description | Default |
|---|---|---|
| `API_PORT` | HTTP server port | — |
| `API_VERSION` | API version prefix (e.g. `1` → `/api/v1/`) | — |

### Database

| Variable | Description | Default |
|---|---|---|
| `DB_HOST` | PostgreSQL host | — |
| `DB_PORT` | PostgreSQL port | — |
| `DB_USER` | Database user | — |
| `DB_PASS` | Database password | — |
| `DB_NAME` | Database name | — |
| `DB_SSLMODE` | SSL mode (`disable`, `require`, etc.) | — |

### Connection Pool

| Variable | Description | Default |
|---|---|---|
| `MAX_CONNECTIONS` | Max open DB connections | `10` |
| `MAX_IDLE_CONNECTIONS` | Max idle DB connections | `5` |
| `CONNECTION_MAX_LIFETIME_MINUTES` | Max connection lifetime (minutes) | `5` |

### Authentication

| Variable | Description | Default |
|---|---|---|
| `SECRET_KEY` | JWT signing secret — **change in production** | — |
| `TOKEN_EXPIRATION_MINUTES` | Access token lifetime (minutes) | `60` |
| `REFRESH_TOKEN_EXPIRATION_MINUTES` | Refresh token lifetime (minutes) | `10080` (7 days) |

### Rate Limiting

| Variable | Description | Default |
|---|---|---|
| `RATE_LIMIT_RPS` | Requests per second per IP (float) | `10` |
| `RATE_LIMIT_BURST` | Burst capacity per IP | `30` |
| `RATE_LIMIT_EXPIRES_MINUTES` | Inactivity window before limiter state is cleared (minutes) | `3` |

## Running Locally

**Prerequisites:** Go 1.24+, PostgreSQL, [golang-migrate CLI](https://github.com/golang-migrate/migrate)

```bash
# 1. Clone and install dependencies
git clone <repo-url>
cd golang-task-manager-api
go mod download

# 2. Configure environment
cp .env.example .env
# Edit .env with your database credentials and secrets

# 3. Run database migrations
make migrate-up

# 4. Start the server
make run
```

The server starts at `http://localhost:<API_PORT>`.

### Makefile Commands

| Command | Description |
|---|---|
| `make run` | Start the API server |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Rollback the last migration |

## API Endpoints

Base path: `/api/v{API_VERSION}`

### Auth (public)

| Method | Path | Description |
|---|---|---|
| `POST` | `/auth/register` | Create a new account |
| `POST` | `/auth/login` | Authenticate and receive access + refresh tokens |
| `POST` | `/auth/refresh` | Rotate tokens (old refresh token is revoked) |
| `POST` | `/auth/logout` | Revoke current session's refresh token |
| `POST` | `/auth/logout-all` | Revoke all active sessions for the user |

### Users (protected)

| Method | Path | Description |
|---|---|---|
| `GET` | `/users/me` | Get current user profile |
| `PATCH` | `/users/me` | Update username, email, or password |
| `DELETE` | `/users/me` | Soft-delete account |

### Products (GET is public, mutations are protected)

| Method | Path | Description |
|---|---|---|
| `GET` | `/products` | List products (pagination + filtering) |
| `GET` | `/products/{id}` | Get a single product |
| `POST` | `/products` | Create a product |
| `PATCH` | `/products/{id}` | Update a product |
| `DELETE` | `/products/{id}` | Soft-delete a product |

**Query parameters for `GET /products`:**

| Param | Type | Description |
|---|---|---|
| `page` | int | Page number (default: 1) |
| `limit` | int | Items per page (default: 10, max: 100) |
| `name` | string | Partial, case-insensitive name filter |
| `min_price` | float | Minimum price filter |
| `max_price` | float | Maximum price filter |
| `in_stock` | bool | If `true`, returns only products with stock > 0 |

### Orders (all protected)

| Method | Path | Description |
|---|---|---|
| `GET` | `/orders` | List authenticated user's orders (pagination + status filter) |
| `GET` | `/orders/{id}` | Get a specific order (scoped to the authenticated user) |
| `POST` | `/orders` | Create a new order |
| `PATCH` | `/orders/{id}/status` | Transition an order's status |

**Query parameters for `GET /orders`:**

| Param | Type | Description |
|---|---|---|
| `page` | int | Page number (default: 1) |
| `limit` | int | Items per page (default: 10, max: 100) |
| `status` | string | Filter by status (`pending`, `paid`, `completed`, `cancelled`) |

### Health

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Returns API and database connectivity status |

### Swagger

| Method | Path | Description |
|---|---|---|
| `GET` | `/swagger/*` | Interactive API documentation |

## Business Rules

### Authentication

- Passwords are hashed with **bcrypt at cost 14**.
- Refresh tokens are stored as **SHA-256 hashes** — the raw token is only returned once at login/refresh.
- On token refresh, the old refresh token is **immediately revoked** and a new pair is issued (rotation).
- **Logout-all** revokes every active refresh token for the user, terminating all device sessions.
- Protected endpoints require a valid `Authorization: Bearer <token>` header.

### Users

- Email and username must be **unique**.
- Deleting an account is a **soft delete** — the record is retained with a `deleted_at` timestamp and excluded from all queries.
- Password and username are **never returned** in API responses.

### Products

- `price` must be `>= 0`.
- `stock` must be `>= 0`.
- Deleted products are **soft-deleted** and excluded from all responses.
- Results are ordered by `created_at DESC`.

### Orders

**Status transitions** — only the following are allowed:

```
pending → paid
pending → cancelled
paid    → completed
paid    → cancelled
```

Any other transition is rejected with a `400` error.

**Order creation:**

- At least one item with `quantity > 0` is required.
- `total_amount` is calculated server-side from current product prices (not trusted from client).
- Each `OrderItem` stores the **unit price and subtotal at the time of purchase** (price snapshot).

**Stock management:**

- When an order transitions to **`paid`**: stock is decremented atomically inside a database transaction. If any product has insufficient stock, the entire operation is rolled back.
- When a **`paid`** order is **cancelled**: stock is automatically **restored** for all items.

**Isolation:**

- Users can only view and manage **their own orders** — cross-user access returns `404`.

## Response Format

### Success

```json
{
  "message": "success message",
  "data": { }
}
```

### Paginated list

```json
{
  "message": "success message",
  "data": [],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 42,
    "total_pages": 5
  }
}
```

### Error

```json
{
  "error": "error description"
}
```

## Database Schema

```
users
  id UUID PK, username TEXT, email TEXT UNIQUE,
  password TEXT (bcrypt), created_at, updated_at, deleted_at

items  (products)
  id UUID PK, name TEXT, description TEXT,
  price DECIMAL, stock INT,
  created_at, updated_at, deleted_at

orders
  id UUID PK, user_id UUID FK(users),
  status TEXT (pending|paid|completed|cancelled),
  total_amount DECIMAL,
  created_at, updated_at, deleted_at

order_items
  id UUID PK, order_id UUID FK(orders) ON DELETE CASCADE,
  item_id UUID FK(items), quantity INT,
  price_at_purchase DECIMAL, subtotal DECIMAL,
  created_at, updated_at

refresh_tokens
  user_id UUID FK(users), token TEXT (SHA-256 hash),
  expires_at TIMESTAMP, revoked_at TIMESTAMP
```
