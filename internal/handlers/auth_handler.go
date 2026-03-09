package handlers

import (
	"context"
	"net/http"
	"tgss/internal/config"

	"tgss/internal/telegram"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	logger *zap.Logger
	tg     *telegram.Service
}

func NewAuthHandler(tg *telegram.Service, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{tg: tg, logger: logger}
}

type phoneReq struct {
	Phone string `json:"phone"`
}

type codeReq struct {
	Code string `json:"code"`
}

type passwordReq struct {
	Password string `json:"password"`
}

func (h *AuthHandler) SendCode(c *gin.Context) {
	var req phoneReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.AuthOperationTimeout)
	defer cancel()

	if err := h.tg.SendCode(ctx, req.Phone); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "code sent"})
}

func (h *AuthHandler) Verify(c *gin.Context) {
	var req codeReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.AuthOperationTimeout)
	defer cancel()

	if err := h.tg.VerifyCode(ctx, req.Code); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "authenticated"})
}

func (h *AuthHandler) Password(c *gin.Context) {
	var req passwordReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.AuthOperationTimeout)
	defer cancel()

	if err := h.tg.Password(ctx, req.Password); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "authenticated"})
}
