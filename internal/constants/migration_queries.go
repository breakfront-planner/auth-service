package constants

const (
	CreateUsersTable = `
    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        login VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMPTZ DEFAULT now(),
		updated_at TIMESTAMPTZ DEFAULT now()
    );`

	//nolint:gosec // G101: False positive - this is a SQL schema definition, not hardcoded credentials
	CreateRefreshTokensTable = `
    CREATE TABLE refresh_tokens (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    	token_hash VARCHAR(64) UNIQUE NOT NULL,
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMPTZ DEFAULT now(),
		expires_at TIMESTAMPTZ NOT NULL,
		revoked_at TIMESTAMPTZ
	);
	
	CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id 
	ON refresh_tokens(user_id);`

	CreateMigrationsTable = `
    CREATE TABLE IF NOT EXISTS schema_migrations (
        version VARCHAR(255) PRIMARY KEY,
        applied_at TIMESTAMPTZ DEFAULT NOW()
    );`
)
