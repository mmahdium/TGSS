package telegram

import (
	"context"
	"net"
	"net/url"
	"tgss/internal/config"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
)

func NewTelegramClient(cfg *config.Config, logger *zap.Logger) *telegram.Client {
	opts := telegram.Options{
		DC:             2,
		DCList:         dcs.Prod(),
		Logger:         logger,
		SessionStorage: &session.FileStorage{Path: cfg.SessionPath},
		DialTimeout: 20*time.Second,
		ExchangeTimeout: 20*time.Second,
		MigrationTimeout: 20*time.Second,
	}

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

func NewService(client *telegram.Client, logger *zap.Logger) *Service {
	return &Service{client: client, log: logger}
}

func RunClient(lc fx.Lifecycle, client *telegram.Client, cfg *config.Config, logger *zap.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			errCh := make(chan error, 1)
			readyCh := make(chan struct{})

			go func() {
				if err := client.Run(ctx, func(ctx context.Context) error {
					if _, err := client.Auth().Bot(ctx, cfg.TgBotToken); err != nil {
						return err
					}
					logger.Info("Telegram bot authenticated and running")
					close(readyCh)
					<-ctx.Done()
					return nil
				}); err != nil {
					errCh <- err
				}
			}()

			select {
			case <-readyCh:
				return nil
			case err := <-errCh:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		},
		OnStop: func(context.Context) error {
			cancel()
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
