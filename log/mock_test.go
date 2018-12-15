package log

import "testing"

func TestNewTestLogger(t *testing.T) {
	logger, out := NewTestLogger()
	logger.Info("hi")
	if out.All()[0].Message != "hi" {
		t.Error("bad logger")
	}
}
