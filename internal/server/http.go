package server

import (
	"context"
	"tgss/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type RouterParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	GinEngine *gin.Engine
	Logger    *zap.Logger

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

			// params.GinEngine.GET("/channel/:id/json", params.ChannelHandler.GetMessagesJson)
			params.GinEngine.GET("/channel/:id", params.ChannelHandler.GetMessagesRSS)

			params.GinEngine.POST("/auth/send-code", params.AuthHandler.SendCode)
			params.GinEngine.POST("/auth/verify", params.AuthHandler.Verify)
			params.GinEngine.POST("/auth/password", params.AuthHandler.Password)

			params.Logger.Info("Starting server")
			go params.GinEngine.Run(":3000")
			params.Logger.Info("Server is running on port 8080")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("Stopping server")
			return nil
		},
	})
}
