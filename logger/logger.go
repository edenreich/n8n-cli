// Package logger provides structured logging capabilities for the n8n CLI
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global logger instance
var logger *zap.SugaredLogger

// InitLogger initializes the global logger
func InitLogger(debug bool) {
	isDebug := debug || os.Getenv("DEBUG") == "1" || os.Getenv("DEBUG") == "true"

	var zapLogger *zap.Logger
	var err error

	if isDebug {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.TimeKey = "time"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		zapLogger, err = cfg.Build(
			zap.AddCaller(),
			zap.AddCallerSkip(1),
		)
		if err != nil {
			zapLogger, _ = zap.NewProduction()
		}

		logger = zapLogger.Sugar().Named("n8n-cli")
		logger.Debug("Debug logging enabled")
	} else {
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		zapLogger, _ = cfg.Build()
		logger = zapLogger.Sugar().Named("n8n-cli")
	}
}

// Debug logs a debug message if debug mode is enabled
func Debug(format string, args ...interface{}) {
	if logger != nil {
		logger.Debugf(format, args...)
	}
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	if logger != nil {
		logger.Infof(format, args...)
	}
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	if logger != nil {
		logger.Warnf(format, args...)
	}
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	if logger != nil {
		logger.Errorf(format, args...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(format string, args ...interface{}) {
	if logger != nil {
		logger.Fatalf(format, args...)
	}
}
