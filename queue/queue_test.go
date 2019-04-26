package queue

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/streadway/amqp"

	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2"
	"github.com/RTradeLtd/gorm"
)

const (
	testCID           = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
	testCID2          = "QmQ5vhrL7uv6tuoN9KeVBwd4PwfQkXdVVmDLUZuTNxqgvm"
	testRabbitAddress = "amqp://127.0.0.1:5672"
	testLogFilePath   = "../testenv/"
	testCfgPath       = "../testenv/config.json"
)

func TestQueue_Publish(t *testing.T) {
	dev = true
	type args struct {
		queueName Queue
		publish   bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"PConfirmQ", args{EthPaymentConfirmationQueue, true}},
		{"DPCQ", args{DashPaymentConfirmationQueue, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			cfg, err := config.LoadConfig(testCfgPath)
			if err != nil {
				t.Fatal(err)
			}
			qm, err := New(tt.args.queueName,
				testRabbitAddress, tt.args.publish, dev, cfg, logger)
			if err != nil {
				t.Fatal(err)
			}
			if tt.name == "PConfirmQ" {
				// test a successful publish
				if err := qm.PublishMessage(EthPaymentConfirmation{
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

func TestRegisterChannelClosure(t *testing.T) {
	dev = true
	logger := zaptest.NewLogger(t).Sugar()
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpfsPinQueue, testRabbitAddress, true, dev, cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
	// declare the channel to receive messages on
	qmPublisher.RegisterConnectionClosure()
	go func() {
		qmPublisher.ErrCh <- &amqp.Error{Code: 400, Reason: "test", Server: true, Recover: false}
	}()
	msg := <-qmPublisher.ErrCh
	if msg.Code != 400 {
		t.Fatal("bad code received")
	}
	if msg.Reason != "test" {
		t.Fatal("bad reason")
	}
	if !msg.Server {
		t.Fatal("bad server")
	}
	if msg.Recover {
		t.Fatal("bad recover")
	}
	if err := qmPublisher.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestQueue_RefundCredits(t *testing.T) {
	dev = true
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	logger := zaptest.NewLogger(t).Sugar()
	// setup our queue backend
	qmConsumer, err := New(IpfsClusterPinQueue, testRabbitAddress, false, dev, cfg, logger)
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

func TestQueue_ConnectionClosure(t *testing.T) {
	dev = true
	logger := zaptest.NewLogger(t).Sugar()
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		queueName Queue
	}
	tests := []struct {
		name string
		args args
	}{
		{IpfsClusterPinQueue.String(), args{IpfsClusterPinQueue}},
		{EmailSendQueue.String(), args{EmailSendQueue}},
		{IpnsEntryQueue.String(), args{IpnsEntryQueue}},
		{IpfsPinQueue.String(), args{IpfsPinQueue}},
		{IpfsKeyCreationQueue.String(), args{IpfsKeyCreationQueue}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qmPublisher, err := New(tt.args.queueName, testRabbitAddress, false, dev, cfg, logger)
			if err != nil {
				t.Fatal(err)
			}
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				if err := qmPublisher.ConsumeMessages(context.Background(), wg, db, cfg); err != nil && err.Error() != ErrReconnect {
					t.Fatal(err)
				}
			}()
			go func() {
				qmPublisher.ErrCh <- &amqp.Error{Code: 400, Reason: "error", Server: true, Recover: false}
			}()
			wg.Wait()
		})
	}
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPFSClusterPin(t *testing.T) {
	dev = true
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer := zaptest.NewLogger(t).Sugar()
	loggerPublisher := zaptest.NewLogger(t).Sugar()
	// setup our queue backend
	qmConsumer, err := New(IpfsClusterPinQueue, testRabbitAddress, false, dev, cfg, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpfsClusterPinQueue, testRabbitAddress, true, dev, cfg, loggerPublisher)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
	// test an already seen update (this comes from the previous database file add)
	if err := qmPublisher.PublishMessage(IPFSClusterPin{
		CID:              testCID,
		NetworkName:      "public",
		UserName:         "testuser",
		HoldTimeInMonths: 10,
		CreditCost:       10,
	}); err != nil {
		t.Fatal(err)
	}
	// test a previously unseen upload
	if err := qmPublisher.PublishMessage(IPFSClusterPin{
		CID:              testCID2,
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
	// test private network detection
	if err := qmPublisher.PublishMessage(IPFSClusterPin{
		CID:              testCID,
		NetworkName:      "myprivatenetwork",
		UserName:         "testuser",
		HoldTimeInMonths: 10,
		CreditCost:       10,
	}); err != nil {
		t.Fatal(err)
	}
	// test invalid cid format
	if err := qmPublisher.PublishMessage(IPFSClusterPin{
		CID:              "notarealcid",
		NetworkName:      "public",
		UserName:         "testuser",
		HoldTimeInMonths: 10,
		CreditCost:       10,
	}); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_EmailSend(t *testing.T) {
	dev = true
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer := zaptest.NewLogger(t).Sugar()
	loggerPublisher := zaptest.NewLogger(t).Sugar()
	// setup our queue backend
	qmConsumer, err := New(EmailSendQueue, testRabbitAddress, false, dev, cfg, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(EmailSendQueue, testRabbitAddress, true, dev, cfg, loggerPublisher)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
	// test a email send
	if err := qmPublisher.PublishMessage(EmailSend{
		Subject:     "test email",
		Content:     "this is a test email",
		ContentType: "text/html",
		UserNames:   []string{"testuser", "testuser"},
		Emails:      []string{"testuser@example.com", "testuser2@example.com"},
	}); err != nil {
		t.Fatal(err)
	}
	// test a bad publish
	if err := qmPublisher.PublishMessage(""); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPNSEntry(t *testing.T) {
	dev = true
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer := zaptest.NewLogger(t).Sugar()
	loggerPublisher := zaptest.NewLogger(t).Sugar()
	// setup our queue backend
	qmConsumer, err := New(IpnsEntryQueue, testRabbitAddress, false, dev, cfg, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpnsEntryQueue, testRabbitAddress, true, dev, cfg, loggerPublisher)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
	// test a valid publish
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
	// test private network detection
	if err := qmPublisher.PublishMessage(IPNSEntry{
		CID:         testCID,
		LifeTime:    time.Minute,
		TTL:         time.Second,
		Resolve:     true,
		Key:         "testkey",
		UserName:    "testuser",
		NetworkName: "myprivatenetwork",
		CreditCost:  10,
	}); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPFSPin(t *testing.T) {
	dev = true
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer := zaptest.NewLogger(t).Sugar()
	loggerPublisher := zaptest.NewLogger(t).Sugar()
	// setup our queue backend
	qmConsumer, err := New(IpfsPinQueue, testRabbitAddress, false, dev, cfg, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpfsPinQueue, testRabbitAddress, true, dev, cfg, loggerPublisher)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
	// test a valid pin
	if err := qmPublisher.PublishMessage(IPFSPin{
		CID:              testCID,
		NetworkName:      "public",
		HoldTimeInMonths: 10},
	); err != nil {
		t.Fatal(err)
	}
	// test a bad publish
	if err := qmPublisher.PublishMessage(""); err != nil {
		t.Fatal(err)
	}
	// test a private network
	if err := qmPublisher.PublishMessage(IPFSPin{
		CID:              testCID,
		NetworkName:      "myprivatenetwork",
		HoldTimeInMonths: 10},
	); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	// set temporary log dig
	cfg.LogDir = "./tmp/"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	waitGroup.Wait()
}

// Does not conduct validation of whether or not a message was successfully processed
func TestQueue_IPFSKeyCreation(t *testing.T) {
	dev = true
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer := zaptest.NewLogger(t).Sugar()
	loggerPublisher := zaptest.NewLogger(t).Sugar()
	// setup our queue backend
	qmConsumer, err := New(IpfsKeyCreationQueue, testRabbitAddress, false, dev, cfg, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpfsKeyCreationQueue, testRabbitAddress, true, dev, cfg, loggerPublisher)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
	// test a normal publish
	if err := qmPublisher.PublishMessage(IPFSKeyCreation{
		UserName:    "testuser",
		Name:        "testuser-mykey",
		Type:        "rsa",
		Size:        2048,
		NetworkName: "public",
		CreditCost:  0},
	); err != nil {
		t.Fatal(err)
	}
	if err := qmPublisher.PublishMessage(IPFSKeyCreation{
		UserName:    "testuser",
		Name:        "mykey",
		Type:        "rsa",
		Size:        2048,
		NetworkName: "public",
		CreditCost:  0},
	); err != nil {
		t.Fatal(err)
	}
	// test a bad publish
	if err := qmPublisher.PublishMessage(""); err != nil {
		t.Fatal(err)
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, waitGroup, db, cfg); err != nil {
		t.Fatal(err)
	}
	waitGroup.Wait()
}

func TestQueue_IPFSKeyCreation_Failure(t *testing.T) {
	dev = true
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer := zaptest.NewLogger(t).Sugar()
	loggerPublisher := zaptest.NewLogger(t).Sugar()
	// setup our queue backend
	qmConsumer, err := New(IpfsKeyCreationQueue, testRabbitAddress, false, dev, cfg, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	qmPublisher, err := New(IpfsKeyCreationQueue, testRabbitAddress, true, dev, cfg, loggerPublisher)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := qmPublisher.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err := qmPublisher.PublishMessage(IPFSKeyCreation{
		UserName:    "testuser",
		Name:        "mykey",
		Type:        "rsa",
		Size:        2048,
		NetworkName: "public",
		CreditCost:  0},
	); err != nil {
		t.Fatal(err)
	}
	cfg.Services.Krab.TLS.CertPath = "/root"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, &sync.WaitGroup{}, db, cfg); err == nil {
		t.Fatal("expected error")
	}
}

func TestQueue_IPFSPin_Failure_RabbitMQ(t *testing.T) {
	dev = true
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer := zaptest.NewLogger(t).Sugar()
	// setup our queue backend
	qmConsumer, err := New(IpfsPinQueue, testRabbitAddress, false, dev, cfg, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	cfg.RabbitMQ.URL = "notarealurl"
	// we don't need time-out since this test will automatically fail
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, &sync.WaitGroup{}, db, cfg); err == nil {
		t.Fatal("expected error")
	}
}

func TestQueue_IPFSPin_Failure_LogFile(t *testing.T) {
	dev = true
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer := zaptest.NewLogger(t).Sugar()
	// setup our queue backend
	qmConsumer, err := New(IpfsPinQueue, testRabbitAddress, false, dev, cfg, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	cfg.LogDir = "/root/toor"
	// we don't need time-out since this test will automatically fail
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, &sync.WaitGroup{}, db, cfg); err == nil {
		t.Fatal("expected error")
	}
}

func TestQueue_IPNSEntry_Failure_Krab(t *testing.T) {
	dev = true
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loggerConsumer := zaptest.NewLogger(t).Sugar()
	// setup our queue backend
	qmConsumer, err := New(IpnsEntryQueue, testRabbitAddress, false, dev, cfg, loggerConsumer)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Services.Krab.TLS.CertPath = "/root/toor"
	// we don't need time-out since this test will automatically fail
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err = qmConsumer.ConsumeMessages(ctx, &sync.WaitGroup{}, db, cfg); err == nil {
		t.Fatal(err)
	}
}

func TestQueue_Bad_TLS_Config_CACertFile(t *testing.T) {
	dev = true
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	logger := zaptest.NewLogger(t).Sugar()
	cfg.RabbitMQ.TLSConfig.CACertFile = "../README.md"
	if _, err := New(IpnsEntryQueue, cfg.RabbitMQ.URL, false, dev, cfg, logger); err == nil {
		t.Fatal("expected error")
	}
	cfg.RabbitMQ.TLSConfig.CACertFile = "/root/toor"
	if _, err := New(IpnsEntryQueue, cfg.RabbitMQ.URL, false, dev, cfg, logger); err == nil {
		t.Fatal("expected error")
	}
}

func loadDatabase(cfg *config.TemporalConfig) (*gorm.DB, error) {
	dbm, err := database.New(cfg, database.Options{SSLModeDisable: true})
	if err != nil {
		return nil, err
	}
	return dbm.DB, nil
}
