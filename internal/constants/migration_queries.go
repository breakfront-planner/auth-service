package constants

const (
	CreateUsersTable = `
    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        login VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMPTZ DEFAULT now(),
		updated_at TIMESTAMP DEFAULT now()
    );`

	CreateRefreshTokensTable = `
    CREATE TABLE refresh_tokens (
		token VARCHAR(255) PRIMARY KEY,
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT now()
	);`

	CreateMigrationsTable = `
    CREATE TABLE IF NOT EXISTS schema_migrations (
        version VARCHAR(255) PRIMARY KEY,
        applied_at TIMESTAMPTZ DEFAULT NOW()
    );`
)
