package queue

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/RTradeLtd/config"
)

const (
	testCID           = "QmPY5iMFjNZKxRbUZZC85wXb9CFgNSyzAy1LxwL62D8VGr"
	testRabbitAddress = "amqp://127.0.0.1:5672"
	logFilePath       = "/tmp/%s_%v.log"
	testCfgPath       = "../testenv/config.json"
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
		{"DFAQ", args{DatabaseFileAddQueue, true, false}},
		{"IPQ", args{IpfsPinQueue, true, false}},
		{"IFQ", args{IpfsFileQueue, true, false}},
		{"IPCQ", args{IpfsClusterPinQueue, true, false}},
		{"ESQ", args{EmailSendQueue, true, false}},
		{"IEQ", args{IpnsEntryQueue, true, false}},
		{"IKQ", args{IpfsKeyCreationQueue, true, false}},
		{"PCreateQ", args{PaymentCreationQueue, true, false}},
		{"PConfirmQ", args{PaymentConfirmationQueue, true, false}},
		{"DPCQ", args{DashPaymentConfirmationQueue, true, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qm, err := Initialize(tt.args.queueName,
				testRabbitAddress, tt.args.publish, tt.args.service)
			if err != nil {
				t.Fatal(err)
			}
			switch tt.name {
			case "IPQ", "IKQ":
				if tt.name == "IPQ" {
					if err := qm.PublishMessageWithExchange(
						IPFSPin{
							CID:              testCID,
							NetworkName:      "public",
							HoldTimeInMonths: 10,
						},
						PinExchange,
					); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := qm.PublishMessageWithExchange(
						IPFSKeyCreation{
							UserName:    "testuser",
							Name:        "mykey",
							Type:        "rsa",
							Size:        2048,
							NetworkName: "public",
							CreditCost:  0,
						},
						IpfsKeyExchange,
					); err != nil {
						t.Fatal(err)
					}
				}
			default:
				if tt.name == "IFQ" {
					if err := qm.PublishMessage(IPFSFile{
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
					if err := qm.PublishMessage(IPFSClusterPin{
						CID:              testCID,
						NetworkName:      "public",
						UserName:         "testuser",
						HoldTimeInMonths: 10,
						CreditCost:       10,
					}); err != nil {
						t.Fatal(err)
					}
				} else if tt.name == "ESQ" {
					if err := qm.PublishMessage(EmailSend{
						Subject:     "test email",
						Content:     "this is a test email",
						ContentType: "text/html",
						UserNames:   []string{"testuser"},
						Emails:      []string{"testuser@example.com"},
					}); err != nil {
						t.Fatal(err)
					}
				} else if tt.name == "IEQ" {
					if err := qm.PublishMessage(IPNSEntry{
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
					if err := qm.PublishMessage(PaymentCreation{
						TxHash:     "testuser-22",
						Blockchain: "eth",
						UserName:   "testuser",
					}); err != nil {
						t.Fatal(err)
					}
				} else if tt.name == "PConfirmQ" {
					if err := qm.PublishMessage(PaymentConfirmation{
						UserName:      "testuser",
						PaymentNumber: 22,
					}); err != nil {
						t.Fatal(err)
					}
				} else if tt.name == "DPCQ" {
					if err := qm.PublishMessage(DashPaymenConfirmation{
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

func TestQueue_DatabaseFileAdd(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := Initialize(DatabaseFileAddQueue, testRabbitAddress, false, true)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := Initialize(DatabaseFileAddQueue, testRabbitAddress, true, false)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err := qmPublisher.PublishMessage(DatabaseFileAdd{
		Hash:             testCID,
		HoldTimeInMonths: 10,
		UserName:         "testuser",
	}); err != nil {
		t.Fatal(err)
	}
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", "", "", "", cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}
