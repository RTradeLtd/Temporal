package queue_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/queue"
)

const (
	defaultConfigFile = "/home/solidity/config.json"
	testCID           = "QmPY5iMFjNZKxRbUZZC85wXb9CFgNSyzAy1LxwL62D8VGr"
	testEthAddress    = "0x7E4A2359c745A982a54653128085eAC69E446DE1"
	testRabbitAddress = "amqp://127.0.0.1:5672"
)

func TestInitialize(t *testing.T) {
	type args struct {
		queueName string
		publish   bool
		service   bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"DFAQ", args{queue.DatabaseFileAddQueue, false, false}},
		{"IPQ", args{queue.IpfsPinQueue, false, false}},
		{"IFQ", args{queue.IpfsFileQueue, false, false}},
		{"PPCQ", args{queue.PinPaymentConfirmationQueue, false, false}},
		{"PPSQ", args{queue.PinPaymentSubmissionQueue, false, false}},
		{"ESQ", args{queue.EmailSendQueue, false, false}},
		{"IEQ", args{queue.IpnsEntryQueue, false, false}},
		{"IPRQ", args{queue.IpfsPinRemovalQueue, false, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := queue.Initialize(tt.args.queueName,
				testRabbitAddress, tt.args.publish, tt.args.service); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestQueues(t *testing.T) {
	qm, err := queue.Initialize(queue.IpfsPinQueue, testRabbitAddress, false, false)
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
