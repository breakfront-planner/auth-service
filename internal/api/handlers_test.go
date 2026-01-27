package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/breakfront-planner/auth-service/internal/api/mocks"
	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/configs"
	"github.com/breakfront-planner/auth-service/internal/models"
)

type HandlersTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	mockAuthService *mocks.MockIAuthService
	handler         *AuthHandler
	testLogin       string
	testPassword    string
	testToken       string
}

func (s *HandlersTestSuite) SetupSuite() {
	err := godotenv.Load("../../.env.test")
	require.NoError(s.T(), err, "Failed to load .env.test")

	s.testLogin = os.Getenv("TEST_LOGIN")
	s.testPassword = os.Getenv("TEST_PASS")
	s.testToken = os.Getenv("TOKEN_TEST_VALUE")

	require.NotEmpty(s.T(), s.testLogin, "TEST_LOGIN must be set in .env.test")
	require.NotEmpty(s.T(), s.testPassword, "TEST_PASS must be set in .env.test")
	require.NotEmpty(s.T(), s.testToken, "TOKEN_TEST_VALUE must be set in .env.test")
}

func (s *HandlersTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockAuthService = mocks.NewMockIAuthService(s.ctrl)
	s.handler = NewAuthHandler(s.mockAuthService, &configs.CredentialsConfig{
		LoginMinLen:    3,
		LoginMaxLen:    50,
		PasswordMinLen: 8,
		PasswordMaxLen: 72,
	})
}

func (s *HandlersTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// --- Login ---

func (s *HandlersTestSuite) TestLoginSuccess() {
	s.mockAuthService.EXPECT().
		Login(s.testLogin, s.testPassword).
		Return(&models.Token{Value: "access-token"}, &models.Token{Value: "refresh-token"}, nil)

	body, _ := json.Marshal(CredentialsRequest{Login: s.testLogin, Password: s.testPassword})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp TokenPairResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "access-token", resp.AccessToken)
	assert.Equal(s.T(), "refresh-token", resp.RefreshToken)
}

func (s *HandlersTestSuite) TestLoginInvalidBody() {
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *HandlersTestSuite) TestLoginEmptyLogin() {
	body, _ := json.Marshal(CredentialsRequest{Login: "", Password: s.testPassword})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *HandlersTestSuite) TestLoginEmptyPassword() {
	body, _ := json.Marshal(CredentialsRequest{Login: s.testLogin, Password: ""})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *HandlersTestSuite) TestLoginEmptyBoth() {
	body := `{"login":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *HandlersTestSuite) TestLoginInvalidCredentials() {
	s.mockAuthService.EXPECT().
		Login(s.testLogin, s.testPassword).
		Return(nil, nil, autherrors.ErrInvalidCredentials)

	body, _ := json.Marshal(CredentialsRequest{Login: s.testLogin, Password: s.testPassword})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(s.T(), "invalid credentials", resp.Error)
}

func (s *HandlersTestSuite) TestLoginInternalError() {
	s.mockAuthService.EXPECT().
		Login(s.testLogin, s.testPassword).
		Return(nil, nil, assert.AnError)

	body, _ := json.Marshal(CredentialsRequest{Login: s.testLogin, Password: s.testPassword})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(s.T(), "internal server error", resp.Error)
}

// --- Register ---

func (s *HandlersTestSuite) TestRegisterSuccess() {
	s.mockAuthService.EXPECT().
		Register(s.testLogin, s.testPassword).
		Return(&models.Token{Value: "access-token"}, &models.Token{Value: "refresh-token"}, nil)

	body, _ := json.Marshal(CredentialsRequest{Login: s.testLogin, Password: s.testPassword})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Register(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp TokenPairResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(s.T(), "access-token", resp.AccessToken)
	assert.Equal(s.T(), "refresh-token", resp.RefreshToken)
}

func (s *HandlersTestSuite) TestRegisterLoginTooShort() {
	body, _ := json.Marshal(CredentialsRequest{Login: "short", Password: s.testPassword})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Register(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *HandlersTestSuite) TestRegisterPasswordTooShort() {
	body, _ := json.Marshal(CredentialsRequest{Login: s.testLogin, Password: "short"})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Register(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *HandlersTestSuite) TestRegisterLoginTaken() {
	s.mockAuthService.EXPECT().
		Register(s.testLogin, s.testPassword).
		Return(nil, nil, autherrors.ErrLoginTaken)

	body, _ := json.Marshal(CredentialsRequest{Login: s.testLogin, Password: s.testPassword})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Register(w, req)

	assert.Equal(s.T(), http.StatusConflict, w.Code)

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(s.T(), "login already taken", resp.Error)
}

// --- Refresh ---

func (s *HandlersTestSuite) TestRefreshSuccess() {
	s.mockAuthService.EXPECT().
		Refresh(s.testToken).
		Return(&models.Token{Value: "new-access"}, &models.Token{Value: "new-refresh"}, nil)

	body, _ := json.Marshal(TokenRequest{RefreshToken: s.testToken})
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Refresh(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp TokenPairResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(s.T(), "new-access", resp.AccessToken)
	assert.Equal(s.T(), "new-refresh", resp.RefreshToken)
}

func (s *HandlersTestSuite) TestRefreshInvalidToken() {
	s.mockAuthService.EXPECT().
		Refresh(s.testToken).
		Return(nil, nil, autherrors.ErrInvalidJWT)

	body, _ := json.Marshal(TokenRequest{RefreshToken: s.testToken})
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Refresh(w, req)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(s.T(), "unauthorized", resp.Error)
}

func (s *HandlersTestSuite) TestRefreshExpiredToken() {
	s.mockAuthService.EXPECT().
		Refresh(s.testToken).
		Return(nil, nil, autherrors.ErrTokenExpired)

	body, _ := json.Marshal(TokenRequest{RefreshToken: s.testToken})
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Refresh(w, req)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

// --- Logout ---

func (s *HandlersTestSuite) TestLogoutSuccess() {
	s.mockAuthService.EXPECT().
		Logout(s.testToken).
		Return(nil)

	body, _ := json.Marshal(TokenRequest{RefreshToken: s.testToken})
	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Logout(w, req)

	assert.Equal(s.T(), http.StatusNoContent, w.Code)
}

func (s *HandlersTestSuite) TestLogoutErrorStillReturns204() {
	s.mockAuthService.EXPECT().
		Logout(s.testToken).
		Return(autherrors.ErrInvalidJWT)

	body, _ := json.Marshal(TokenRequest{RefreshToken: s.testToken})
	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Logout(w, req)

	assert.Equal(s.T(), http.StatusNoContent, w.Code)
}

func (s *HandlersTestSuite) TestLogoutInvalidBody() {
	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	s.handler.Logout(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}
