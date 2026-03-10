package handlers

import (
	"context"
	"encoding/xml"
	"strconv"
	"tgss/internal/config"
	"tgss/internal/rss"
	"tgss/internal/telegram"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ChannelHandler struct {
	logger    *zap.Logger
	tgService *telegram.Service
	rss       *rss.RSSGenerator
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

func (ch *ChannelHandler) GetMessagesJson(c *gin.Context) {
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

	ctx, cancel := context.WithTimeout(context.Background(), config.MessageOperationTimeout)
	defer cancel()

	messages, err := ch.tgService.LastMessages(ctx, channelId, limit)
	if err != nil {
		ch.logger.Error("Error getting channel messages:", zap.Error(err))
		c.JSON(500, gin.H{"error": "Unable to get channel messages"})
		return
	}

	c.JSON(200, messages)
}

func (ch *ChannelHandler) GetMessagesRSS(c *gin.Context) {
	authStatCtx, cancel := context.WithTimeout(context.Background(), config.AuthStatusTimeout)
	defer cancel()

	authStat, err := ch.tgService.AuthStatus(authStatCtx)
	if err != nil || !authStat {
		c.JSON(500, gin.H{"error": "Telegram client is not initialized"})
		return
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), config.MessageOperationTimeout)
	defer cancel()

	messages, err := ch.tgService.LastMessages(ctx, channelId, limit)
	if err != nil {
		ch.logger.Error("Error getting channel messages:", zap.Error(err))
		c.JSON(500, gin.H{"error": "Unable to get channel messages"})
		return
	}

	rssFeed := ch.rss.GenerateFeed(messages, channelId)

	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	c.Header("X-Content-Type-Options", "nosniff")

	parsedFeed, err := xml.MarshalIndent(rssFeed, " ", "  ")
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to parse XML"})
		return
	}

	c.String(200, xml.Header+string(parsedFeed))

}
