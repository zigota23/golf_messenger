package logger

import (
	"fmt"

	"github.com/yourusername/golf_messenger/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Initialize(cfg *config.Config) error {
	var zapConfig zap.Config

	if cfg.Logging.Encoding == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	level, err := parseLogLevel(cfg.Logging.Level)
	if err != nil {
		return err
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	if len(cfg.Logging.OutputPaths) > 0 {
		zapConfig.OutputPaths = cfg.Logging.OutputPaths
	}
	if len(cfg.Logging.ErrorOutputPaths) > 0 {
		zapConfig.ErrorOutputPaths = cfg.Logging.ErrorOutputPaths
	}

	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zapConfig.Build()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	Log = logger
	return nil
}

func parseLogLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}

func Sync() error {
	if Log != nil {
		return Log.Sync()
	}
	return nil
}

func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Fatal(msg, fields...)
	}
}

func Panic(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Panic(msg, fields...)
	}
}
