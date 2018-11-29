package queue_test

import (
	"testing"
	"time"

	"github.com/RTradeLtd/Temporal/queue"
)

const (
	testCID           = "QmPY5iMFjNZKxRbUZZC85wXb9CFgNSyzAy1LxwL62D8VGr"
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
		{"IPCQ", args{queue.IpfsClusterPinQueue, false, false}},
		{"ESQ", args{queue.EmailSendQueue, false, false}},
		{"IEQ", args{queue.IpnsEntryQueue, false, false}},
		{"IKQ", args{queue.IpfsKeyCreationQueue, false, false}},
		{"PCreateQ", args{queue.PaymentCreationQueue, false, false}},
		{"PConfirmQ", args{queue.PaymentConfirmationQueue, false, false}},
		{"DPCQ", args{queue.DashPaymentConfirmationQueue, false, false}},
		{"MUQ", args{queue.MongoUpdateQueue, false, false}},
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
					if err := qm.PublishMessageWithExchange(queue.IPFSPin{
						CID:              testCID,
						NetworkName:      "public",
						HoldTimeInMonths: 10,
					}); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := qm.PublishMessageWithExchange(queue.IPFSKeyCreation{
						UserName:    "testuser",
						Name:        "mykey",
						Type:        "rsa",
						Size:        2048,
						NetworkName: "public",
						CreditCost:  0,
					}); err != nil {
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
						TxHash:        "0xf567fc31bc7bfa9cfd37ec98dd4e2bfb79f20f71fe6862dd9b941f65e4ee28ad",
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
				} else if tt.name == "MUQ" {
					if err := qm.PublishMessage(queue.MongoUpdate{
						DatabaseName:   "myappdb",
						CollectionName: "uploads",
						Fields: map[string]string{
							"foo": "bar",
						},
					}); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
}
