package log

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a default "sugared" logger based on dev toggle
func NewLogger(logpath string, dev bool) (sugar *zap.SugaredLogger, err error) {
	var logger *zap.Logger
	var config zap.Config
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

	// set log paths
	if logpath != "" {
		if err = os.MkdirAll(filepath.Dir(logpath), os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create directories for logpath '%s': %s",
				logpath, err.Error())
		}
		config.OutputPaths = append(config.OutputPaths, logpath)
	}

	if logger, err = config.Build(); err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

// NewProcessLogger creates a new logger that sets prefixes on fields for
// logging a specific process
func NewProcessLogger(l *zap.SugaredLogger, process string, fields ...interface{}) *zap.SugaredLogger {
	args := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i += 2 {
		args[i] = process + "." + fields[i].(string)
		args[i+1] = fields[i+1]
	}
	return l.With(args...)
}
