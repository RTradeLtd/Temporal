package queue

import (
	"time"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/gorm"
	"go.uber.org/zap"

	"github.com/streadway/amqp"
)

// Various variables and types used by our queue package

// Queue is a typed string used to declare the various queue names
type Queue string

func (qt Queue) String() string {
	return string(qt)
}

var (
	dev     = false
	nilTime time.Time
	// IpfsPinQueue is a queue used for ipfs pins
	IpfsPinQueue Queue = "ipfs-pin-queue"
	// IpfsClusterPinQueue is a queue used for ipfs cluster pins
	IpfsClusterPinQueue Queue = "ipfs-cluster-add-queue"
	// EmailSendQueue is a queue used to handle sending email messages
	EmailSendQueue Queue = "email-send-queue"
	// IpnsEntryQueue is a queue used to handle ipns entry creation
	IpnsEntryQueue Queue = "ipns-entry-queue"
	// IpfsKeyCreationQueue is a queue used to handle ipfs key creation
	IpfsKeyCreationQueue Queue = "ipfs-key-creation-queue"
	// EthPaymentConfirmationQueue is a queue used to handle ethereum based payment confirmations
	EthPaymentConfirmationQueue Queue = "eth-payment-confirmation-queue"
	// DashPaymentConfirmationQueue is a queue used to handle confirming dash payments
	DashPaymentConfirmationQueue Queue = "dash-payment-confirmation-queue"
	// AdminEmail is the email used to notify RTrade about any critical errors
	AdminEmail = "temporal.reports@rtradetechnologies.com"
	// IpfsPinFailedContent is a to-be formatted message sent on IPFS pin failures
	IpfsPinFailedContent = "Pin failed for content hash %s on IPFS network %s, for reason %s"
	// IpfsPinFailedSubject is a subject for IPFS pin failed messages
	IpfsPinFailedSubject = "IPFS Pin Failed"
	// IpfsFileFailedContent is a to be formatted message sent on ipfs add failures
	IpfsFileFailedContent = "IPFS File Add Failed for object name %s on IPFS network %s"
	// IpfsFileFailedSubject is a subject for ipfs file add fails
	IpfsFileFailedSubject = "IPFS File Add Failed"
	// IpfsPrivateNetworkUnauthorizedSubject is a subject whenever someone tries to access a bad private network
	IpfsPrivateNetworkUnauthorizedSubject = "Unauthorized access to IPFS private network"
	// IpfsInitializationFailedSubject is a subject used when connecting to ipfs fails
	IpfsInitializationFailedSubject = "Connection to IPFS failed"
	// IpnsEntryFailedSubject is a subject sent upon IPNS failures
	IpnsEntryFailedSubject = "IPNS Entry Creation Failed"
	// IpnsEntryFailedContent is the content used when sending an email for IPNS entry creation failures
	IpnsEntryFailedContent = "IPNS Entry creation failed for content hash %s using key %s for reason %s"
	// PaymentConfirmationFailedSubject is a subject used when payment confirmations fail
	PaymentConfirmationFailedSubject = "Payment Confirmation Failed"
	// PaymentConfirmationFailedContent is a content used when a payment confirmation failure occurs
	PaymentConfirmationFailedContent = "Payment failed for content hash %s with error %s"
	// ErrReconnect is an error emitted when a protocol connection error occurs
	// It is used to signal reconnect of queue consumers and publishers
	ErrReconnect = "protocol connection error, reconnect"
)

// Manager is a helper struct to interact with rabbitmq
type Manager struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	queue        *amqp.Queue
	l            *zap.SugaredLogger
	db           *gorm.DB
	cfg          *config.TemporalConfig
	ErrCh        chan *amqp.Error
	QueueName    Queue
	ExchangeName string
}

// Queue Messages - These are used to format messages to send through rabbitmq

// IPFSKeyCreation is a message used for processing key creation
// only supported for the public IPFS network at the moment
type IPFSKeyCreation struct {
	UserName    string  `json:"user_name"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Size        int     `json:"size"`
	NetworkName string  `json:"network_name"`
	CreditCost  float64 `json:"credit_cost"`
}

// IPFSPin is a struct used when sending pin request
type IPFSPin struct {
	CID              string  `json:"cid"`
	NetworkName      string  `json:"network_name"`
	UserName         string  `json:"user_name"`
	HoldTimeInMonths int64   `json:"hold_time_in_months"`
	CreditCost       float64 `json:"credit_cost"`
	Size             int64   `json:"size"`
	JWT              string  `json:"jwt,omitempty"`
}

// IPFSClusterPin is a queue message used when sending a message to the cluster to pin content
type IPFSClusterPin struct {
	CID              string  `json:"cid"`
	NetworkName      string  `json:"network_name"`
	UserName         string  `json:"user_name"`
	HoldTimeInMonths int64   `json:"hold_time_in_months"`
	Size             int64   `json:"size"`
	CreditCost       float64 `json:"credit_cost"`
}

// IPNSUpdate is our message for the ipns update queue
type IPNSUpdate struct {
	CID         string  `json:"content_hash"`
	IPNSHash    string  `json:"ipns_hash"`
	LifeTime    string  `json:"life_time"`
	TTL         string  `json:"ttl"`
	Key         string  `json:"key"`
	Resolve     bool    `json:"resolve"`
	UserName    string  `json:"user_name"`
	NetworkName string  `json:"network_name"`
	CreditCost  float64 `json:"credit_cost"`
}

// EmailSend is a helper struct used to contained formatted content ot send as an email
type EmailSend struct {
	Subject     string   `json:"subject"`
	Content     string   `json:"content"`
	ContentType string   `json:"content_type"`
	UserNames   []string `json:"user_names"`
	Emails      []string `json:"emails,omitempty"`
}

// IPNSEntry is used to hold relevant information needed to process IPNS entry creation requests
type IPNSEntry struct {
	CID         string        `json:"cid"`
	LifeTime    time.Duration `json:"life_time"`
	TTL         time.Duration `json:"ttl"`
	Resolve     bool          `json:"resolve"`
	Key         string        `json:"key"`
	UserName    string        `json:"user_name"`
	NetworkName string        `json:"network_name"`
	CreditCost  float64       `json:"credit_cost"`
}

// DashPaymenConfirmation is a message used to signal processing of a dash payment
type DashPaymenConfirmation struct {
	UserName         string `json:"user_name"`
	PaymentForwardID string `json:"payment_forward_id"`
	PaymentNumber    int64  `json:"payment_number"`
}

// EthPaymentConfirmation is a message used to confirm an ethereum based payment
type EthPaymentConfirmation struct {
	UserName      string `json:"user_name"`
	PaymentNumber int64  `json:"payment_number"`
}
