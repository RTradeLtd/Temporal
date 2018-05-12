package queue

import (
	"log"

	"github.com/streadway/amqp"
)

var ipfsQueue = "ipfs"
var ipfsClusterQueue = "ipfs-cluster"

type QueueManager struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      *amqp.Queue
}

func Setup() (*amqp.Connection, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (qm *QueueManager) OpenChannel() {
	ch, err := qm.Connection.Channel()
	if err != nil {
		log.Fatal(err)
	}
	qm.Channel = ch
}

func (qm *QueueManager) DeclareQueue() {
	q, err := qm.Channel.QueueDeclare(
		"name", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		log.Fatal(err)
	}
	qm.Queue = &q
}

func (qm *QueueManager) PublishMessage(msg string) {
	body := msg
	err := qm.Channel.Publish(
		"",            // exchange
		qm.Queue.Name, // routing key
		false,         // mandatory
		false,         //immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func (qm *QueueManager) Close() {
	qm.Connection.Close()
}
