package infra

import "go.uber.org/zap"

func NewLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

// Simple logger for fx
type FXLogger struct {
	logger *zap.Logger
}

func NewFXLogger() *FXLogger {
	logger, _ := zap.NewDevelopment()
	return &FXLogger{logger: logger}
}

func (l *FXLogger) Printf(str string, args ...any) {
	l.logger.Sugar().Infof(str, args...)
}