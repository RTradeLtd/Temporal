package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/streadway/amqp"
)

var DatabaseFileAddQueue = "dfa-queue"
var IpfsPinQueue = "ipfs-pin-queue"
var IpfsFileQueue = "ipfs-file-queue"
var IpfsClusterPinQueue = "ipfs-cluster-add-queue"
var PinPaymentConfirmationQueue = "pin-payment-confirmation-queue"
var PinPaymentSubmissionQueue = "pin-payment-submission-queue"
var EmailSendQueue = "email-send-queue"
var IpnsEntryQueue = "ipns-entry-queue"
var IpfsPinRemovalQueue = "ipns-pin-removal-queue"
var IpfsKeyCreationQueue = "ipfs-key-creation-queue"

var AdminEmail = "temporal.reports@rtradetechnologies.com"

// QueueManager is a helper struct to interact with rabbitmq
type QueueManager struct {
	Connection   *amqp.Connection
	Channel      *amqp.Channel
	Queue        *amqp.Queue
	Logger       *log.Logger
	QueueName    string
	Service      string
	ExchangeName string
}

// IPFSKeyCreation is a message used for processing key creation
// only supported for the public IPFS network at the moment
type IPFSKeyCreation struct {
	UserName    string `json:"user_name"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Size        int    `json:"size"`
	NetworkName string `json:"network_name"`
}

// IPFSPin is a struct used when sending pin request
type IPFSPin struct {
	CID              string `json:"cid"`
	NetworkName      string `json:"network_name"`
	UserName         string `json:"user_name"`
	HoldTimeInMonths int64  `json:"hold_time_in_months"`
}

type IPFSFile struct {
	BucketName       string `json:"bucket_name"`
	ObjectName       string `json:"object_name"`
	UserName         string `json:"user_name"`
	NetworkName      string `json:"network_name"`
	HoldTimeInMonths string `json:"hold_time_in_months"`
}

// IPFSClusterPin is a queue message used when sending a message to the cluster to pin content
type IPFSClusterPin struct {
	CID              string `json:"cid"`
	NetworkName      string `json:"network_name"`
	UserName         string `json:"user_name"`
	HoldTimeInMonths int64  `json:"hold_time_in_months"`
}

type IPFSPinRemoval struct {
	ContentHash string `json:"content_hash"`
	NetworkName string `json:"network_name"`
	UserName    string `json:"user_name"`
}

// DatabaseFileAdd is a struct used when sending data to rabbitmq
type DatabaseFileAdd struct {
	Hash             string `json:"hash"`
	HoldTimeInMonths int64  `json:"hold_time_in_months"`
	UserName         string `json:"user_name"`
	NetworkName      string `json:"network_name"`
}

type IPNSUpdate struct {
	CID         string `json:"content_hash"`
	IPNSHash    string `json:"ipns_hash"`
	LifeTime    string `json:"life_time"`
	TTL         string `json:"ttl"`
	Key         string `json:"key"`
	Resolve     bool   `json:"resolve"`
	UserName    string `json:"user_name"`
	NetworkName string `json:"network_name"`
}

func (qm *QueueManager) setupLogging() error {
	logFileName := fmt.Sprintf("/var/log/temporal/%s_serice.log", qm.QueueName)
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	logger := log.New()
	logger.Out = logFile
	qm.Logger = logger
	qm.Logger.Info("Logging initialized")
	return nil
}

func (qm *QueueManager) parseQueueName(queueName string) error {
	host, err := os.Hostname()
	if err != nil {
		return err
	}
	qm.QueueName = fmt.Sprintf("%s+%s", host, queueName)
	return nil
}

// Initialize is used to connect to the given queue, for publishing or consuming purposes
func Initialize(queueName, connectionURL string, publish, service bool) (*QueueManager, error) {
	conn, err := setupConnection(connectionURL)
	if err != nil {
		return nil, err
	}
	qm := QueueManager{Connection: conn}
	if err := qm.OpenChannel(); err != nil {
		return nil, err
	}

	qm.QueueName = queueName
	qm.Service = queueName

	if service {
		err = qm.setupLogging()
		if err != nil {
			return nil, err
		}
	}
	// Declare Non Default exchanges for the particular queue
	switch queueName {
	case IpfsPinRemovalQueue:
		err = qm.parseQueueName(queueName)
		if err != nil {
			return nil, err
		}
		err = qm.DeclareIPFSPinRemovalExchange()
		if err != nil {
			return nil, err
		}
		qm.ExchangeName = PinRemovalExchange
		if publish {
			return &qm, nil
		}
	case IpfsPinQueue:
		err = qm.parseQueueName(queueName)
		if err != nil {
			return nil, err
		}
		err = qm.DeclareIPFSPinExchange()
		if err != nil {
			return nil, err
		}
		qm.ExchangeName = PinExchange
		if publish {
			return &qm, nil
		}
	case IpfsKeyCreationQueue:
		err = qm.parseQueueName(queueName)
		if err != nil {
			return nil, err
		}
		err = qm.DeclareIPFSKeyExchange()
		if err != nil {
			return nil, err
		}
		qm.ExchangeName = IpfsKeyExchange
		if publish {
			return &qm, nil
		}
	}
	if err := qm.DeclareQueue(); err != nil {
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
	if qm.Logger != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("channel opened")
	}
	qm.Channel = ch
	return nil
}

// DeclareQueue is used to declare a queue for which messages will be sent to
func (qm *QueueManager) DeclareQueue() error {
	// we declare the queue as durable so that even if rabbitmq server stops
	// our messages won't be lost
	q, err := qm.Channel.QueueDeclare(
		qm.QueueName, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}
	if qm.Logger != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("queue declared")
	}
	qm.Queue = &q
	return nil
}

// ConsumeMessage is used to consume messages that are sent to the queue
// Question, do we really want to ack messages that fail to be processed?
// Perhaps the error was temporary, and we allow it to be retried?
func (qm *QueueManager) ConsumeMessage(consumer, dbPass, dbURL, dbUser string, cfg *config.TemporalConfig) error {
	db, err := database.OpenDBConnection(database.DBOptions{
		User: dbUser, Password: dbPass, Address: dbURL})
	if err != nil {
		return err
	}

	// ifs the queue is using an exchange, we will need to bind the queue to the exchange
	switch qm.ExchangeName {
	case PinRemovalExchange:
		err = qm.Channel.QueueBind(
			qm.QueueName,       // name of the queue
			"",                 // routing key
			PinRemovalExchange, // exchange
			false,              // noWait
			nil,                // arguments
		)
		if err != nil {
			return err
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("queue bound")
	case PinExchange:
		err = qm.Channel.QueueBind(
			qm.QueueName, // name of the queue
			"",           // routing key
			PinExchange,  // exchange
			false,        // noWait
			nil,          // arguments
		)
		if err != nil {
			return err
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("queue bound")
	case IpfsKeyExchange:
		err = qm.Channel.QueueBind(
			qm.QueueName,    // name of the queue
			"",              // routing key
			IpfsKeyExchange, // exchange
			false,           // no wait
			nil,             // arguments
		)
		if err != nil {
			return err
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("queue bound")
	default:
		break
	}

	// consider moving to true for auto-ack
	msgs, err := qm.Channel.Consume(
		qm.QueueName, // queue
		consumer,     // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return err
	}

	// check the queue name
	switch qm.Service {
	// only parse datbase file requests
	case DatabaseFileAddQueue:
		qm.ProcessDatabaseFileAdds(msgs, db)
	case IpfsPinQueue:
		err = qm.ProccessIPFSPins(msgs, db, cfg)
		if err != nil {
			return err
		}
	case IpfsFileQueue:
		err = qm.ProccessIPFSFiles(msgs, cfg, db)
		if err != nil {
			return err
		}
	case PinPaymentConfirmationQueue:
		err = qm.ProcessPinPaymentConfirmation(msgs, db, cfg)
		if err != nil {
			return err
		}
	case PinPaymentSubmissionQueue:
		err = qm.ProcessPinPaymentSubmissions(msgs, db, cfg)
		if err != nil {
			return err
		}
	case EmailSendQueue:
		err = qm.ProcessMailSends(msgs, cfg)
		if err != nil {
			return err
		}
	case IpnsEntryQueue:
		err = qm.ProcessIPNSEntryCreationRequests(msgs, db, cfg)
		if err != nil {
			return err
		}
	case IpfsPinRemovalQueue:
		err = qm.ProcessIPFSPinRemovals(msgs, cfg, db)
		if err != nil {
			return err
		}
	case IpfsKeyCreationQueue:
		err = qm.ProcessIPFSKeyCreation(msgs, db, cfg)
		if err != nil {
			return err
		}
	case IpfsClusterPinQueue:
		err = qm.ProcessIPFSClusterPins(msgs, cfg, db)
		if err != nil {
			return err
		}
	default:
		log.Fatal("invalid queue name")
	}
	return nil
}

//PublishMessageWithExchange is used to publish a message to a given exchange
func (qm *QueueManager) PublishMessageWithExchange(body interface{}, exchangeName string) error {
	switch exchangeName {
	case PinExchange:
		break
	case PinRemovalExchange:
		break
	case IpfsKeyExchange:
		break
	default:
		return errors.New("invalid exchange name provided")
	}
	bodyMarshaled, err := json.Marshal(body)
	if err != nil {
		return err
	}
	err = qm.Channel.Publish(
		exchangeName, // exchange
		"",           // routing key
		false,        // mandatory
		false,        // immediate
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

func (qm *QueueManager) Close() error {
	return qm.Connection.Close()
}
