package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// NewTestLogger bootstraps a test logger that allows interrogation of output
func NewTestLogger() (sugar *zap.SugaredLogger, out *observer.ObservedLogs) {
	observer, out := observer.New(zap.InfoLevel)
	return zap.New(observer).Sugar(), out
}
