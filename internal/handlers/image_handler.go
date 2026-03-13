package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"tgss/internal/config"
	"tgss/internal/telegram"
	"tgss/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ImageHandler struct {
	logger        *zap.Logger
	imageService  *telegram.ImageService
	hmacGenerator *utils.ImageHMACGenerator
}

func NewImageHandler(
	imageService *telegram.ImageService,
	logger *zap.Logger,
	hmacGenerator *utils.ImageHMACGenerator,
) *ImageHandler {
	return &ImageHandler{
		logger:        logger,
		imageService:  imageService,
		hmacGenerator: hmacGenerator,
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

	expStr := c.Query("exp")
	sig := c.Query("sig")

	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	expiresAt := time.Unix(expUnix, 0)

	if !ih.hmacGenerator.VerifyMAC(msgId, expiresAt, sig) {
		c.AbortWithStatus(403)
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
