package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/RTradeLtd/Temporal/payment_server"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/streadway/amqp"
)

var IpfsQueue = "ipfs"
var IpfsClusterQueue = "ipfs-cluster"
var DatabaseFileAddQueue = "dfa-queue"
var DatabasePinAddQueue = "dpa-queue"
var PaymentRegisterQueue = "payment-register-queue"
var PaymentReceivedQueue = "payment-received-queue"
var PinPaymentRequestQueue = "pin-payment-request-queue"

// QueueManager is a helper struct to interact with rabbitmq
type QueueManager struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      *amqp.Queue
}

// DatabaseFileAdd is a struct used when sending data to rabbitmq
type DatabaseFileAdd struct {
	Hash             string `json:"hash"`
	HoldTimeInMonths int64  `json:"hold_time_in_months"`
	UploaderAddress  string `json:"uploader_address"`
}

// DatabasePinAdd is a struct used wehn sending data to rabbitmq
type DatabasePinAdd struct {
	Hash             string `json:"hash"`
	HoldTimeInMonths int64  `json:"hold_time_in_months"`
	UploaderAddress  string `json:"uploader_address"`
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
			if ethKeyFile == "" || ethKeyPass == "" {
				log.Fatal("no valid key parameters passed")
			}
			pm, err := payment_server.NewPaymentManager(true, ethKeyFile, ethKeyPass, db)
			if err != nil {
				log.Fatal(err)
			}
			var b [32]byte
			for d := range msgs {
				var ppr PinPaymentRequest
				fmt.Println("unmarshaling data")
				err := json.Unmarshal(d.Body, &ppr)
				if err != nil {
					fmt.Println("error unmarshaling data ", err)
					d.Ack(false)
					continue
				}
				ethAddress := ppr.UploaderAddress
				contentHash := ppr.CID
				retentionPeriod := ppr.HoldTimeInMonths
				chargeAmountInWei := ppr.ChargeAmountInWei
				method := ppr.Method
				data := []byte(contentHash)
				hashedCIDByte := crypto.Keccak256(data)
				hashedCID := common.BytesToHash(hashedCIDByte)
				copy(b[:], hashedCID.Bytes()[:32])
				tx, err := pm.Contract.RegisterPayment(pm.Auth, common.HexToAddress(ethAddress), b, big.NewInt(retentionPeriod), chargeAmountInWei, method)
				if err != nil {
					fmt.Println("error submitting payment ", err)
					d.Ack(false)
					continue
				}
				// TODO: add call to database
				fmt.Printf("%+v\n", tx)
				d.Ack(false)
			}
		case IpfsClusterQueue:
			var clusterPin IpfsClusterPin
			clusterManager := rtfs_cluster.Initialize()
			for d := range msgs {
				err := json.Unmarshal(d.Body, &clusterPin)
				if err != nil {
					fmt.Println("error unmarshaling data ", err)
					// TODO: handle error
					d.Ack(false)
					continue
				}
				contentHash := clusterPin.CID
				decodedContentHash, err := clusterManager.DecodeHashString(contentHash)
				if err != nil {
					fmt.Println("error decoded content hash to cid ", err)
					//TODO: handle error
					d.Ack(false)
					continue
				}
				err = clusterManager.Pin(decodedContentHash)
				if err != nil {
					fmt.Println("error pinning to cluster ", err)
					//TODO: handle error
					d.Ack(false)
					continue
				}
				fmt.Println("content pinned to cluster ", contentHash)
				d.Ack(false)

			}
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
