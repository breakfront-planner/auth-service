package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/breakfront-planner/auth-service/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx)
	if err != nil {
		log.Println("error while inject dependencies: ", err)
		os.Exit(1)
	}

	if err := application.Start(ctx); err != nil {
		log.Println("running the program: ", err)
		os.Exit(1)
	}

	closeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Close(closeCtx); err != nil {
		log.Println("closing the program: ", err)
	}
	log.Println("application stopped")
}
