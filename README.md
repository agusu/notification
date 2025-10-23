# Notification API

Multi-channel notification system (Email, SMS, Push) with asynchronous processing using the Transactional Outbox pattern.

## Architecture

### Project Structure

```
notification/
├── cmd/api/              # Application entry point
│   ├── main.go          # Configuration and bootstrap
│   ├── route.go         # HTTP routes definition
│   └── middleware/      # Middlewares (authentication)
├── controllers/         # HTTP handlers
├── services/           # Business logic
│   ├── notifier/       # Notification service + worker
│   └── user/           # User service and authentication
├── models/             # Data models (GORM)
├── channels/           # Notification channel implementations
├── storage/            # Database configuration
└── docs/               # Swagger documentation
```

### Transactional Outbox Pattern

The system uses the **Transactional Outbox** pattern to guarantee eventual delivery:

1. Notification and outbox event are created in the same transaction
2. A periodic worker reads pending events from the outbox
3. Notifications are sent to the corresponding channel
4. Status is updated to `SENT` or a retry is scheduled

This approach ensures consistency between database and messaging, automatic retries on failure, and prevents message loss.

## Quick Start

### Prerequisites

- Go 1.24+
- MySQL 8.0+
- Docker (optional)

### Setup

```bash
# Clone and configure
git clone <repository-url>
cd notification
cp .env.example .env

# Edit .env with your configuration
# MYSQL_HOST, MYSQL_PORT, MYSQL_DB, MYSQL_USER, MYSQL_PASSWORD, JWT_SECRET

# Start MySQL (Docker Compose)
docker-compose up -d

# Or create database manually
mysql -u root -p
CREATE DATABASE notification;

# Install dependencies and run
go mod download
go run ./cmd/api
```

API available at `http://localhost:8080`  
Swagger UI at `http://localhost:8080/swagger/index.html`

## Notification Channels

### Email
Sends templated HTML emails.

**Required metadata:** `to` (email), `subject` (optional), `template` (optional: "titled" or "plain")

### SMS
Sends SMS messages (max 160 characters).

**Required metadata:** `phone` (E.164 format), `send_date` (YYYY-MM-DD)

### Push
Sends push notifications to mobile devices.

**Required metadata:** `token` (device token), `platform` ("android" or "ios"), `data` (optional)

## Authentication

The API uses JWT (JSON Web Tokens) for authentication.

### Register User

```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com","password":"pass123"}'
```

### Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"pass123"}'
# Returns: {"token": "eyJhbGc..."}
```

### Use Token

```bash
curl http://localhost:8080/notifications \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## API Endpoints

### Public (no authentication required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/signup` | Register user |
| POST | `/login` | Login and get token |
| GET | `/notifications/channels/schemas` | Get metadata schemas per channel |

### Protected (authentication required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/notifications` | Create notification |
| GET | `/notifications` | List notifications |
| GET | `/notifications/:id` | Get notification |
| PATCH | `/notifications/:id` | Update notification |
| DELETE | `/notifications/:id` | Delete notification |

## Usage Examples

### Create Email Notification

```bash
curl -X POST http://localhost:8080/notifications \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Welcome",
    "content": "Welcome to our platform!",
    "channel_name": "email",
    "meta": {"to": "user@example.com", "subject": "Welcome!"}
  }'
```

See `/docs/notification-examples.md` for more examples (SMS, Push).

## Development

### Regenerate Swagger
```bash
swag init -g cmd/api/main.go -d . -o ./docs --parseDependency --parseInternal
```

### Run Tests
```bash
go test ./...
```

### Build
```bash
go build -o bin/api ./cmd/api
```

## Design Decisions

**Transactional Outbox Pattern**: Guarantees eventual delivery, prevents message loss, enables automatic retries.

**Channel Separation**: Each channel implements the `Channel` interface. Easy to add new channels without modifying the core.

**Custom Errors**: Typed errors (`ErrInvalidCredentials`, `ErrNotificationNotFound`) with semantically correct HTTP codes.

**JWT Authentication**: 30-day token expiration, reusable middleware, clean token extraction with `strings.CutPrefix`.

**Metadata Validation**: Each channel validates its required fields with clear error messages.

**Async Worker**: Processes outbox every 30 seconds (configurable), handles retries and status updates.

**Idempotency**: SHA-256 hash of (user_id + channel + title + content + meta) prevents duplicates.



## Data Models

**User**: `id`, `name`, `email` (unique), `password` (bcrypt hashed), `created_at`

**Notification**: `id`, `user_id`, `title`, `content`, `channel_name`, `idempotency_key` (unique), `created_at`, `deleted_at` (soft delete)

**Outbox**: `id`, `notification_id`, `channel_name`, `payload_json`, `status` (PENDING/PROCESSING/SENT/FAILED), `attempts`, `max_attempts`, `last_error`, `next_attempt_at`

## Documentation

- Swagger UI: http://localhost:8080/swagger/index.html
- Detailed examples: `/docs/notification-examples.md`
- Channel schemas: `GET /notifications/channels/schemas`

