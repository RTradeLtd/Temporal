package queue

import "errors"

const (
	// PinExchange is the name of the fanout exchange for regular ipfs pins
	PinExchange = "ipfs-pin"
	// PinExchangeKey is the key used for ipfs pin exchanges
	PinExchangeKey = "ipfs-pin-key"
	// IpfsKeyExchange is the exchange topic used for key creation requests
	IpfsKeyExchange = "ipfs-key-exchange"
	// IpfsKeyExchangeKey is the exchange key used for key creation requests
	IpfsKeyExchangeKey = "ipfs-key-exchange-key"
)

// DeclareIPFSPinExchange is used to declare the exchange used to handle ipfs pins
func (qm *Manager) declareIPFSPinExchange() error {
	return qm.channel.ExchangeDeclare(
		PinExchange, // name
		"fanout",    // type
		true,        // durable
		false,       // auto-delete
		false,       // internal
		false,       // no wait
		nil,         // args
	)
}

// DeclareIPFSKeyExchange is used to declare the exchange used to handle ipfs key creation requests
func (qm *Manager) declareIPFSKeyExchange() error {
	return qm.channel.ExchangeDeclare(
		IpfsKeyExchange, // name
		"fanout",        // type
		true,            // durable
		false,           // auto-delete
		false,           // internal
		false,           // no wait
		nil,             // args
	)
}

// SetupExchange is used to setup our various exchanges
func (qm *Manager) setupExchange(queueName string) error {
	switch queueName {
	case IpfsPinQueue:
		qm.ExchangeName = PinExchange
		return qm.declareIPFSPinExchange()
	case IpfsKeyCreationQueue:
		qm.ExchangeName = IpfsKeyExchange
		return qm.declareIPFSKeyExchange()
	default:
		return errors.New("invalid queue name for non default exchange")
	}
}
