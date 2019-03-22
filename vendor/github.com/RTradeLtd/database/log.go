package database

import "go.uber.org/zap"

// LogLevel indicates different logging levels
type LogLevel string

const (
	// LogLevelInfo denotes info-level logging
	LogLevelInfo LogLevel = "info"

	// LogLevelDebug denotes debug-level logging
	LogLevelDebug LogLevel = "debug"
)

// Logger defines the database's logging interface
type Logger interface{ Print(...interface{}) }

// PrintLogger wraps a single logFunc to implement the Logger interface
type PrintLogger struct{ logFunc func(...interface{}) }

// NewZapLogger wraps a zap logger in a PrintLogger
func NewZapLogger(level LogLevel, l *zap.SugaredLogger) *PrintLogger {
	var logFunc = func(...interface{}) {}
	if l != nil {
		switch level {
		case LogLevelInfo:
			logFunc = l.Info
		case LogLevelDebug:
			logFunc = l.Debug
		}
	}
	return &PrintLogger{logFunc}
}

// Print logs the given message
func (p *PrintLogger) Print(args ...interface{}) { p.logFunc(args...) }
