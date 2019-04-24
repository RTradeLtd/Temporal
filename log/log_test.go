package log

import (
	"testing"

	zapx "github.com/bobheadxi/zapx/test"
)

func TestNewProcessLogger(t *testing.T) {
	l, out := zapx.NewObservable()
	logger := NewProcessLogger(l.Sugar(), "network_up", "id", "1234")
	logger.Info("hi")
	if out.All()[0].ContextMap()["network_up.id"].(string) != "1234" {
		t.Error("bad logger")
	}
}
