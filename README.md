# Auth Service

JWT-based authentication service for Breakfront Planner with token rotation and secure credential management.

## Status
**WIP** - Core authentication logic implemented, HTTP handlers in progress

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

## Authentication Flow

### Token Types
- **Access Token**
  - Short-lived: 10 minutes
  - Sent in `Authorization: Bearer <token>` header
  - Used for API requests
  - Stateless JWT

- **Refresh Token**
  - Long-lived: 30 days
  - Stored hashed in the database
  - Used to obtain new access & refresh tokens
  - Supports rotation for security

## Security Features

- **Password Hashing**: bcrypt with automatic salt generation
- **Token Hashing**: Refresh tokens hashed before database storage
- **Token Rotation**: Old refresh tokens invalidated on rotation
- **Short-lived Access Tokens**: Minimize exposure window
- **Expiration Validation**: Tokens checked against `expires_at` timestamp
- **Revocation Support**: Soft delete via `revoked_at` field

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

Database migrations are managed in `internal/constants/migration_queries.go`

### Testing

TBD

```bash
go test ./...
```

## Deployment

The service includes a Docker Compose configuration for PostgreSQL.

## Roadmap

- [ ] HTTP handlers implementation
- [ ] Input validation
- [ ] Logout endpoint (token revocation)
- [ ] Password strength requirements
- [ ] Account verification, access recover (email)
- [ ] Observability tools (Prometheus, Graphana, Thanos)

## License

See [LICENSE](LICENSE) file for details.
