package infra

import (
	"os"

	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	if os.Getenv("APP_ENV") == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

// Simple logger for fx
type FXLogger struct {
	logger *zap.Logger
}

func NewFXLogger() *FXLogger {
	if os.Getenv("APP_ENV") == "production" {
		logger, _ := zap.NewProduction()
		return &FXLogger{logger: logger}
	}
	logger, _ := zap.NewDevelopment()
	return &FXLogger{logger: logger}
}

func (l *FXLogger) Printf(str string, args ...any) {
	l.logger.Sugar().Infof(str, args...)
}
