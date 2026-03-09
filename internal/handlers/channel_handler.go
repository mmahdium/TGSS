package handlers

import (
	"context"
	"tgss/internal/telegram"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ChannelHandler struct {
	logger    *zap.Logger
	tgService *telegram.Service
}

func NewChannelHandler(
	tgService *telegram.Service,
	logger *zap.Logger,
) *ChannelHandler {
	return &ChannelHandler{
		logger:    logger,
		tgService: tgService,
	}
}

func (ch *ChannelHandler) GetLatestMessages(c *gin.Context) {
	chennelId := c.Param("id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	m, err := ch.tgService.LastMessages(ctx, chennelId, 10)
	if err != nil {
		ch.logger.Error("Error getting channel messages:", zap.Error(err))
	}
	c.JSON(200, m)
}
