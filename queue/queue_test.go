package queue_test

import (
	"testing"
	"time"

	"github.com/RTradeLtd/Temporal/queue"
)

const (
	testCID           = "QmPY5iMFjNZKxRbUZZC85wXb9CFgNSyzAy1LxwL62D8VGr"
	testRabbitAddress = "amqp://127.0.0.1:5672"
	logFilePath       = "/tmp/%s_%v.log"
)

func TestQueue_Publish(t *testing.T) {
	type args struct {
		queueName string
		publish   bool
		service   bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"DFAQ", args{queue.DatabaseFileAddQueue, true, false}},
		{"IPQ", args{queue.IpfsPinQueue, true, false}},
		{"IFQ", args{queue.IpfsFileQueue, true, false}},
		{"IPCQ", args{queue.IpfsClusterPinQueue, true, false}},
		{"ESQ", args{queue.EmailSendQueue, true, false}},
		{"IEQ", args{queue.IpnsEntryQueue, true, false}},
		{"IKQ", args{queue.IpfsKeyCreationQueue, true, false}},
		{"PCreateQ", args{queue.PaymentCreationQueue, true, false}},
		{"PConfirmQ", args{queue.PaymentConfirmationQueue, true, false}},
		{"DPCQ", args{queue.DashPaymentConfirmationQueue, true, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qm, err := queue.Initialize(tt.args.queueName,
				testRabbitAddress, tt.args.publish, tt.args.service)
			if err != nil {
				t.Fatal(err)
			}
			switch tt.name {
			case "IPQ", "IKQ":
				if tt.name == "IPQ" {
					if err := qm.PublishMessageWithExchange(
						queue.IPFSPin{
							CID:              testCID,
							NetworkName:      "public",
							HoldTimeInMonths: 10,
						},
						queue.PinExchange,
					); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := qm.PublishMessageWithExchange(
						queue.IPFSKeyCreation{
							UserName:    "testuser",
							Name:        "mykey",
							Type:        "rsa",
							Size:        2048,
							NetworkName: "public",
							CreditCost:  0,
						},
						queue.IpfsKeyExchange,
					); err != nil {
						t.Fatal(err)
					}
				}
			default:
				if tt.name == "DFAQ" {
					if err := qm.PublishMessage(queue.DatabaseFileAdd{
						Hash:             testCID,
						HoldTimeInMonths: 10,
						UserName:         "testuser",
					}); err != nil {
						t.Fatal(err)
					}
				} else if tt.name == "IFQ" {
					if err := qm.PublishMessage(queue.IPFSFile{
						MinioHostIP:      "127.0.0.1:9090",
						FileName:         "testfile",
						FileSize:         100,
						BucketName:       "filesuploadbucket",
						ObjectName:       "myobject",
						UserName:         "testuser",
						NetworkName:      "public",
						HoldTimeInMonths: "10",
						CreditCost:       10,
						Encrypted:        false,
					}); err != nil {
						t.Fatal(err)
					}
				} else if tt.name == "IPCQ" {
					if err := qm.PublishMessage(queue.IPFSClusterPin{
						CID:              testCID,
						NetworkName:      "public",
						UserName:         "testuser",
						HoldTimeInMonths: 10,
						CreditCost:       10,
					}); err != nil {
						t.Fatal(err)
					}
				} else if tt.name == "ESQ" {
					if err := qm.PublishMessage(queue.EmailSend{
						Subject:     "test email",
						Content:     "this is a test email",
						ContentType: "text/html",
						UserNames:   []string{"testuser"},
						Emails:      []string{"testuser@example.com"},
					}); err != nil {
						t.Fatal(err)
					}
				} else if tt.name == "IEQ" {
					if err := qm.PublishMessage(queue.IPNSEntry{
						CID:         testCID,
						LifeTime:    time.Minute,
						TTL:         time.Second,
						Resolve:     true,
						Key:         "testkey",
						UserName:    "testuser",
						NetworkName: "public",
						CreditCost:  10,
					}); err != nil {
						t.Fatal(err)
					}
				} else if tt.name == "PCreateQ" {
					if err := qm.PublishMessage(queue.PaymentCreation{
						TxHash:     "testuser-22",
						Blockchain: "eth",
						UserName:   "testuser",
					}); err != nil {
						t.Fatal(err)
					}
				} else if tt.name == "PConfirmQ" {
					if err := qm.PublishMessage(queue.PaymentConfirmation{
						UserName:      "testuser",
						PaymentNumber: 22,
					}); err != nil {
						t.Fatal(err)
					}
				} else if tt.name == "DPCQ" {
					if err := qm.PublishMessage(queue.DashPaymenConfirmation{
						UserName:         "testuser",
						PaymentForwardID: "paymentforwardi",
						PaymentNumber:    23,
					}); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
}
