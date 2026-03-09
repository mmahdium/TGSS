package telegram

import (
	"context"
	"net"
	"net/url"
	"tgss/internal/config"
	"time"

	"github.com/gotd/contrib/bg"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
	"golang.org/x/time/rate"
)

func NewTelegramClient(cfg *config.Config, logger *zap.Logger) *telegram.Client {
	opts := telegram.Options{
		DC:               2,
		DCList:           dcs.Prod(),
		Logger:           logger,
		SessionStorage:   &session.FileStorage{Path: cfg.SessionPath},
		DialTimeout:      config.DialTimeout,
		ExchangeTimeout:  config.ExchangeTimeout,
		MigrationTimeout: config.MigrationTimeout,

		Middlewares: []telegram.Middleware{
			ratelimit.New(rate.Every(time.Millisecond*500), 5),
		},
	}

	// TODO: Add pebble and bbolt https://github.com/gotd/td/blob/6f8e63c553210a2901f5c3586f6b88d6524fe9b3/examples/userbot/main.go#L106
	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			logger.Warn("Failed to parse proxy URL, ignoring proxy", zap.String("url", cfg.ProxyURL), zap.Error(err))
		} else {
			dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
			if err != nil {
				logger.Warn("Failed to create proxy dialer, ignoring proxy", zap.String("url", cfg.ProxyURL), zap.Error(err))
			} else {
				opts.Resolver = dcs.Plain(dcs.PlainOptions{
					Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
						return dialer.Dial(network, addr)
					},
				})
				logger.Info("Using proxy", zap.String("url", cfg.ProxyURL))
			}
		}
	}

	return telegram.NewClient(cfg.TgAppId, cfg.TgAppHash, opts)
}

func RunClient(lc fx.Lifecycle, client *telegram.Client, service *Service, logger *zap.Logger) {
	var stop func() error

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			s, err := bg.Connect(client)
			if err != nil {
				return err
			}

			stop = s
			logger.Info("telegram client connected")

			if err := service.InitAuthStatus(ctx); err != nil {
				logger.Warn("Failed to initialize Telegram auth status", zap.Error(err))
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if stop != nil {
				logger.Info("telegram client stopping")
				return stop()
			}
			return nil
		},
	})
}

var Module = fx.Options(
	fx.Provide(
		NewTelegramClient,
		NewService,
	),
	fx.Invoke(RunClient),
)
