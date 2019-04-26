package log

import (
	"go.uber.org/zap"
)

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
