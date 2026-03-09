package handlers

import (
	"context"
	"strconv"
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
	channelId := c.Param("id")
	limit := 5

	if limitStr := c.Query("limit"); limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit > 50 {
			ch.logger.Warn("Invalid limit value provided", zap.String("limit", limitStr))
			c.JSON(400, gin.H{"error": "Limit value provided is invalid or bigger than 50"})
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messages, err := ch.tgService.LastMessages(ctx, channelId, limit)
	if err != nil {
		ch.logger.Error("Error getting channel messages:", zap.Error(err))
		c.JSON(500, gin.H{"error": "Unable to get channel messages"})
		return
	}

	c.JSON(200, messages)
}
