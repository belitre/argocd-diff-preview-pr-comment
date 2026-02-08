package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log is the global logger instance
	Log *zap.SugaredLogger
)

// LogLevel represents the available log levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// ValidLogLevels returns all valid log levels
func ValidLogLevels() []string {
	return []string{
		string(DebugLevel),
		string(InfoLevel),
		string(WarnLevel),
		string(ErrorLevel),
		string(FatalLevel),
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToLower(level) {
	case string(DebugLevel):
		return DebugLevel, nil
	case string(InfoLevel):
		return InfoLevel, nil
	case string(WarnLevel):
		return WarnLevel, nil
	case string(ErrorLevel):
		return ErrorLevel, nil
	case string(FatalLevel):
		return FatalLevel, nil
	default:
		return InfoLevel, fmt.Errorf("invalid log level: %s (valid levels: %s)",
			level, strings.Join(ValidLogLevels(), ", "))
	}
}

// toZapLevel converts LogLevel to zapcore.Level
func toZapLevel(level LogLevel) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Initialize initializes the global logger with the specified log level
func Initialize(level LogLevel) error {
	zapLevel := toZapLevel(level)

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.Encoding = "console"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	Log = logger.Sugar()
	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *zap.SugaredLogger {
	if Log == nil {
		// Initialize with default level if not initialized
		if err := Initialize(InfoLevel); err != nil {
			panic(fmt.Sprintf("failed to initialize default logger: %v", err))
		}
	}
	return Log
}

// Sync flushes any buffered log entries
func Sync() error {
	if Log != nil {
		return Log.Sync()
	}
	return nil
}
