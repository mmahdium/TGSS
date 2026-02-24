package server

import (
	"context"

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
}

func NewGin() *gin.Engine {
	g := gin.New()
	g.Use(gin.Recovery())
	g.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
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
			params.Logger.Info("Starting server")
			go params.GinEngine.Run(":8080")
			params.Logger.Info("Server is running on port 8080")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("Stopping server")
			return nil
		},
	})
}
