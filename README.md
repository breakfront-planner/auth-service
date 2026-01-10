# Auth Service

JWT-based authentication service for Breakfront Planner with token rotation and secure credential management.

## Status
**In Development** - Core authentication and service layers complete with comprehensive test coverage. HTTP handlers and API layer in progress.

## Documentation

- [High-Level Design](https://scandalous-speedwell-5d7.notion.site/HLD-Breakfront-Planner-2c101219b91b8012b56dd6b3ac617e39) - System architecture and design
- [ADR-001: JWT Access & Refresh Tokens](https://scandalous-speedwell-5d7.notion.site/ADR-001-JWT-Access-Refresh-Tokens-2d401219b91b8028a9b5e5324b049c0a) - Architecture decision record for token-based authentication
- [Software Requirements Specification](https://scandalous-speedwell-5d7.notion.site/SRS-2c101219b91b805eab9cff45aa372683) - Complete system requirements

## Architecture

### Tech Stack
- **Language**: Go 1.24.5
- **Database**: PostgreSQL 15
- **Key Libraries**:
  - `golang-jwt/jwt/v5` - JWT token generation & validation (HS256)
  - `golang.org/x/crypto/bcrypt` - Password hashing with automatic salt
  - `lib/pq` - PostgreSQL driver
  - `google/uuid` - UUID generation
  - `joho/godotenv` - Environment variable management
  - `testify/suite` - Test framework with setup/teardown support
  - `gomock` - Mock generation for unit testing

## Core Components

### Service Layer
- **AuthService**: Coordinates user authentication operations (register, login, refresh, logout)
- **UserService**: Manages user accounts and password verification
- **TokenService**: Handles token lifecycle (creation, validation, rotation, revocation)
- **HashService**: Provides password and token hashing using bcrypt and SHA-256

### Repository Layer
- **UserRepository**: Database operations for user management
- **TokenRepository**: Token persistence and validation

### JWT Manager
- Generates access and refresh tokens with configurable expiration
- Includes user ID, token type, expiration, and JTI (unique identifier) in claims

## Authentication Flow

### Token Types
- **Access Token**
  - Short-lived: 10 minutes (configurable)
  - Sent in `Authorization: Bearer <token>` header
  - Used for API requests
  - Stateless JWT

- **Refresh Token**
  - Long-lived: 1 hour - 30 days (configurable)
  - Stored hashed in the database (SHA-256)
  - Used to obtain new access & refresh tokens
  - Supports rotation for security

## Security Features

- **Password Hashing**: bcrypt (cost factor 10) with automatic salt generation
- **Token Hashing**: Refresh tokens hashed with SHA-256 before database storage
- **Token Rotation**: Old refresh tokens automatically revoked on successful refresh
- **Short-lived Access Tokens**: Minimize exposure window (default 10 minutes)
- **Expiration Validation**: Tokens checked against `expires_at` timestamp
- **Revocation Support**: Soft delete via `revoked_at` field with database validation
- **Unique Token IDs**: JTI (JWT ID) claim for token tracking and replay prevention

## Development

### Prerequisites
- Go 1.24.5+
- PostgreSQL 15+
- Docker & Docker Compose (optional)

### Setup

1. Clone the repository

2. Create `.env` file in the root directory:
   ```env
   DB_HOST=
   DB_PORT=
   DB_NAME=
   DB_USER=
   DB_PASSWORD=

   JWT_SECRET=
   ACCESS_TOKEN_DURATION=
   REFRESH_TOKEN_DURATION=

   ```

3. Start PostgreSQL:
   ```bash
   docker-compose up -d
   ```

4. Run the service:
   ```bash
   go run cmd/main.go
   ```

### Database Schema

Database migrations are managed in [migration_queries.go](internal/constants/migration_queries.go). Schema includes:
- `users` table with bcrypt password hashes
- `tokens` table with SHA-256 hashed values, expiration, and revocation tracking

### Testing

The project includes comprehensive test coverage with both integration and unit tests.

#### Integration Tests (Repository Layer)
Tests use a real PostgreSQL database with `testify/suite`:

```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Run all repository integration tests
go test -v ./internal/repositories

# Run specific test suites
go test -v ./internal/repositories -run TestUserRepositoryTestSuite
go test -v ./internal/repositories -run TestTokenRepositoryTestSuite

# Stop test database
docker-compose -f docker-compose.test.yml down
```

#### Unit Tests (Service Layer)
Tests use `gomock` for dependency injection:

```bash
# Run all service unit tests
go test -v ./internal/services

# Run specific test suites
go test -v ./internal/services -run TestAuthServiceTestSuite
go test -v ./internal/services -run TestUserServiceTestSuite
go test -v ./internal/services -run TestTokenServiceTestSuite
go test -v ./internal/services -run TestHashServiceTestSuite

# Generate mocks (when interfaces change)
go generate ./internal/services/mocks/...
```

#### Test Coverage Summary
- **71 total tests** across repository and service layers
- **Integration tests (9)**: User and token repository operations
- **Unit tests (62)**: Authentication flows, token lifecycle, password hashing
- Test scenarios include: success paths, error handling, edge cases, and security validations

## Deployment

The service includes a Docker Compose configuration for PostgreSQL.

## Project Structure

```
auth-service/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── autherrors/                # Custom error definitions
│   ├── configs/                   # Configuration management
│   ├── constants/                 # Constants and migration queries
│   ├── database/                  # Database connection and migrations
│   ├── jwt/                       # JWT token generation
│   ├── models/                    # Domain models (User, Token)
│   ├── repositories/              # Data access layer
│   │   ├── user_repository.go
│   │   ├── token_repository.go
│   │   └── *_test.go             # Integration tests
│   └── services/                  # Business logic layer
│       ├── auth_service.go        # Main authentication service
│       ├── user_service.go        # User management
│       ├── token_service.go       # Token lifecycle
│       ├── hash_service.go        # Hashing operations
│       ├── mocks/                 # Generated mocks for testing
│       └── *_test.go             # Unit tests
├── .env                          # Environment configuration
├── .env.test                     # Test environment configuration
├── docker-compose.yml            # PostgreSQL for development
└── docker-compose.test.yml       # PostgreSQL for testing
```

## Roadmap

### Completed
- [x] Core authentication logic (register, login, refresh, logout)
- [x] Password hashing with bcrypt
- [x] JWT token generation and validation
- [x] Token rotation and revocation
- [x] Database repositories with PostgreSQL
- [x] Comprehensive test suite (71 tests)
- [x] Mock generation for unit testing

### In Progress
- [ ] HTTP handlers and REST API endpoints
- [ ] Input validation middleware
- [ ] API documentation (OpenAPI/Swagger)

### Planned
- [ ] Password strength requirements and validation
- [ ] Rate limiting for authentication endpoints
- [ ] Account verification and password recovery (email integration)
- [ ] Observability tools (Prometheus, Grafana, Thanos)
- [ ] Docker containerization for service deployment

## License

See [LICENSE](LICENSE) file for details.
