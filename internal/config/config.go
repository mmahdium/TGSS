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
	TgBotToken    string // Only for getting authenticated without phone number
	TgAppId       int
	TgAppHash     string

	SessionPath string
}

func Load(logger *zap.Logger) *Config {
	err := godotenv.Load()
	if err != nil {
		logger.Warn("Error loading .env file - Using environment variables instead")
	}

	phoneNumber := os.Getenv("TG_PHONE_NUMBER")
	if phoneNumber == "" {
		logger.Fatal("No phone number provided")
	}

	botToken := os.Getenv("TG_BOT_TOKEN")
	if botToken == "" {
		logger.Fatal("No bot token provided")
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

	return &Config{
		TgPhoneNumber: strings.TrimSpace(phoneNumber),
		TgBotToken:    strings.TrimSpace(botToken),
		TgAppId:       func() int { i, _ := strconv.Atoi(appId); return i }(),
		TgAppHash:     strings.TrimSpace(appHash),

		SessionPath: sessionFile,
	}
}
