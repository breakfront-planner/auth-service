package repositories

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"fmt"
	"os"
	"time"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/database"

	"github.com/joho/godotenv"
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

	s.RefreshDuration, err = time.ParseDuration(os.Getenv("REFRESH_DURATION"))
	if err != nil {
		s.RefreshDuration = 15 * time.Minute
	}
	s.TokenHashedValue = os.Getenv("TOKEN_HASHED_VALUE")

	// Connection string for test database
	requiredEnvVars := []string{"TEST_DB_HOST", "TEST_DB_PORT", "TEST_DB_USER", "TEST_DB_PASSWORD", "TEST_DB_NAME", "TEST_DB_SSLMODE"}

	envVars := make(map[string]string)
	var missingVars []string
	for _, varName := range requiredEnvVars {
		if os.Getenv(varName) == "" {
			missingVars = append(missingVars, varName)
		} else {
			envVars[varName] = os.Getenv(varName)
		}
	}

	if len(missingVars) > 0 {
		err := autherrors.ErrMissingEnvVars(missingVars)
		require.NoError(s.T(), err, "Failed to get required credentials to test database")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		envVars["TEST_DB_HOST"],
		envVars["TEST_DB_PORT"],
		envVars["TEST_DB_USER"],
		envVars["TEST_DB_PASSWORD"],
		envVars["TEST_DB_NAME"],
		envVars["TEST_DB_SSLMODE"])

	db, err := sql.Open("postgres", dsn)
	require.NoError(s.T(), err, "Failed to connect to test database")

	// Verify connection
	err = db.Ping()
	require.NoError(s.T(), err, "Failed to ping test database")

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
