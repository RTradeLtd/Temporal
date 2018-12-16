package queue

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/RTradeLtd/Temporal/log"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/jinzhu/gorm"
)

const (
	testCID           = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
	testRabbitAddress = "amqp://127.0.0.1:5672"
	testLogFilePath   = "../testenv/"
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
		{"PConfirmQ", args{PaymentConfirmationQueue, true}},
		{"DPCQ", args{DashPaymentConfirmationQueue, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := log.NewLogger("", true)
			if err != nil {
				t.Fatal(err)
			}
			qm, err := New(tt.args.queueName,
				testRabbitAddress, tt.args.publish, logger)
			if err != nil {
				t.Fatal(err)
			}
			if tt.name == "PConfirmQ" {
				// test a successful publish
				if err := qm.PublishMessage(PaymentConfirmation{
					UserName:      "testuser",
					PaymentNumber: 22,
				}); err != nil {
					t.Fatal(err)
				}
				// test a bad publish
				if err := qm.PublishMessage(map[string]interface{}{
					"foo": make(chan int),
				}); err == nil {
					t.Fatal("expected error")
				}
			} else if tt.name == "DPCQ" {
				// test a successful publish
				if err := qm.PublishMessage(DashPaymenConfirmation{
					UserName:         "testuser",
					PaymentForwardID: "paymentforwardi",
					PaymentNumber:    23,
				}); err != nil {
					t.Fatal(err)
				}
				// test a bad publish
				if err := qm.PublishMessage(map[string]interface{}{
					"foo": make(chan int),
				}); err == nil {
					t.Fatal("expected error")
				}
			}
		})
	}
}

// Covers the default case in `setupExchange`
func TestQueue_ExchangeFail(t *testing.T) {
	logger, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpfsPinQueue, testRabbitAddress, true, logger)
	if err != nil {
		t.Fatal(err)
	}
	if err = qmPublisher.setupExchange("bad-queue"); err == nil {
		t.Fatal("error expected")
	}
}

func TestQueue_RefundCredits(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	logger, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(DatabaseFileAddQueue, testRabbitAddress, false, logger)
	if err != nil {
		t.Fatal(err)
	}
	qmConsumer.db = db
	type args struct {
		username string
		callType string
		cost     float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"HasCost", args{"testuser", "ipfs-pin", 1}, false},
		{"NoCost", args{"testuser", "ipfs-pin", 0}, false},
		{"RefundFail", args{"userdoesnotexist", "ipfs-pin", 10}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := qmConsumer.refundCredits(tt.args.username, tt.args.callType, tt.args.cost); (err != nil) != tt.wantErr {
				t.Fatal(err)
			}
		})
	}
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_DatabaseFileAdd(t *testing.T) {
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	loggerPublisher, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(DatabaseFileAddQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(DatabaseFileAddQueue, testRabbitAddress, true, loggerPublisher)
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
	// test a bad publish
	if err := qmPublisher.PublishMessage(""); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPFSFile(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	loggerPublisher, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsFileQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpfsFileQueue, testRabbitAddress, true, loggerPublisher)
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
	// test a bad publish
	if err := qmPublisher.PublishMessage(""); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	// set temporary log dig
	cfg.LogDir = "./tmp/"
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPFSClusterPin(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	loggerPublisher, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsClusterPinQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpfsClusterPinQueue, testRabbitAddress, true, loggerPublisher)
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
	// test a bad publish
	if err := qmPublisher.PublishMessage(""); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_EmailSend(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	loggerPublisher, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(EmailSendQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(EmailSendQueue, testRabbitAddress, true, loggerPublisher)
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
	// test a bad publish
	if err := qmPublisher.PublishMessage(""); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPNSEntry(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	loggerPublisher, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpnsEntryQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpnsEntryQueue, testRabbitAddress, true, loggerPublisher)
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
	// test a bad publish
	if err := qmPublisher.PublishMessage(""); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPFSPin(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	loggerPublisher, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsPinQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	if qmConsumer.ExchangeName != PinExchange {
		t.Fatal("failed to properly set exchange name on consumer")
	}
	qmPublisher, err := New(IpfsPinQueue, testRabbitAddress, true, loggerPublisher)
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
	// test a bad publish
	if err := qmPublisher.PublishMessage(""); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	// set temporary log dig
	cfg.LogDir = "./tmp/"
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPFSKeyCreation(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	loggerPublisher, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsKeyCreationQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	if qmConsumer.ExchangeName != IpfsKeyExchange {
		t.Fatal("failed to properly set exchange name on consumer")
	}
	qmPublisher, err := New(IpfsKeyCreationQueue, testRabbitAddress, true, loggerPublisher)
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
	// test a bad publish
	if err := qmPublisher.PublishMessage(""); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	cancel()
	waitGroup.Wait()
}

func TestQueue_IPFSKeyCreation_Failure(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	loggerPublisher, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsKeyCreationQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	if qmConsumer.ExchangeName != IpfsKeyExchange {
		t.Fatal("failed to properly set exchange name on consumer")
	}
	qmPublisher, err := New(IpfsKeyCreationQueue, testRabbitAddress, true, loggerPublisher)
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
	cfg.Endpoints.Krab.TLS.CertPath = "/root"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, &sync.WaitGroup{}, db, cfg); err == nil {
		t.Fatal("expected error")
	}
}

func TestQueue_IPFSPin_Failure_RabbitMQ(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsPinQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	if qmConsumer.ExchangeName != PinExchange {
		t.Fatal("failed to properly set exchange name on consumer")
	}
	cfg.RabbitMQ.URL = "notarealurl"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, &sync.WaitGroup{}, db, cfg); err == nil {
		t.Fatal("expected error")
	}
}

func TestQueue_IPFSPin_Failure_LogFile(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsPinQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	if qmConsumer.ExchangeName != PinExchange {
		t.Fatal("failed to properly set exchange name on consumer")
	}
	cfg.LogDir = "/root/toor"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, &sync.WaitGroup{}, db, cfg); err == nil {
		t.Fatal("expected error")
	}
}

func TestQueue_IPFSFile_Failure_RTFS(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsFileQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	cfg.IPFS.APIConnection.Host = "notarealip"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, &sync.WaitGroup{}, db, cfg); err == nil {
		t.Fatal("expected error")
	}
}

func TestQueue_IPFSFile_Failure_RabbitMQ(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup our queue backend
	qmConsumer, err := New(IpfsFileQueue, testRabbitAddress, false, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	cfg.RabbitMQ.URL = "notarealip"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, &sync.WaitGroup{}, db, cfg); err == nil {
		t.Fatal("expected error")
	}
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
