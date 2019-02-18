package v2

import (
	clients "github.com/RTradeLtd/Temporal/grpc-clients"
	"github.com/RTradeLtd/Temporal/queue"
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
	kb1 *clients.KaasClient
	kb2 *clients.KaasClient
}
