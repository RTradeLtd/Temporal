package v2

import "github.com/RTradeLtd/Temporal/queue"

// CreditRefund is a data object to contain refund information
type CreditRefund struct {
	Username string
	CallType string
	Cost     float64
}

type queues struct {
	pin        *queue.Manager
	file       *queue.Manager
	cluster    *queue.Manager
	email      *queue.Manager
	ipns       *queue.Manager
	key        *queue.Manager
	database   *queue.Manager
	dash       *queue.Manager
	payConfirm *queue.Manager
}
