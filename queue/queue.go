package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/api/rtfs"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/streadway/amqp"
)

var IpfsQueue = "ipfs"
var IpfsClusterQueue = "ipfs-cluster"
var DatabaseFileAddQueue = "dfa-queue"
var DatabasePinAddQueue = "dpa-queue"
var PaymentRegisterQueue = "payment-register-queue"
var PaymentReceivedQueue = "payment-received-queue"

// QueueManager is a helper struct to interact with rabbitmq
type QueueManager struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      *amqp.Queue
}

// TODO: cluster pinning will be moved to a rabbitmq system shortly
type IpfsClusterPin struct{}

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

// PaymentReceived is used when we need to mark that
// a payment has been received, and we will upload
// the content
type PaymentReceived struct {
	UploaderAddress string `json:"uploader_address"`
	PaymentID       string `json:"payment_id"`
}

// Initialize is used to connect to the given queue, for publishing or consuming purposes
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
	// So we don't cause hanging prcesses when consuming messages, it is processed in a goroutine
	go func() {
		// check the queue name
		switch qm.Queue.Name {
		// only parse database pin requests
		case DatabasePinAddQueue:
			for d := range msgs {
				if d.Body != nil {
					dpa := DatabasePinAdd{}
					upload := models.Upload{}
					log.Printf("receive a message: %s", d.Body)
					// unmarshal the message into the struct
					// if it can't be decoded into dpa struct, acknowledge message receival and continue to the nextm essage
					err := json.Unmarshal(d.Body, &dpa)
					// make this system more robust
					if err != nil {
						d.Ack(false)
						continue
					}
					upload.Hash = dpa.Hash
					upload.HoldTimeInMonths = dpa.HoldTimeInMonths
					upload.Type = "pin"
					upload.UploadAddress = dpa.UploaderAddress
					// get current time
					currTime := time.Now()
					// get the hold time from in64 and convert to int
					holdTime, err := strconv.Atoi(fmt.Sprint(dpa.HoldTimeInMonths))
					if err != nil {
						d.Ack(false)
						continue
					}
					// get the date the file wiill be garbage collected by adding the number of months
					gcd := currTime.AddDate(0, holdTime, 0)
					lastUpload := models.Upload{
						Hash: dpa.Hash,
					}
					db.Last(&lastUpload)
					// check to see whether or not the file will be garbage collected before the last upload
					// if so we'll skip
					if lastUpload.GarbageCollectDate.Unix() >= gcd.Unix() {
						fmt.Println("skipping since we already have an instance that will be GC'd later")
						d.Ack(false)
						continue
					}
					upload.UploaderAddresses = append(lastUpload.UploaderAddresses, dpa.UploaderAddress)
					upload.GarbageCollectDate = gcd
					db.Create(&upload)
					// submit message acknowledgement
					d.Ack(false)
				}
			}
		// only parse datbase file requests
		case DatabaseFileAddQueue:
			for d := range msgs {
				if d.Body != nil {
					if d.Body != nil {
						dfa := DatabaseFileAdd{}
						upload := models.Upload{}
						// unmarshal the message body into the dfa struct
						err := json.Unmarshal(d.Body, &dfa)
						if err != nil {
							d.Ack(false)
							continue
						}
						// convert the int64 to an int. We need to make sure to add a check that we won't overflow
						holdTime, err := strconv.Atoi(fmt.Sprintf("%v", dfa.HoldTimeInMonths))
						if err != nil {
							d.Ack(false)
							continue
						}
						// we will take the current time, and add the number of months to get the date
						// that we will garbage collect this from our repo
						gcd := time.Now().AddDate(0, holdTime, 0)
						upload.Hash = dfa.Hash
						upload.HoldTimeInMonths = dfa.HoldTimeInMonths
						upload.Type = "file"
						upload.UploadAddress = dfa.UploaderAddress
						upload.GarbageCollectDate = gcd
						lastUpload := models.Upload{
							Hash: dfa.Hash,
						}
						// retrieve the last upload matching this hash.
						// this upload will have the latest Garbage Collect Date
						db.Last(&lastUpload)
						// check the garbage collect dates, if the current upload to be pinned will be
						// GCd before the latest one from the database, we will skip it
						// however if it will be GCd at a later date, we will keep it
						// and update the database
						if lastUpload.GarbageCollectDate.Unix() >= upload.GarbageCollectDate.Unix() {
							d.Ack(false)
							// skip the rest of the message, preventing a database record from being created
							continue
						}
						upload.UploaderAddresses = append(lastUpload.UploaderAddresses, dfa.UploaderAddress)
						// we have a valid upload request, so lets store it to the database
						db.Create(&upload)
						d.Ack(false)
					}
				}
			}
		case PaymentRegisterQueue:
			for d := range msgs {
				var nullTime time.Time
				var payment models.Payment
				pr := PaymentRegister{}
				fmt.Println("unmarshaling payment registered data")
				err := json.Unmarshal(d.Body, &pr)
				if err != nil {
					fmt.Println("error unmarshaling data", err)
					d.Ack(false)
					continue
				}
				fmt.Println("data unmarshaled successfully")
				fmt.Println(pr.PaymentID)
				db.Where("payment_id = ?", pr.PaymentID).Find(&payment)
				fmt.Println(payment)
				if payment.CreatedAt != nullTime {
					fmt.Println("payment is already in the database")
					d.Ack(false)
					continue
				}
				payment.CID = pr.CID
				payment.HashedCID = pr.HashedCID
				payment.PaymentID = pr.PaymentID
				payment.Paid = false
				fmt.Println(payment)
				fmt.Println("creating payment in database")
				db.Create(&payment)
				fmt.Println("payment entry in database created")
				d.Ack(false)
			}
		case PaymentReceivedQueue:
			ipfsManager := rtfs.Initialize("")
			for d := range msgs {
				var nullTime time.Time
				var payment models.Payment
				pr := PaymentReceived{}
				fmt.Println("unmarshaling payment received data")
				err := json.Unmarshal(d.Body, &pr)
				if err != nil {
					fmt.Println("error unmarhsaling data", err)
					d.Ack(false)
					continue
				}
				fmt.Println("data unmarshaled successfully")
				db.First(&payment).Where("payment_id = ?", payment.PaymentID)
				if payment.CreatedAt == nullTime {
					fmt.Println("payment is not a valid payment")
					d.Ack(false)
					continue
				}
				fmt.Println("updating database with payment received")
				payment.Paid = true
				db.Update(&payment)
				fmt.Println("database updated successfully, pinning to node")
				go ipfsManager.Pin(payment.CID)
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
