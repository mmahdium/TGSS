package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"tgss/internal/config"
	"tgss/internal/telegram"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ImageHandler struct {
	logger       *zap.Logger
	imageService *telegram.ImageService
}

func NewImageHandler(
	imageService *telegram.ImageService,
	logger *zap.Logger,
) *ImageHandler {
	return &ImageHandler{
		logger:       logger,
		imageService: imageService,
	}
}

func (ih *ImageHandler) GetImage(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), config.AuthStatusTimeout)
	defer cancel()

	channelId := c.Param("channelId")
	if channelId == "" {
		c.JSON(400, gin.H{"error": "invalid channel ID or message ID"})
		return
	}

	msgIdStr := c.Param("messageId")
	msgId, err := strconv.Atoi(msgIdStr)
	if err != nil {
		ih.logger.Error("message Id is invalid", zap.Error(err))
		c.JSON(400, gin.H{"error": "message Id is invalid"})
		return
	}

	reader, err := ih.imageService.GetChannelsMessageImageById(ctx, channelId, msgId)
	if err != nil {
		ih.logger.Error("failed to get image", zap.Error(err))
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer reader.(io.ReadCloser).Close()

	if closer, ok := reader.(io.ReadCloser); ok {
		defer closer.Close()
	}

	extraHeaders := map[string]string{
		"Cache-Control": "public, max-age=31536000, immutable",
		"Expires":       time.Now().Add(1 * 365 * 24 * time.Hour).Format(http.TimeFormat),
	}

	c.DataFromReader(http.StatusOK, -1, "image/jpeg", reader, extraHeaders)
}
