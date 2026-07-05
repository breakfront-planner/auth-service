package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/breakfront-planner/auth-service/internal/configs"
	"github.com/breakfront-planner/auth-service/internal/database"
	"github.com/breakfront-planner/auth-service/internal/jwt"
	"github.com/breakfront-planner/auth-service/internal/repositories"
	"github.com/breakfront-planner/auth-service/internal/services"
	"github.com/breakfront-planner/auth-service/internal/validators"

	"google.golang.org/grpc"
)

type Application struct {
	conf   *configs.Configuration
	logger *slog.Logger

	db *sql.DB

	userRepo  *repositories.UserRepository
	tokenRepo *repositories.TokenRepository

	jwtManager   *jwt.Manager
	hashService  *services.HashService
	userService  *services.UserService
	tokenService *services.TokenService
	validator    *validators.TokenValidator
	authService  *services.AuthService

	grpcServer *grpc.Server
	listener   net.Listener
}

func New(ctx context.Context) (*Application, error) {
	app := &Application{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}

	if err := app.setConfig(); err != nil {
		return nil, fmt.Errorf("set configuration: %w", err)
	}

	if err := app.setDatabase(); err != nil {
		return nil, fmt.Errorf("set database: %w", err)
	}

	app.setRepositories()

	if err := app.setServices(); err != nil {
		return nil, fmt.Errorf("set services: %w", err)
	}

	if err := app.setServer(); err != nil {
		return nil, fmt.Errorf("set server: %w", err)
	}

	return app, nil
}

func (a *Application) setConfig() error {
	conf, err := configs.New()
	if err != nil {
		return fmt.Errorf("new configuration: %w", err)
	}

	a.conf = conf
	return nil
}

func (a *Application) setDatabase() error {
	db, err := database.Connect(a.conf.Postgres)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	a.db = db

	if err := database.RunMigrations(db); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func (a *Application) setRepositories() {
	a.userRepo = repositories.NewUserRepository(a.db)
	a.tokenRepo = repositories.NewTokenRepository(a.db)
}

func (a *Application) setServices() error {
	a.jwtManager = jwt.NewManager(
		a.conf.Token.JWTSecret,
		a.conf.Token.AccessDuration,
		a.conf.Token.RefreshDuration,
	)

	a.hashService = services.NewHashService()
	a.userService = services.NewUserService(a.userRepo, a.hashService)
	a.tokenService = services.NewTokenService(a.tokenRepo, a.hashService, a.jwtManager)
	a.validator = validators.NewTokenValidator(a.jwtManager, a.userService)
	a.authService = services.NewAuthService(a.tokenService, a.userService, a.validator)

	return nil
}

func (a *Application) setServer() error {
	lis, err := net.Listen("tcp", a.conf.GRPCServer.GRPCServerAddress)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	a.listener = lis

	a.grpcServer = grpc.NewServer()
	// TODO: register generated AuthService handler once .proto/codegen exist:
	// authpb.RegisterAuthServiceServer(a.grpcServer, grpcapi.NewAuthHandler(a.authService, a.conf.Credentials))

	return nil
}

func (a *Application) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("grpc server starting", "addr", a.listener.Addr().String())
		if err := a.grpcServer.Serve(a.listener); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("grpc server: %w", err)
	case <-ctx.Done():
		return nil
	}
}

func (a *Application) Close(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-ctx.Done():
		a.grpcServer.Stop()
	}

	var errs []error
	if err := a.db.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close db: %w", err))
	}

	return errors.Join(errs...)
}
