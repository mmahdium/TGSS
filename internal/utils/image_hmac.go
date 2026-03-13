package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"tgss/internal/config"
	"time"

	"go.uber.org/zap"
)

type ImageHMACGenerator struct {
	config *config.Config
	logger *zap.Logger
}

func NewImageHMACGenerator(logger *zap.Logger, config *config.Config) *ImageHMACGenerator {
	return &ImageHMACGenerator{logger: logger, config: config}
}

func (h *ImageHMACGenerator) GenerateMAC(messageId int, expiresAt time.Time) string {
	mac := hmac.New(sha256.New, []byte(h.config.ImageSigSecret))

	msg := strconv.Itoa(messageId) + ":" + strconv.FormatInt(expiresAt.Unix(), 10)

	mac.Write([]byte(msg))

	return hex.EncodeToString(mac.Sum(nil))
}

func (h *ImageHMACGenerator) VerifyMAC(messageID int, expiresAt time.Time, sig string) bool {
	if time.Now().After(expiresAt) {
		return false
	}

	expected := h.GenerateMAC(messageID, expiresAt)

	return hmac.Equal([]byte(expected), []byte(sig))
}
