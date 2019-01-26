package queue

import "errors"

const (
	// IpfsKeyExchange is the exchange topic used for key creation requests
	IpfsKeyExchange = "ipfs-key-exchange"
	// IpfsKeyExchangeKey is the exchange key used for key creation requests
	IpfsKeyExchangeKey = "ipfs-key-exchange-key"
)

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
func (qm *Manager) setupExchange(queueType Queue) error {
	switch queueType {
	case IpfsKeyCreationQueue:
		qm.ExchangeName = IpfsKeyExchange
		return qm.declareIPFSKeyExchange()
	default:
		return errors.New("invalid queue name for non default exchange")
	}
}
