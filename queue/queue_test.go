package queue_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/queue"
)

var defaultConfigFile = "/home/solidity/config.json"

func TestQueueInitialization(t *testing.T) {
	cfg, err := config.LoadConfig(defaultConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	_, err = queue.Initialize(queue.DatabaseFileAddQueue, cfg.RabbitMQ.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = queue.Initialize(queue.DatabasePinAddQueue, cfg.RabbitMQ.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = queue.Initialize(queue.IpnsUpdateQueue, cfg.RabbitMQ.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = queue.Initialize(queue.IpfsPinQueue, cfg.RabbitMQ.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = queue.Initialize(queue.IpfsFileQueue, cfg.RabbitMQ.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = queue.Initialize(queue.PinPaymentConfirmationQueue, cfg.RabbitMQ.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = queue.Initialize(queue.PinPaymentSubmissionQueue, cfg.RabbitMQ.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = queue.Initialize(queue.EmailSendQueue, cfg.RabbitMQ.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = queue.Initialize(queue.IpnsEntryQueue, cfg.RabbitMQ.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = queue.Initialize(queue.IpfsPinRemovalQueue, cfg.RabbitMQ.URL)
	if err != nil {
		t.Fatal(err)
	}
}
