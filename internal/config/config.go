package config

import (
	"net"
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

	SessionPath    string
	ProxyURL       string
	AppHost        string
	AppPort        int
	BaseURL        string
	ImageSigSecret string
	AppEnv         string
}

func Load(logger *zap.Logger) *Config {
	err := godotenv.Load()
	if err != nil {
		logger.Warn("Error loading .env file - Using environment variables instead")
	}

	appHost := os.Getenv("APP_HOST")
	if appHost == "" {
		appHost = "0.0.0.0"
		logger.Warn("No APP_HOST provided, using default", zap.String("default_host", appHost))
	}
	if ip := net.ParseIP(appHost); ip == nil {
		logger.Fatal("Invalid APP_HOST format", zap.String("host", appHost))
	}

	appPortStr := os.Getenv("APP_PORT")
	if appPortStr == "" {
		appPortStr = "3000"
		logger.Warn("No APP_PORT provided, using default", zap.String("default_port", appPortStr))
	}
	appPort, err := strconv.Atoi(strings.TrimSpace(appPortStr))
	if err != nil || appPort < 1 || appPort > 65535 {
		logger.Fatal("Invalid APP_PORT value, must be between 1 and 65535", zap.String("port", appPortStr))
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

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		logger.Fatal("BASE_URL cant be empty")
	}
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	imageSigSecret := os.Getenv("IMAGE_SIGNATURE_SECRET")
	if imageSigSecret == "" {
		logger.Warn("IMAGE_SIGNATURE_SECRET is not set, which leaves a backlink vulnerability on images")
		imageSigSecret = "notAsafeSecreT"
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}
	// TODO: logging and gin

	return &Config{
		TgAppId:   func() int { i, _ := strconv.Atoi(appId); return i }(),
		TgAppHash: strings.TrimSpace(appHash),

		SessionPath: sessionFile,
		ProxyURL:    strings.TrimSpace(proxyURL),
		AppHost:     appHost,
		AppPort:     appPort,
		BaseURL:     baseURL,
		AppEnv:      appEnv,
	}
}
