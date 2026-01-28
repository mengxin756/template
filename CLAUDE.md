# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project based on Clean Architecture principles with a modern technology stack. The project implements a user management system with HTTP APIs, database persistence, background job processing, and follows best practices for maintainable Go applications.

### Architecture Layers

1. **Handler Layer** (`internal/handler`) - HTTP request/response handling with Gin framework
2. **Service Layer** (`internal/service`) - Business logic implementation
3. **Repository Layer** (`internal/repository`) - Data access interfaces and implementations
4. **Domain Layer** (`internal/domain`) - Core business entities and interfaces
5. **Data Layer** (`internal/data`) - Database and external service integrations
6. **Infrastructure Layer** (`internal/server`, `internal/config`) - HTTP server setup, configuration management

### Key Technologies

- **Web Framework**: Gin
- **ORM**: Ent (Facebook's entity framework)
- **Dependency Injection**: Google Wire
- **Logging**: Zerolog with structured logging
- **Configuration**: Viper with YAML and environment variable support
- **Database**: MySQL (default, configurable to SQLite/PostgreSQL)
- **Task Queue**: Asynq (Redis-based background jobs)
- **Testing**: testify framework

## Common Development Commands

### Running the Application

```bash
# Run the API server
go run ./cmd/api
task run:api

# Run Asynq worker
go run ./cmd/asynq -mode worker

# Run Asynq scheduler
go run ./cmd/asynq -mode scheduler
```

### Code Generation

```bash
# Generate Ent ORM code from schema
go run entgo.io/ent/cmd/ent generate ./internal/data/ent/schema
task gen:ent
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./internal/service

# Run tests with coverage
go test -cover ./...
```

### Building

```bash
# Build the API binary
go build ./cmd/api
```

### Dependency Management

```bash
# Update dependencies
go mod tidy
```

## Project Structure

```
template/
├── cmd/                    # Application entry points
│   ├── api/               # HTTP API server
│   └── asynq/             # Asynq worker and scheduler
├── internal/              # Core business logic (private packages)
│   ├── config/           # Configuration management
│   ├── data/             # Data layer (database, Redis)
│   ├── domain/           # Domain models and interfaces
│   ├── handler/          # HTTP handlers (Gin controllers)
│   ├── job/              # Background jobs (Asynq)
│   ├── logger/           # Internal logger wrapper
│   ├── repository/       # Data access layer
│   ├── server/           # HTTP server setup
│   ├── service/          # Business logic
│   └── wire/             # Dependency injection (Google Wire)
├── pkg/                   # Public utility packages
│   ├── errors/           # Custom error types
│   ├── logger/           # Public logger utilities
│   └── response/         # HTTP response formatting
├── config/                # Configuration files
├── api/                   # API definitions
└── test/                  # Test files
```

## Configuration

The application uses YAML configuration (`config/config.yaml`) with environment variable overrides. Key configuration sections include:

- HTTP server settings (address, timeouts)
- Database connection (MySQL by default)
- Redis connection for caching and task queues
- Asynq task queue configuration
- Logging settings

Environment variables follow the pattern `SECTION_KEY` (e.g., `HTTP_ADDRESS`, `DB_DRIVER`).

## Infrastructure

Docker Compose is used for local development with:
- MySQL 8.0 (port 3307)
- Redis 7.2 (port 6379)

Start dependencies with:
```bash
docker-compose up -d
```

## API Endpoints

User management APIs:
- POST `/api/v1/users` - Register user
- GET `/api/v1/users` - List users (paginated)
- GET `/api/v1/users/:id` - Get user by ID
- PUT `/api/v1/users/:id` - Update user
- DELETE `/api/v1/users/:id` - Delete user
- PATCH `/api/v1/users/:id/status` - Change user status

## Known Issues

From README.md:
- Logger package (`pkg/logger`) has zerolog API usage issues
- Ent code generation needs regeneration
- Partial dependency injection configuration requires improvement

## Development Workflow

1. Start dependencies: `docker-compose up -d`
2. Generate Ent code if schema changed: `task gen:ent`
3. Run tests: `go test ./...`
4. Run application: `task run:api`
5. For background jobs: `go run ./cmd/asynq -mode worker`