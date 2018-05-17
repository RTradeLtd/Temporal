package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/rtfs_cluster"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/streadway/amqp"
)

/*
NOTES:
	For 1 month we use 730 hours
*/

var IpfsQueue = "ipfs"
var IpfsClusterQueue = "ipfs-cluster"
var DatabaseFileAddQueue = "dfa-queue"
var DatabasePinAddQueue = "dpa-queue"

type QueueManager struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      *amqp.Queue
}

type IpfsClusterPin struct{}

type DatabaseFileAdd struct {
	Hash             string `json:"hash"`
	HoldTimeInMonths int64  `json:"hold_time_in_months"`
	UploaderAddress  string `json:"uploader_address"`
}

type DatabasePinAdd struct {
	Hash             string `json:"hash"`
	HoldTimeInMonths int64  `json:"hold_time_in_months"`
	UploaderAddress  string `json:"uploader_address"`
}

func Initialize(queueName string) (*QueueManager, error) {
	conn, err := setupConnection()
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

func setupConnection() (*amqp.Connection, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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
func (qm *QueueManager) ConsumeMessage(consumer string) error {
	db := database.OpenDBConnection()
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
	go func() {
		for d := range msgs {
			switch qm.Queue.Name {
			case DatabasePinAddQueue:
				if d.Body != nil {
					dpa := DatabasePinAdd{}
					upload := models.Upload{}
					log.Printf("receive a message: %s", d.Body)
					err := json.Unmarshal(d.Body, &dpa)
					if err != nil {
						continue
					}
					upload.Hash = dpa.Hash
					upload.HoldTimeInMonths = dpa.HoldTimeInMonths
					upload.Type = "pin"
					upload.UploadAddress = dpa.UploaderAddress
					currTime := time.Now()
					holdTime, err := strconv.Atoi(fmt.Sprint(dpa.HoldTimeInMonths))
					if err != nil {
						continue
					}
					gcd := currTime.AddDate(0, holdTime, 0)
					lastUpload := models.Upload{
						Hash: dpa.Hash,
					}
					db.Last(&lastUpload)
					fmt.Printf("%+v\n", lastUpload)
					upload.GarbageCollectDate = gcd
					db.Create(&upload)
					// submit message acknowledgement
					d.Ack(false)
					// TODO: change this to an async process
					cm := rtfs_cluster.Initialize()
					decoded := cm.DecodeHashString(dpa.Hash)
					err = cm.Pin(decoded)
					if err != nil {
						log.Fatal(err)
					}
				}
			case DatabaseFileAddQueue:
				if d.Body != nil {
					if d.Body != nil {
						dfa := DatabaseFileAdd{}
						upload := models.Upload{}
						log.Printf("receive a message: %s", d.Body)
						err := json.Unmarshal(d.Body, &dfa)
						if err != nil {
							continue
						}
						upload.Hash = dfa.Hash
						upload.HoldTimeInMonths = dfa.HoldTimeInMonths
						upload.Type = "file"
						upload.UploadAddress = dfa.UploaderAddress
						db.Create(&upload)
						// submit message acknowledgement
						d.Ack(false)
						// TODO: add cluster pin event
					}
				}
			default:
				log.Fatal("invalid queue name")
			}
		}
	}()
	<-forever
	return nil
}

// PublishMessage is used to produce messages that are sent to the queue
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
