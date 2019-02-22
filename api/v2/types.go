package v2

import (
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/kaas"
)

// CreditRefund is a data object to contain refund information
type CreditRefund struct {
	Username string
	CallType string
	Cost     float64
}

type queues struct {
	pin     *queue.Manager
	cluster *queue.Manager
	email   *queue.Manager
	ipns    *queue.Manager
	key     *queue.Manager
	dash    *queue.Manager
	eth     *queue.Manager
}

// kaas key managers
type keys struct {
	kb1 *kaas.Client
	kb2 *kaas.Client
}
