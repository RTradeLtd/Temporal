package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a default zap logger
func NewLogger(logPath string, dev bool) (*zap.SugaredLogger, error) {
	// setup logging variables
	var (
		logger *zap.Logger
		config zap.Config
		err    error
	)

	// setup logging configuration
	if dev {
		// Log:         DebugLevel
		// Encoder:     console
		// Errors:      stderr
		// Sampling:    no
		// Stacktraces: WarningLevel
		// Colors:      capitals
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		// Log:         InfoLevel
		// Encoder:     json
		// Errors:      stderr
		// Sampling:    yes
		// Stacktraces: ErrorLevel
		config = zap.NewProductionConfig()
	}

	// set the destination directory for logs
	config.OutputPaths = append(config.OutputPaths, logPath)

	if logger, err = config.Build(); err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
