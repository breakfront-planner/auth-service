package grpc

import (
	"context"
	"errors"
	"testing"

	authv1 "github.com/breakfront-planner/proto/gen/go/auth/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/breakfront-planner/auth-service/internal/api/grpc/mocks"
	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/configs"
)

type HandlerTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	mockAuthService *mocks.MockIAuthService
	handler         *AuthHandler
}

func (s *HandlerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockAuthService = mocks.NewMockIAuthService(s.ctrl)
	s.handler = NewAuthHandler(s.mockAuthService, &configs.CredentialsConfig{
		LoginMinLen:    3,
		LoginMaxLen:    50,
		PasswordMinLen: 8,
		PasswordMaxLen: 72,
	})
}

func (s *HandlerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *HandlerTestSuite) TestLogout_SwallowsServiceError() {
	s.mockAuthService.EXPECT().
		Logout(gomock.Any()).
		Return(errors.New("token not found"))

	resp, err := s.handler.Logout(context.Background(), &authv1.TokenRequest{RefreshToken: "whatever"})

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
}

func (s *HandlerTestSuite) TestLogout_Success() {
	s.mockAuthService.EXPECT().
		Logout(gomock.Any()).
		Return(nil)

	resp, err := s.handler.Logout(context.Background(), &authv1.TokenRequest{RefreshToken: "whatever"})

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func TestMapError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{"invalid credentials", autherrors.ErrInvalidCredentials, codes.Unauthenticated},
		{"login taken", autherrors.ErrLoginTaken, codes.AlreadyExists},
		{"token expired", autherrors.ErrTokenExpired, codes.Unauthenticated},
		{"token type", autherrors.ErrTokenType, codes.Unauthenticated},
		{"invalid jwt", autherrors.ErrInvalidJWT, codes.Unauthenticated},
		{"token sign method", autherrors.ErrTokenSignMethod, codes.Unauthenticated},
		{"invalid user id", autherrors.ErrInvalidUserID, codes.Unauthenticated},
		{"unknown error", errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, ok := status.FromError(mapError(tt.err))
			require.True(t, ok)
			assert.Equal(t, tt.code, st.Code())
		})
	}
}
