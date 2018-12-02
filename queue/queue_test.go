package queue

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/jinzhu/gorm"
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
	}
	tests := []struct {
		name string
		args args
	}{
		{"PCreateQ", args{PaymentCreationQueue, true}},
		{"PConfirmQ", args{PaymentConfirmationQueue, true}},
		{"DPCQ", args{DashPaymentConfirmationQueue, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qm, err := New(tt.args.queueName,
				testRabbitAddress, tt.args.publish)
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

// Covers the default case in `setupExchange`
func TestQueue_ExchangeFail(t *testing.T) {
	qmPublisher, err := New(IpfsPinQueue, testRabbitAddress, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if err = qmPublisher.setupExchange("bad-queue"); err == nil {
		t.Fatal("error expected")
	}
}

func TestQueue_RefundCredits(t *testing.T) {
	defer func() {
		os.Remove(fmt.Sprintf("%s-%s_service.log", testLogFilePath, DatabaseFileAddQueue))
	}()
	// setup our queue backend
	qmConsumer, err := New(DatabaseFileAddQueue, testRabbitAddress, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		username string
		callType string
		cost     float64
	}
	tests := []struct {
		name string
		args args
	}{
		{"HasCost", args{"testuser", "ipfs-pin", 1}},
		{"NoCost", args{"testuser", "ipfs-pin", 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := qmConsumer.refundCredits(tt.args.username, tt.args.callType, tt.args.cost); err != nil {
				t.Fatal(err)
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
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(DatabaseFileAddQueue, testRabbitAddress, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(DatabaseFileAddQueue, testRabbitAddress, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
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
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", db, cfg); err != nil {
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
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsFileQueue, testRabbitAddress, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpfsFileQueue, testRabbitAddress, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
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
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", db, cfg); err != nil {
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
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsClusterPinQueue, testRabbitAddress, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpfsClusterPinQueue, testRabbitAddress, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
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
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", db, cfg); err != nil {
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
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(EmailSendQueue, testRabbitAddress, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(EmailSendQueue, testRabbitAddress, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
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
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", db, cfg); err != nil {
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
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpnsEntryQueue, testRabbitAddress, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpnsEntryQueue, testRabbitAddress, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
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
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", db, cfg); err != nil {
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
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsPinQueue, testRabbitAddress, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if qmConsumer.ExchangeName != PinExchange {
		t.Fatal("failed to properly set exchange name on consumer")
	}
	qmPublisher, err := New(IpfsPinQueue, testRabbitAddress, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if qmPublisher.ExchangeName != PinExchange {
		t.Fatal("failed to properly set exchange name on publisher")
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
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
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", db, cfg); err != nil {
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
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsKeyCreationQueue, testRabbitAddress, false, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if qmConsumer.ExchangeName != IpfsKeyExchange {
		t.Fatal("failed to properly set exchange name on consumer")
	}
	qmPublisher, err := New(IpfsKeyCreationQueue, testRabbitAddress, true, testLogFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if qmPublisher.ExchangeName != IpfsKeyExchange {
		t.Fatal("failed to properly set exchange name on publisher")
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
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
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, "", db, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

func loadDatabase(cfg *config.TemporalConfig) (*gorm.DB, error) {
	return database.OpenDBConnection(database.DBOptions{
		User:           cfg.Database.Username,
		Password:       cfg.Database.Password,
		Address:        cfg.Database.URL,
		Port:           cfg.Database.Port,
		SSLModeDisable: true,
	})
}
