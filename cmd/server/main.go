package main

import (
	"context"
	"log"
	"tgss/internal/config"
	"tgss/internal/infra"
	"tgss/internal/server"
	"tgss/internal/telegram"
	"time"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			infra.NewLogger,
			config.Load,
			server.NewGin,
		),
		telegram.Module,
		fx.Logger(infra.NewFXLogger()),
	)

	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		log.Fatalf("failed to start: %v", err)
	}

	app.Wait()
}
