package queue

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/streadway/amqp"
)

var IpfsQueue = "ipfs"
var IpfsClusterQueue = "ipfs-cluster"
var DatabaseFileAddQueue = "dfa-queue"
var DatabasePinAddQueue = "dpa-queue"
var IpnsUpdateQueue = "ipns-update-queue"
var IpfsPinQueue = "ipfs-pin-queue"
var IpfsFileQueue = "ipfs-file-queue"
var PinPaymentQueue = "pin-payment-queue"

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

type IPFSFile struct {
	BucketName       string `json:"bucket_name"`
	ObjectName       string `json:"object_name"`
	EthAddress       string `json:"eth_address"`
	NetworkName      string `json:"network_name"`
	HoldTimeInMonths string `json:"hold_time_in_months"`
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
	// Declare Non Default exchanges for the particular queue
	switch queueName {
	case PinExchange:
		err = qm.DeclareIPFSPinExchange()
		if err != nil {
			return nil, err
		}
	case ClusterPinExchange:
		err = qm.DeclareIPFSClusterPinExchange()
		if err != nil {
			return nil, err
		}
	case FileExchange:
		err = qm.DeclareIPFSFileExchange()
		if err != nil {
			return nil, err
		}
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
func (qm *QueueManager) ConsumeMessage(consumer, dbPass, dbURL, ethKeyFile, ethKeyPass, dbUser string, cfg *config.TemporalConfig) error {
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
		case IpfsClusterQueue:
			ProcessIpfsClusterQueue(msgs, db)
		case IpfsPinQueue:
			ProccessIPFSPins(msgs, db)
		case IpfsFileQueue:
			ProccessIPFSFiles(msgs, cfg, db)
		case PinPaymentQueue:
			ProcessPinPaymentConfirmation(msgs, db, cfg.Ethereum.Connection.IPC.Path, "0x0")
		default:
			log.Fatal("invalid queue name")
		}
	}()
	<-forever
	return nil
}

//PublishFanOutMessage is used to publish a message to a fanout exchange.
func (qm *QueueManager) PublishMessageWithExchange(body interface{}, exchangeName string) error {
	switch exchangeName {
	case PinExchange:
		break
	case ClusterPinExchange:
		break
	default:
		return errors.New("invalid exchange name provided")
	}
	bodyMarshaled, err := json.Marshal(body)
	if err != nil {
		return err
	}
	err = qm.Channel.Publish(
		exchangeName,  // exchange
		qm.Queue.Name, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         bodyMarshaled,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// PublishMessage is used to produce messages that are sent to the queue, with a worker queue (one consumer)
func (qm *QueueManager) PublishMessage(body interface{}) error {
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
	return nil
}

func (qm *QueueManager) Close() {
	qm.Connection.Close()
}
