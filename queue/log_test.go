package queue

import (
	"testing"

	"github.com/RTradeLtd/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLog_Publisher(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	observer, out := observer.New(zap.InfoLevel)
	logger := zap.New(observer).Sugar()
	_, err = New(IpfsClusterPinQueue, cfg.RabbitMQ.URL, true, cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
	if out.All()[out.Len()-1].Message != "channel opened" {
		t.Fatal("failed to recover correct message")
	}
}

func TestLog_Consumer(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	observer, out := observer.New(zap.InfoLevel)
	logger := zap.New(observer).Sugar()
	qm, err := New(IpfsClusterPinQueue, cfg.RabbitMQ.URL, false, cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
	if out.All()[out.Len()-1].Message != "queue declared" {
		t.Fatal("failed to recover correct message")
	}
	if err = qm.Close(); err != nil {
		t.Fatal(err)
	}
}
