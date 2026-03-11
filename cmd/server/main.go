package main

import (
	"tgss/internal/config"
	"tgss/internal/handlers"
	"tgss/internal/infra"
	"tgss/internal/rss"
	"tgss/internal/server"
	"tgss/internal/telegram"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			infra.NewLogger,
			config.Load,
			server.NewGin,
			server.NewRateLimiter,

			handlers.NewChannelHandler,
			handlers.NewAuthHandler,

			rss.NewRSSGenerator,
			rss.NewFeedCache,
		),
		telegram.Module,
		fx.Invoke(server.RegisterRoutes),
		fx.Invoke((server.RegisterRateLimiterCleanup)),
		fx.Invoke(rss.RegisterFeedCacheCleanup),
		fx.Logger(infra.NewFXLogger()),
	)

	app.Run()
}
