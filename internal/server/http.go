package server

import (
	"context"
	"strconv"
	"tgss/internal/config"
	"tgss/internal/handlers"
	"tgss/internal/telegram"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type RouterParams struct {
	fx.In

	Lifecycle   fx.Lifecycle
	GinEngine   *gin.Engine
	Logger      *zap.Logger
	TgService   *telegram.Service
	Config      *config.Config
	RateLimiter *RateLimiter

	ChannelHandler *handlers.ChannelHandler
	AuthHandler    *handlers.AuthHandler
}

func NewGin() *gin.Engine {
	g := gin.New()
	g.Use(gin.Recovery())
	g.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET"},
		AllowHeaders:    []string{"Origin", "Content-Type"},
	}))
	return g
}

func RegisterRoutes(params RouterParams) {
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.GinEngine.GET("/ping", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "pong",
				})
			})

			// params.GinEngine.GET("/feed/:id/json", params.ChannelHandler.GetMessagesJson)
			params.GinEngine.GET("/feed/:id", params.RateLimiter.RateLimit(), params.ChannelHandler.GetMessagesRSS)

			// TODO: improve accurecy in rss channel fields
			// TODO: add ui with templates under /setup with fetch
			// TODO: add gin level cache (look for higher limit)
			// TODO: image endpoint + hash and expiry
			authStatCtx, cancel := context.WithTimeout(context.Background(), config.AuthStatusTimeout)
			defer cancel()

			authStat, err := params.TgService.AuthStatus(authStatCtx)
			if err != nil {
				params.Logger.Fatal("Unable to get auth state from telegram", zap.Error(err))
			}
			if !authStat {
				params.Logger.Warn("You have not logged in to telegram, log in to use the api")
				params.GinEngine.POST("/auth/send-code", params.AuthHandler.SendCode)
				params.GinEngine.POST("/auth/verify", params.AuthHandler.Verify)
				params.GinEngine.POST("/auth/password", params.AuthHandler.Password)
			}

			params.Logger.Info("Starting server", zap.String("host", params.Config.AppHost), zap.Int("port", params.Config.AppPort))
			go params.GinEngine.Run(params.Config.AppHost + ":" + strconv.Itoa(params.Config.AppPort))
			params.Logger.Info("Server is running", zap.String("address", params.Config.AppHost+":"+strconv.Itoa(params.Config.AppPort)))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("Stopping server")
			return nil
		},
	})
}
