# Auth Service

JWT-based authentication service for Breakfront Planner with token rotation and secure credential management.

## Status
**In Development** — core auth logic, service/repository layers, and config/app bootstrap are done. The service is being migrated from a standalone HTTP API to a gRPC service (HTTP will move to a separate API gateway); gRPC handler registration is pending a `.proto` contract.

## Documentation

- [High-Level Design](https://scandalous-speedwell-5d7.notion.site/HLD-Breakfront-Planner-2c101219b91b8012b56dd6b3ac617e39)
- [ADR-001: JWT Access & Refresh Tokens](https://scandalous-speedwell-5d7.notion.site/ADR-001-JWT-Access-Refresh-Tokens-2d401219b91b8028a9b5e5324b049c0a)
- [Software Requirements Specification](https://scandalous-speedwell-5d7.notion.site/SRS-2c101219b91b805eab9cff45aa372683)

## Tech Stack
- **Language**: Go 1.24.5
- **Database**: PostgreSQL 15
- **Key libraries**: `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`, `lib/pq`, `google/uuid`, `caarlos0/env/v11` (config), `joho/godotenv`, `google.golang.org/grpc`, `testify/suite` + `gomock` (tests)

## Architecture

```
cmd/main.go        — entrypoint: signal handling, app.New/Start/Close
internal/app        — bootstrap: config → db → repositories → services → grpc server
internal/configs     — single Configuration struct, loaded from .env via env tags
internal/api         — HTTP handlers/router (legacy; moving to a separate gateway)
internal/services    — AuthService, UserService, TokenService, HashService
internal/validators  — TokenValidator (functional-options based)
internal/repositories— Postgres repositories + generic reflection-based filter parser
internal/jwt         — JWT generation/parsing
```

## Configuration

All configuration is loaded from a single `.env` file into `configs.Configuration` (env-tag based, see `internal/configs/config.go`):

| Variable | Default | Purpose |
|---|---|---|
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` | required | Postgres connection |
| `JWT_SECRET` | required | JWT signing secret |
| `ACCESS_TOKEN_DURATION` | `10m` | Access token TTL |
| `REFRESH_TOKEN_DURATION` | `48h` | Refresh token TTL |
| `LOGIN_MIN_LEN` / `LOGIN_MAX_LEN` | `3` / `32` | Login length bounds |
| `PASSWORD_MIN_LEN` / `PASSWORD_MAX_LEN` | `8` / `64` | Password length bounds |
| `GRPC_SERVER_ADDRESS` | `:50051` | gRPC listen address |

Missing required fields (JWT secret, Postgres creds) fail config loading at startup.

## Security Features

- Password hashing: bcrypt (cost 10) with automatic salt
- Refresh tokens hashed with SHA-256 before storage; rotated on every refresh
- Short-lived access tokens (default 10m), long-lived refresh tokens (default 48h)
- Revocation via `revoked_at`; unique JTI per token for replay tracking

## Development

### Prerequisites
- Go 1.24.5+, PostgreSQL 15+, Docker & Docker Compose (optional)

### Setup
1. Create a `.env` file (see [Configuration](#configuration) for all variables)
2. `docker-compose up -d` — start Postgres
3. `go run cmd/main.go` — run the service (applies migrations on startup)

Database migrations live in [migration_queries.go](internal/constants/migration_queries.go) (`users`, `refresh_tokens` tables).

### Testing
```bash
# Integration tests (repository layer) — needs the test DB
docker-compose -f docker-compose.test.yml up -d
go test -v ./internal/repositories
docker-compose -f docker-compose.test.yml down

# Unit tests (services, validators, HTTP handlers) — mocked, no DB needed
go test -v ./internal/services ./internal/validators ./internal/api

# Regenerate mocks after interface changes
go generate ./internal/services/mocks/... ./internal/validators/mocks/... ./internal/api/mocks/...
```

## Roadmap

- [x] Core auth (register/login/refresh/logout), password/token hashing, rotation, revocation
- [x] Postgres repositories with generic filter system
- [x] Config consolidation (single env-driven `Configuration`) and app bootstrap rewrite
- [ ] gRPC `.proto` contract + generated handler registration
- [ ] Move HTTP API into a separate gateway service
- [ ] Rate limiting, password strength rules, account recovery, observability

## License

See [LICENSE](LICENSE) file for details.
