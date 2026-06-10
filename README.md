# Auth Service

JWT-based authentication service for Breakfront Planner with token rotation and secure credential management.

## Status
**In Development** - Core authentication, service layers, and HTTP API handlers complete with comprehensive test coverage.

## Documentation

- [High-Level Design](https://scandalous-speedwell-5d7.notion.site/HLD-Breakfront-Planner-2c101219b91b8012b56dd6b3ac617e39) - System architecture and design
- [ADR-001: JWT Access & Refresh Tokens](https://scandalous-speedwell-5d7.notion.site/ADR-001-JWT-Access-Refresh-Tokens-2d401219b91b8028a9b5e5324b049c0a) - Architecture decision record for token-based authentication
- [Software Requirements Specification](https://scandalous-speedwell-5d7.notion.site/SRS-2c101219b91b805eab9cff45aa372683) - Complete system requirements

## Architecture

Layered Architecture + Repository Pattern
├── API Layer
├── Application/Service Layer
├── Repository Layer
└── Database Layer

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

## API Endpoints

| Method | Path        | Description                        | Request Body                     | Success Code |
|--------|-------------|------------------------------------|----------------------------------|--------------|
| POST   | `/login`    | Authenticate user                  | `{"login", "password"}`          | 200          |
| POST   | `/register` | Register new user                  | `{"login", "password"}`          | 200          |
| POST   | `/refresh`  | Refresh token pair                 | `{"refresh_token"}`              | 200          |
| POST   | `/logout`   | Revoke refresh token               | `{"refresh_token"}`              | 204          |

### Response Formats

**Token pair** (login, register, refresh):
```json
{"access_token": "...", "refresh_token": "..."}
```

**Error**:
```json
{"error": "error message"}
```

### Error Handling
- Client errors (400, 401, 409) return descriptive messages
- Internal errors (500) return a generic `"internal server error"` — details are logged server-side via `slog`
- Token-related errors (expired, invalid, wrong type) return `"unauthorized"` without leaking specifics
- Logout always returns 204, regardless of token validity

### Input Validation
Registration credentials are validated against configurable limits from `validation_config.json`:
- Login length: 3–50 characters
- Password length: 8–72 characters

Login endpoint only checks that fields are non-empty (business validation happens in the service layer).

## Core Components

### API Layer
- **AuthHandler**: HTTP handlers for all authentication endpoints
- **DTOs**: Request/response types (`CredentialsRequest`, `TokenRequest`, `TokenPairResponse`, `ErrorResponse`)
- **Validation**: Credential format validation with configurable limits from JSON config

### Service Layer
- **AuthService**: Coordinates user authentication operations (register, login, refresh, logout)
- **UserService**: Manages user accounts and password verification
- **TokenService**: Handles token lifecycle (creation, validation, rotation, revocation)
- **HashService**: Provides password and token hashing using bcrypt and SHA-256

### Validators
- **TokenValidator**: Flexible token validation with Functional Options pattern
  - Always validates signature and expiration (security requirement)
  - Optional validations: token type (access/refresh), user existence
  - Enables reusable validation logic across services and future middleware
  - Example:
    ```go
    // Validate with all checks
    parsed, err := validator.ValidateRefreshToken(token)

    // Validate with custom options
    parsed, err := validator.Validate(token,
        WithTokenType(constants.TokenTypeAccess),
        WithUserExistenceCheck())
    ```

### Repository Layer
- **UserRepository**: Database operations for user management with flexible filtering
- **TokenRepository**: Token persistence and validation
- **Filter System**: Generic reflection-based filter parser for dynamic query building

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

## Filter System

The repository layer uses a generic reflection-based filter parser for flexible query building:

- **Type-safe filters**: Uses struct tags (`db:"column_name"`) to map fields to database columns
- **Pointer-based fields**: Only non-nil pointer fields are included in queries
- **Dynamic query generation**: Builds SQL WHERE clauses automatically from filter structs
- **Validation**: Ensures all filter fields are pointers and at least one field is populated
- **Example usage**:
  ```go
  filter := models.UserFilter{
      Login: &userLogin,  // Only search by login
      ID: nil,            // ID is ignored
  }
  user, err := userRepo.FindUser(&filter)
  ```

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

#### Unit Tests (Service, Validator & API Layers)
Tests use `gomock` for dependency injection:
```bash
# Run all service unit tests
go test -v ./internal/services

# Run validator unit tests
go test -v ./internal/validators

# Run handler unit tests
go test -v ./internal/api

# Run specific test suites
go test -v ./internal/services -run TestAuthServiceTestSuite
go test -v ./internal/services -run TestUserServiceTestSuite
go test -v ./internal/services -run TestTokenServiceTestSuite
go test -v ./internal/services -run TestHashServiceTestSuite
go test -v ./internal/validators -run TestTokenValidatorTestSuite
go test -v ./internal/api -run TestHandlersTestSuite

# Generate mocks (when interfaces change)
go generate ./internal/services/mocks/...
go generate ./internal/validators/mocks/...
go generate ./internal/api/mocks/...
```

#### Test Coverage Summary
- **109 total tests** across repository, service, validator, and API layers
- **Integration tests (9)**: User and token repository operations with filter validation
- **Unit tests (89)**: Authentication flows, token lifecycle, password hashing, token validation, HTTP handlers
  - Service layer (62 tests): AuthService, UserService, TokenService, HashService
  - Validator layer (10 tests): TokenValidator with various validation options
  - API layer (17 tests): Login, Register, Refresh, Logout handlers (success, validation, errors)
- **Filter unit tests (11)**: Reflection-based filter parsing, validation, and error handling
- Test scenarios include: success paths, error handling, edge cases, and security validations

## Deployment

The service includes a Docker Compose configuration for PostgreSQL.

## Roadmap

### Completed
- [x] Core authentication logic (register, login, refresh, logout)
- [x] Password hashing with bcrypt
- [x] JWT token generation and validation
- [x] Token rotation and revocation
- [x] Flexible token validator with Functional Options pattern
- [x] Database repositories with PostgreSQL
- [x] Generic filter system with reflection-based parsing
- [x] HTTP handlers and REST API endpoints
- [x] Input validation with configurable limits (JSON config)
- [x] Error handling: generic client responses, structured server-side logging
- [x] Comprehensive test suite (109 tests)
- [x] Mock generation for unit testing

### In Progress
- [ ] API documentation (OpenAPI/Swagger)

### Planned
- [ ] Password strength requirements and validation
- [ ] Rate limiting for authentication endpoints
- [ ] Account verification and password recovery (email integration)
- [ ] Observability tools (Prometheus, Grafana, Thanos)
- [ ] Docker containerization for service deployment

## License

See [LICENSE](LICENSE) file for details.
