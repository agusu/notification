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

- Docker & Docker Compose
- Go 1.24+ (for local development)
- MySQL 8.0+ (only if not using Docker)

### Setup with Docker (Recommended)

```bash
# Clone and configure
git clone <repository-url>
cd notification

# Create environment file (optional, uses defaults)
cp env.example .env
# Edit .env if needed

# Build and start all services
docker-compose up --build

# Or run in background
docker-compose up -d --build
```

To view logs: `docker-compose logs -f api`

To stop the services:
```bash
docker-compose down

# Stop and remove volumes (clears database)
docker-compose down -v
```

API available at `http://localhost:8080`  
Swagger UI at `http://localhost:8080/swagger/index.html`

### Setup without Docker (Local Development)

```bash
# Clone and configure
git clone <repository-url>
cd notification
cp env.example .env

# Edit .env with your configuration
# MYSQL_HOST=localhost, MYSQL_PORT=3306, etc.

# Start MySQL (Docker Compose or locally)
docker-compose up -d mysql

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

**Required metadata:** `phone` (E.164 format), `carrier` (e.g., "verizon", "att")

### Push
Sends push notifications to mobile devices.

**Required metadata:** `token` (device token), `platform` ("android" or "ios"), `data` (optional)

## Scheduled Notifications

Notifications can be scheduled for future delivery using the `scheduled_at` field (RFC3339 format).

### Send Immediately (default)
```bash
curl -X POST http://localhost:8080/notifications \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Welcome",
    "content": "Welcome to our platform!",
    "channel_name": "email",
    "meta": {"to": "user@example.com"}
  }'
```

### Schedule for Later
```bash
curl -X POST http://localhost:8080/notifications \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Reminder",
    "content": "Don't forget your meeting",
    "channel_name": "push",
    "meta": {"token": "device_token_xyz"},
    "scheduled_at": "2025-10-27T10:00:00Z"
  }'
```

### Reschedule PENDING Notification
```bash
curl -X PATCH http://localhost:8080/notifications/123 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "scheduled_at": "2025-10-27T15:00:00Z"
  }'
```

**Note:** Only notifications with `PENDING` status can be rescheduled. The worker respects `scheduled_at` and will not process notifications before their scheduled time.

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

### Create Email Notification (Immediate)

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

### Create Scheduled Notification

```bash
curl -X POST http://localhost:8080/notifications \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Reminder",
    "content": "Your appointment is tomorrow",
    "channel_name": "sms",
    "meta": {"phone": "+1234567890", "carrier": "verizon"},
    "scheduled_at": "2025-10-27T10:00:00Z"
  }'
```

See `/docs/notification-examples.md` for more examples (SMS, Push).

## Development

### Run with Docker
```bash
# Start all services
docker-compose up --build

# Rebuild only API service
docker-compose up --build api

# Run tests inside container
docker-compose run --rm api go test ./...
```

### Local Development
```bash
# Regenerate Swagger
swag init -g cmd/api/main.go
# Or with Docker
docker-compose run --rm api swag init -g cmd/api/main.go

# Run tests
go test ./...

# Build binary
go build -o bin/api ./cmd/api
```

### Docker Commands
```bash
# View logs
docker-compose logs -f api
docker-compose logs -f mysql

# Access MySQL shell
docker-compose exec mysql mysql -uapp -papp notification

# Access API container shell
docker-compose exec api sh

# Restart services
docker-compose restart api

# Remove everything and start fresh
docker-compose down -v
docker-compose up --build
```

## Design Decisions

### Design Patterns

**Strategy Pattern (Channels)**  
Each notification channel (Email, SMS, Push) implements the `Channel` interface with `Send()`, `Validate()`, and `Name()` methods. This allows adding new channels without modifying the core notification logic. The `NotifierService` depends on the interface, not concrete implementations.

**Dependency Injection (Services)**  
Services receive their dependencies through constructors (`NewNotifierService`, `NewUserController`). Database connections, channel lists, and other services are injected, making the code testable and loosely coupled. No global state or singletons.

**Repository Pattern (Database Abstraction)**  
GORM acts as the repository layer, abstracting database operations. Services interact with models through the ORM, not raw SQL. This allows switching databases (SQLite for tests, MySQL for production) without changing business logic.

**Transactional Outbox Pattern (Reliable Messaging)**  
Guarantees eventual delivery of notifications. When a notification is created, both the `Notification` and `Outbox` records are inserted in the same transaction. A separate worker processes the outbox asynchronously, ensuring no message loss even if the sending fails. Includes automatic retries with exponential backoff.

### Additional Design Decisions

**Idempotency**: SHA-256 hash of (user_id + channel + title + content + meta) stored in `idempotency_key` prevents duplicate notifications from concurrent requests.

**Custom Errors**: Typed errors (`ErrInvalidChannel`, `ErrNotificationNotFound`, `ErrInvalidMetadata`) enable semantically correct HTTP status codes and clear error handling.

**JWT Authentication**: 30-day token expiration with reusable middleware. Token verification is centralized and injected via `AuthMiddleware`.

**Metadata Validation**: Each channel validates its required fields (`Validate` method) with clear, channel-specific error messages before creating the notification.

**Async Worker**: Processes outbox every 30 seconds (configurable). Uses batch claiming with transactional status updates to prevent race conditions between multiple workers.

**Scheduled Notifications**: Composite index on `(status, scheduled_at)` enables efficient querying of pending jobs. Worker respects `scheduled_at` and processes notifications only when their time arrives.



## Data Models

**User**: `id`, `name`, `email` (unique), `password` (bcrypt hashed), `created_at`

**Notification**: `id`, `user_id`, `title`, `content`, `channel_name`, `idempotency_key` (unique), `created_at`, `deleted_at` (soft delete)

**Outbox**: `id`, `notification_id`, `channel_name`, `payload_json`, `status` (PENDING/PROCESSING/SENT/FAILED), `attempts`, `max_attempts`, `last_error`, `next_attempt_at`, `scheduled_at`, `created_at`, `updated_at`

## Database Migrations

The project uses GORM's `AutoMigrate` feature. When you start the application, it automatically creates tables if they don't exist, adds new columns to existing tables and creates indexes defined in model tags


## Documentation

- Swagger UI: http://localhost:8080/swagger/index.html
- Detailed examples: `/docs/notification-examples.md`
- Channel schemas: `GET /notifications/channels/schemas`

