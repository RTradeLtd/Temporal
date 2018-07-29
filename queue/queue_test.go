package queue_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/queue"
)

var defaultConfigFile = "/home/solidity/config.json"
var testCID = "QmPY5iMFjNZKxRbUZZC85wXb9CFgNSyzAy1LxwL62D8VGr"
var testEthAddress = "0x7E4A2359c745A982a54653128085eAC69E446DE1"

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

func TestQueues(t *testing.T) {
	cfg, err := config.LoadConfig(defaultConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	qm, err := queue.Initialize(queue.IpfsPinQueue, cfg.RabbitMQ.URL)
	if err != nil {
		t.Fatal(err)
	}

	pin := queue.IPFSPin{
		CID:              testCID,
		NetworkName:      "public",
		EthAddress:       testEthAddress,
		HoldTimeInMonths: 10,
	}

	err = qm.PublishMessageWithExchange(pin, queue.PinExchange)
	if err != nil {
		t.Fatal(err)
	}

}
