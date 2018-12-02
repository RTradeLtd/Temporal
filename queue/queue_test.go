package queue

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/RTradeLtd/config"
)

const (
	testCID           = "QmPY5iMFjNZKxRbUZZC85wXb9CFgNSyzAy1LxwL62D8VGr"
	testRabbitAddress = "amqp://127.0.0.1:5672"
	testLogFilePath   = "../templogs"
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
			if tt.name == "PCreateQ" {
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
		})
	}
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_DatabaseFileAdd(t *testing.T) {
	defer func() {
		os.Remove(fmt.Sprintf("%s-%s_service.log", testLogFilePath, DatabaseFileAddQueue))
	}()
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := Initialize(DatabaseFileAddQueue, testRabbitAddress, false, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := Initialize(DatabaseFileAddQueue, testRabbitAddress, true, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	if err := qmPublisher.PublishMessage(DatabaseFileAdd{
		Hash:             testCID,
		HoldTimeInMonths: 10,
		UserName:         "testuser",
	}); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", true, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPFSFile(t *testing.T) {
	defer func() {
		os.Remove(fmt.Sprintf("%s-%s_service.log", testLogFilePath, IpfsFileQueue))
	}()
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := Initialize(IpfsFileQueue, testRabbitAddress, false, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := Initialize(IpfsFileQueue, testRabbitAddress, true, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer qmPublisher.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	if err := qmPublisher.PublishMessage(IPFSFile{
		MinioHostIP:      "127.0.0.1:9090",
		FileName:         "testfile",
		FileSize:         100,
		BucketName:       "filesuploadbucket",
		ObjectName:       "myobject",
		UserName:         "testuser",
		NetworkName:      "public",
		HoldTimeInMonths: "10",
		CreditCost:       10,
	}); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", true, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPFSClusterPin(t *testing.T) {
	defer func() {
		os.Remove(fmt.Sprintf("%s-%s_service.log", testLogFilePath, IpfsClusterPinQueue))
	}()
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := Initialize(IpfsClusterPinQueue, testRabbitAddress, false, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := Initialize(IpfsClusterPinQueue, testRabbitAddress, true, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer qmPublisher.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	if err := qmPublisher.PublishMessage(IPFSClusterPin{
		CID:              testCID,
		NetworkName:      "public",
		UserName:         "testuser",
		HoldTimeInMonths: 10,
		CreditCost:       10,
	}); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", true, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_EmailSend(t *testing.T) {
	defer func() {
		os.Remove(fmt.Sprintf("%s-%s_service.log", testLogFilePath, EmailSendQueue))
	}()
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := Initialize(EmailSendQueue, testRabbitAddress, false, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := Initialize(EmailSendQueue, testRabbitAddress, true, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer qmPublisher.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	if err := qmPublisher.PublishMessage(EmailSend{
		Subject:     "test email",
		Content:     "this is a test email",
		ContentType: "text/html",
		UserNames:   []string{"testuser"},
		Emails:      []string{"testuser@example.com"},
	}); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", true, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPNSEntry(t *testing.T) {
	defer func() {
		os.Remove(fmt.Sprintf("%s-%s_service.log", testLogFilePath, IpnsEntryQueue))
	}()
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := Initialize(IpnsEntryQueue, testRabbitAddress, false, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := Initialize(IpnsEntryQueue, testRabbitAddress, true, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer qmPublisher.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	if err := qmPublisher.PublishMessage(IPNSEntry{
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
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", true, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPFSPin(t *testing.T) {
	defer func() {
		os.Remove(fmt.Sprintf("%s-%s_service.log", testLogFilePath, IpfsPinQueue))
	}()
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := Initialize(IpfsPinQueue, testRabbitAddress, false, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if qmConsumer.ExchangeName != PinExchange {
		t.Fatal("failed to properly set exchange name on consumer")
	}
	qmPublisher, err := Initialize(IpfsPinQueue, testRabbitAddress, true, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if qmPublisher.ExchangeName != PinExchange {
		t.Fatal("failed to properly set exchange name on publisher")
	}
	defer qmPublisher.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	if err := qmPublisher.PublishMessageWithExchange(IPFSPin{
		CID:              testCID,
		NetworkName:      "public",
		HoldTimeInMonths: 10},
		PinExchange,
	); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", true, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPFSKeyCreation(t *testing.T) {
	defer func() {
		os.Remove(fmt.Sprintf("%s-%s_service.log", testLogFilePath, IpfsKeyCreationQueue))
	}()
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := Initialize(IpfsKeyCreationQueue, testRabbitAddress, false, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if qmConsumer.ExchangeName != IpfsKeyExchange {
		t.Fatal("failed to properly set exchange name on consumer")
	}
	qmPublisher, err := Initialize(IpfsKeyCreationQueue, testRabbitAddress, true, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if qmPublisher.ExchangeName != IpfsKeyExchange {
		t.Fatal("failed to properly set exchange name on publisher")
	}
	defer qmPublisher.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	if err := qmPublisher.PublishMessageWithExchange(IPFSKeyCreation{
		UserName:    "testuser",
		Name:        "mykey",
		Type:        "rsa",
		Size:        2048,
		NetworkName: "public",
		CreditCost:  0}, IpfsKeyExchange,
	); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", true, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}
