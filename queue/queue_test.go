package queue_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/queue"
)

const (
	defaultConfigFile = "/home/solidity/config.json"
	testCID           = "QmPY5iMFjNZKxRbUZZC85wXb9CFgNSyzAy1LxwL62D8VGr"
	testEthAddress    = "0x7E4A2359c745A982a54653128085eAC69E446DE1"
)

func TestInitialize(t *testing.T) {
	cfg, err := config.LoadConfig(defaultConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		queueName     string
		connectionURL string
		publish       bool
		service       bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"DFAQ", args{queue.DatabaseFileAddQueue, cfg.RabbitMQ.URL, false, false}},
		{"IPQ", args{queue.IpfsPinQueue, cfg.RabbitMQ.URL, false, false}},
		{"IFQ", args{queue.IpfsFileQueue, cfg.RabbitMQ.URL, false, false}},
		{"PPCQ", args{queue.PinPaymentConfirmationQueue, cfg.RabbitMQ.URL, false, false}},
		{"PPSQ", args{queue.PinPaymentSubmissionQueue, cfg.RabbitMQ.URL, false, false}},
		{"ESQ", args{queue.EmailSendQueue, cfg.RabbitMQ.URL, false, false}},
		{"IEQ", args{queue.IpnsEntryQueue, cfg.RabbitMQ.URL, false, false}},
		{"IPRQ", args{queue.IpfsPinRemovalQueue, cfg.RabbitMQ.URL, false, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := queue.Initialize(tt.args.queueName, tt.args.connectionURL, tt.args.publish, tt.args.service)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestQueues(t *testing.T) {
	cfg, err := config.LoadConfig(defaultConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	qm, err := queue.Initialize(queue.IpfsPinQueue, cfg.RabbitMQ.URL, false, false)
	if err != nil {
		t.Fatal(err)
	}

	pin := queue.IPFSPin{
		CID:              testCID,
		NetworkName:      "public",
		HoldTimeInMonths: 10,
	}

	err = qm.PublishMessageWithExchange(pin, queue.PinExchange)
	if err != nil {
		t.Fatal(err)
	}
}
