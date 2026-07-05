package repositories

import (
	"database/sql"
	"os"
	"time"

	"github.com/breakfront-planner/auth-service/internal/configs"
	"github.com/breakfront-planner/auth-service/internal/database"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RepositoryTestSuite struct {
	suite.Suite
	DB               *sql.DB
	UserRepo         *UserRepository
	TokenRepo        *TokenRepository
	TestLogin        string
	TestPassword     string
	RefreshDuration  time.Duration
	TokenHashedValue string
}

func (s *RepositoryTestSuite) SetupSuite() {

	err := godotenv.Load("../../.env.test")
	require.NoError(s.T(), err, "Error loading .env file")

	s.TestLogin = os.Getenv("TEST_LOGIN")
	s.TestPassword = os.Getenv("TEST_PASS")

	require.NotEmpty(s.T(), s.TestLogin, "TEST_LOGIN must be set in .env.test")
	require.NotEmpty(s.T(), s.TestPassword, "TEST_PASS must be set in .env.test")

	s.RefreshDuration, err = time.ParseDuration(os.Getenv("REFRESH_DURATION"))
	if err != nil {
		s.RefreshDuration = 15 * time.Minute
	}
	s.TokenHashedValue = os.Getenv("TOKEN_HASHED_VALUE")

	require.NotEmpty(s.T(), s.TokenHashedValue, "TEST_HASH must be set in .env.test")

	// Connection config for test database, read from TEST_DB_* variables
	pgCfg := &configs.PostgresConfig{}
	err = env.ParseWithOptions(pgCfg, env.Options{Prefix: "TEST_"})
	require.NoError(s.T(), err, "Failed to parse test database configuration")

	db, err := database.Connect(*pgCfg)
	require.NoError(s.T(), err, "Failed to connect to test database")

	// Run migrations
	err = database.RunMigrations(db)
	require.NoError(s.T(), err, "Failed to run migrations")

	s.DB = db

	// Initialize all repositories
	s.UserRepo = NewUserRepository(db)
	s.TokenRepo = NewTokenRepository(db)

}

func (s *RepositoryTestSuite) TearDownTest() {
	_, err := s.DB.Exec("DELETE FROM refresh_tokens")
	require.NoError(s.T(), err, "Failed to cleanup refresh_tokens")

	_, err = s.DB.Exec("DELETE FROM users")
	require.NoError(s.T(), err, "Failed to cleanup users")
}

func (s *RepositoryTestSuite) TearDownSuite() {

	if s.DB != nil {
		err := s.DB.Close()
		require.NoError(s.T(), err, "Failed to close database connection")
	}
}
