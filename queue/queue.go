package queue

import (
	"log"

	"github.com/streadway/amqp"
)

var IpfsQueue = "ipfs"
var IpfsClusterQueue = "ipfs-cluster"

type QueueManager struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      *amqp.Queue
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
	q, err := qm.Channel.QueueDeclare(
		queueName, // name
		false,     // durable
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

func (qm *QueueManager) ConsumeMessage(consumer string) error {
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
			log.Printf("receive a message: %s", d.Body)
			// submit message acknowledgement
			d.Ack(false)
		}
	}()
	<-forever
	return nil
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
