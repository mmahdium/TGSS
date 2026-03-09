package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	TgPhoneNumber string
	TgAppId       int
	TgAppHash     string

	SessionPath string
	ProxyURL    string
}

func Load(logger *zap.Logger) *Config {
	// TODO: get app port and host
	err := godotenv.Load()
	if err != nil {
		logger.Warn("Error loading .env file - Using environment variables instead")
	}

	appId := os.Getenv("TG_APP_ID")
	if appId == "" {
		logger.Fatal("No APP ID provided")
	}

	appHash := os.Getenv("TG_APP_HASH")
	if appHash == "" {
		logger.Fatal("No APP HASH provided")
	}

	sessionPath := os.Getenv("TG_SESSION_PATH")
	if sessionPath == "" {
		sessionPath = "/tmp/tgss/"
		logger.Warn("No session path provided, using default path", zap.String("default_path", sessionPath))
	}

	if info, err := os.Stat(sessionPath); err == nil && !info.IsDir() {
		sessionPath = filepath.Dir(sessionPath)
	}
	if err := os.MkdirAll(sessionPath, 0o755); err != nil {
		logger.Fatal("Failed to create session directory", zap.String("path", sessionPath), zap.Error(err))
	}

	sessionFile := filepath.Join(sessionPath, "session.json")

	proxyURL := os.Getenv("TG_PROXY_URL")

	return &Config{
		TgAppId:       func() int { i, _ := strconv.Atoi(appId); return i }(),
		TgAppHash:     strings.TrimSpace(appHash),

		SessionPath: sessionFile,
		ProxyURL:    strings.TrimSpace(proxyURL),
	}
}
