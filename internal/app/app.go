package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/docker/docker/daemon/config"
	"google.golang.org/grpc"
)

type Application struct {
	conf   *configs.Configuration
	logger *slog.Logger
	server *grpc.Server
}

func New(ctx context.Context) (*Application, error) {
	app := &Application{
		server: *grpc.Server,
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}

	if err := app.setConfig(); err != nil {
		return nil, fmt.Errorf("set configuration: %w", err)
	}

	if err := app.setServer(ctx); err != nil {

		return nil, fmt.Errorf("set server: %w", err)
	}

	return app, nil
}

func (a *Application) setConfig() error {
	conf, err := config.New()
	if err != nil {
		return fmt.Errorf("new configuration: %w", err)
	}

	a.conf = conf
	return nil
}

func (a *Application) setServer(ctx context.Context) error {
}

func (a *Application) Start(ctx context.Context) error {
	return a.httpServer.Start(ctx)
}

func (a *Application) Close(ctx context.Context) error {
	errs := []error{
		a.httpServer.Close(ctx),
	}

	return errors.Join(errs...)
}
