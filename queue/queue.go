package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/streadway/amqp"
)

var IpfsQueue = "ipfs"
var IpfsClusterQueue = "ipfs-cluster"
var DatabaseFileAddQueue = "dfa-queue"
var DatabasePinAddQueue = "dpa-queue"
var PaymentRegisterQueue = "payment-register-queue"
var PaymentReceivedQueue = "payment-received-queue"
var PinPaymentRequestQueue = "pin-payment-request-queue"
var IpnsUpdateQueue = "ipns-update-queue"
var IpfsPinQueue = "ipfs-pin-queue"

// QueueManager is a helper struct to interact with rabbitmq
type QueueManager struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      *amqp.Queue
}

// IPFSPin is a struct used when sending pin request
type IPFSPin struct {
	CID         string `json:"cid"`
	NetworkName string `json:"network_name"`
	EthAddress  string `json:"eth_address"`
}

// DatabaseFileAdd is a struct used when sending data to rabbitmq
type DatabaseFileAdd struct {
	Hash             string `json:"hash"`
	HoldTimeInMonths int64  `json:"hold_time_in_months"`
	UploaderAddress  string `json:"uploader_address"`
	NetworkName      string `json:"network_name"`
}

// DatabasePinAdd is a struct used wehn sending data to rabbitmq
type DatabasePinAdd struct {
	Hash             string `json:"hash"`
	HoldTimeInMonths int64  `json:"hold_time_in_months"`
	UploaderAddress  string `json:"uploader_address"`
	NetworkName      string `json:"network_name"`
}

// PaymentRegister is a struct used when a payment has been regsitered and needs
// to be added to the database
type PaymentRegister struct {
	UploaderAddress string `json:"uploader_address"`
	CID             string `json:"cid"`
	HashedCID       string `json:"hash_cid"`
	PaymentID       string `json:"payment_id"`
}

// PinPaymentRequest is used by the frontend to submit a payment request
// to allow our authenticated backend to register a payment
type PinPaymentRequest struct {
	UploaderAddress   string   `json:"uploader_address"`
	CID               string   `json:"cid"`
	HoldTimeInMonths  int64    `json:"hold_time_in_months"`
	Method            uint8    `json:"method"`
	ChargeAmountInWei *big.Int `json:"charge_amount_in_wei"`
}

// PaymentReceived is used when we need to mark that
// a payment has been received, and we will upload
// the content
type PaymentReceived struct {
	UploaderAddress string `json:"uploader_address"`
	PaymentID       string `json:"payment_id"`
}

// IpfsClusterPin is used to handle pinning items to the cluster
// that have been pinned locally
type IpfsClusterPin struct {
	CID string `json:"content_hash"`
}

type IPNSUpdate struct {
	CID         string `json:"content_hash"`
	IPNSHash    string `json:"ipns_hash"`
	LifeTime    string `json:"life_time"`
	TTL         string `json:"ttl"`
	Key         string `json:"key"`
	Resolve     bool   `json:"resolve"`
	EthAddress  string `json:"eth_address"`
	NetworkName string `json:"network_name"`
}

// Initialize is used to connect to the given queue, for publishing or consuming purposes
func Initialize(queueName, connectionURL string) (*QueueManager, error) {
	conn, err := setupConnection(connectionURL)
	if err != nil {
		return nil, err
	}
	qm := QueueManager{Connection: conn}
	if err := qm.OpenChannel(); err != nil {
		return nil, err
	}
	if err := qm.DeclareQueue(queueName); err != nil {
		return nil, err
	}
	return &qm, nil
}

func setupConnection(connectionURL string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(connectionURL)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (qm *QueueManager) OpenChannel() error {
	ch, err := qm.Connection.Channel()
	if err != nil {
		return err
	}
	qm.Channel = ch
	return nil
}

// DeclareQueue is used to declare a queue for which messages will be sent to
func (qm *QueueManager) DeclareQueue(queueName string) error {
	// we declare the queue as durable so that even if rabbitmq server stops
	// our messages won't be lost
	q, err := qm.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}
	qm.Queue = &q
	return nil
}

// ConsumeMessage is used to consume messages that are sent to the queue
// Question, do we really want to ack messages that fail to be processed?
// Perhaps the error was temporary, and we allow it to be retried?
func (qm *QueueManager) ConsumeMessage(consumer, dbPass, dbURL, ethKeyFile, ethKeyPass, dbUser string) error {
	db, err := database.OpenDBConnection(dbPass, dbURL, dbUser)
	if err != nil {
		return err
	}
	// we use a false flag for auto-ack since we will use
	// manually acknowledgemnets to ensure message delivery
	// even if a worker dies
	msgs, err := qm.Channel.Consume(
		qm.Queue.Name, // queue
		consumer,      // consumer
		false,         // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	// So we don't cause hanging prcesses when consuming messages, it is processed in a goroutine
	go func() {
		// check the queue name
		switch qm.Queue.Name {
		// only parse database pin requests
		case DatabasePinAddQueue:
			ProcessDatabasePinAdds(msgs, db)
		// only parse datbase file requests
		case DatabaseFileAddQueue:
			ProcessDatabaseFileAdds(msgs, db)
		case PaymentRegisterQueue:
			ProcessPaymentRegisterQueue(msgs, db)
		case PaymentReceivedQueue:
			ProcessPaymentReceivedQueue(msgs, db)
		case PinPaymentRequestQueue:
			ProcessPinPaymentRequestQueue(msgs, db, ethKeyFile, ethKeyPass)
		case IpfsClusterQueue:
			ProcessIpfsClusterQueue(msgs, db)
		case IpfsPinQueue:
			ProccessIPFSPins(msgs, db)
		default:
			log.Fatal("invalid queue name")
		}
	}()
	<-forever
	return nil
}

// PublishMessage is used to produce messages that are sent to the queue
func (qm *QueueManager) PublishMessage(body interface{}) error {
	fmt.Println("publishing message")
	// we use a persistent delivery mode to combine with the durable queue
	bodyMarshaled, err := json.Marshal(body)
	if err != nil {
		return err
	}
	err = qm.Channel.Publish(
		"",            // exchange
		qm.Queue.Name, // routing key
		false,         // mandatory
		false,         //immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         bodyMarshaled,
		},
	)
	if err != nil {
		return err
	}
	fmt.Println("message published")
	return nil
}

func (qm *QueueManager) Close() {
	qm.Connection.Close()
}
