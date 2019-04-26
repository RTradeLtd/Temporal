package log

import (
	"testing"

	"github.com/bobheadxi/zapx/ztest"
)

func TestNewProcessLogger(t *testing.T) {
	l, out := ztest.NewObservable()
	logger := NewProcessLogger(l.Sugar(), "network_up", "id", "1234")
	logger.Info("hi")
	if out.All()[0].ContextMap()["network_up.id"].(string) != "1234" {
		t.Error("bad logger")
	}
}
