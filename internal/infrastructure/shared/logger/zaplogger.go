package logger

import (
	"os"
	"sync"
	"task-processor/internal/infrastructure/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
    *zap.Logger
}

// Ensure ZapLogger implements Logger interface
var _ Logger = (*ZapLogger)(nil)

var (
    instance Logger
    once     sync.Once
)
    
// GetLogger returns a singleton instance of Logger
func GetLogger() Logger {
    once.Do(func() {
        cfg := config.GetConfig()

        // Map config log level to zapcore.Level
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
        default:
            level = zap.InfoLevel
        }

        // Choose encoder based on environment
        var encoder zapcore.Encoder
        switch cfg.App.Env {
        case "dev", "test":
            encCfg := zap.NewDevelopmentEncoderConfig()
            encCfg.TimeKey = "T"
            encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
            encoder = zapcore.NewConsoleEncoder(encCfg)
        default: // staging/prod
            encCfg := zap.NewProductionEncoderConfig()
            encCfg.TimeKey = "ts"
            encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
            encCfg.LevelKey = "level"
            encCfg.MessageKey = "msg"
            encCfg.CallerKey = "caller"
            encoder = zapcore.NewJSONEncoder(encCfg)
        }

        // Create zap core
        core := zapcore.NewCore(
            encoder,
            zapcore.Lock(os.Stderr), // logs go to stderr
            level,
        )

        // Add caller info
        opts := []zap.Option{zap.AddCaller(), zap.AddCallerSkip(1)}
        if cfg.App.Env == "dev" || cfg.App.Env == "test" {
            opts = append(opts, zap.Development()) // enable development mode for dev/test
        }

		instance = &ZapLogger{zap.New(core, opts...)}
    })

    return instance
}

// Implementation of Logger interface methods for ZapLogger

func (l *ZapLogger) Debug(msg string, fields ...Field) {
    l.Logger.Debug(msg, fields...)
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
    l.Logger.Info(msg, fields...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
    l.Logger.Warn(msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
    l.Logger.Error(msg, fields...)
}

func (l *ZapLogger) Fatal(msg string, fields ...Field) {
    l.Logger.Fatal(msg, fields...)
}

func (l *ZapLogger) Panic(msg string, fields ...Field) {
    l.Logger.Panic(msg, fields...)
}

// With creates a new logger with additional fields
func (l *ZapLogger) With(fields ...Field) Logger {
    return &ZapLogger{l.Logger.With(fields...)}
}

// Named adds a name to the logger
func (l *ZapLogger) Named(name string) Logger {
    return &ZapLogger{l.Logger.Named(name)}
}

// Sync flushes any buffered log entries
func (l *ZapLogger) Sync() error {
    return l.Logger.Sync()
}
