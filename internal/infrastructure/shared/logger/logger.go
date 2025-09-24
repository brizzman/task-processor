package logger

import (
	"os"
	"task-processor/internal/infrastructure/config"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

var (
	logger Logger
	once   sync.Once
)

func GetLogger() *Logger {
	once.Do(func() {
		cfg := config.GetConfig()

		// Set log level from config
		var level zapcore.Level
		switch cfg.Log.Level {
		case "debug":
			level = zap.DebugLevel
		case "info":
			level = zap.InfoLevel
		case "warn", "warning":
			level = zap.WarnLevel
		case "err", "error":
			level = zap.ErrorLevel
		case "fatal":
			level = zap.FatalLevel
		case "panic":
			level = zap.PanicLevel
		default:
			level = zap.InfoLevel
		}

		// Configure JSON encoder (better for log aggregation systems like Loki/ELK)
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "ts"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.LevelKey = "level"
		encoderCfg.MessageKey = "msg"
		encoderCfg.CallerKey = "caller"

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stderr),
			level,
		)

		zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

		logger = Logger{zapLogger}
	})

	return &logger
}
