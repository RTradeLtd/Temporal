package queue

import (
	"github.com/streadway/amqp"
)

var ipfsQueue = "ipfs"
var ipfsClusterQueue = "ipfs-cluster"

type QueueManager struct {
	Connection *amqp.Connection
}

func Setup() (*amqp.Connection, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (qm *QueueManager) Close() {
	qm.Connection.Close()
}

func (qm *QueueManager) DeclareQueue(queueName string) {

}
