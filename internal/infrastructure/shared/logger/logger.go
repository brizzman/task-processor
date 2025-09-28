package logger

import (
	"go.uber.org/zap"
)


// Logger is an interface for structured logging
type Logger interface {
    // Debug logs a message at Debug level
    Debug(msg string, fields ...Field)
    // Info logs a message at Info level
    Info(msg string, fields ...Field)
    // Warn logs a message at Warn level
    Warn(msg string, fields ...Field)
    // Error logs a message at Error level
    Error(msg string, fields ...Field)
    // Fatal logs a message at Fatal level and terminates the program
    Fatal(msg string, fields ...Field)
    // Panic logs a message at Panic level and calls panic
    Panic(msg string, fields ...Field)

    // With creates a new logger with additional fields
    With(fields ...Field) Logger
    // Named adds a name to the logger
    Named(name string) Logger

    // Sync flushes any buffered log entries (used at program exit)
    Sync() error
}

// Field represents a field for structured logging
type Field = zap.Field

// Constants for creating fields (re-exported from zap)
var (
    String   = zap.String
    Int      = zap.Int
    Int32    = zap.Int32
    Int64    = zap.Int64
    Float32  = zap.Float32
    Float64  = zap.Float64
    Bool     = zap.Bool
    Any      = zap.Any
    Error    = zap.Error
    Time     = zap.Time
    Duration = zap.Duration
)