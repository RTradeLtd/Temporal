package queue

const (
	// PinExchange is the name of the fanout exchange for regular ipfs pins
	PinExchange = "ipfs-pin"
	// PinExchangeKey is the key used for ipfs pin exchanges
	PinExchangeKey = "ipfs-pin-key"
	// PinRemovalExchange is the fanout exchange we use for pin removals
	PinRemovalExchange = "ipfs-pin-removal"
	// PinRemovalExchangeKey is the key used for pin removal exchanges
	PinRemovalExchangeKey = "ipfs-pin-removal-key"
	// IpfsKeyExchange is the exchange topic used for key creation requests
	IpfsKeyExchange = "ipfs-key-exchange"
	// IpfsKeyExchangeKey is the exchange key used for key creation requests
	IpfsKeyExchangeKey = "ipfs-key-exchange-key"
)

// DeclareIPFSPinRemovalExchange is used to declare the exchange used to handle ipfs pins
func (qm *QueueManager) DeclareIPFSPinRemovalExchange() error {
	return qm.Channel.ExchangeDeclare(
		PinRemovalExchange, // name
		"fanout",           // type
		true,               // durable
		false,              // auto-delete
		false,              // internal
		false,              // no wait
		nil,                // args
	)
}

// DeclareIPFSPinExchange is used to declare the exchange used to handle ipfs pins
func (qm *QueueManager) DeclareIPFSPinExchange() error {
	return qm.Channel.ExchangeDeclare(
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
func (qm *QueueManager) DeclareIPFSKeyExchange() error {
	return qm.Channel.ExchangeDeclare(
		IpfsKeyExchange, // name
		"fanout",        // type
		true,            // durable
		false,           // auto-delete
		false,           // internal
		false,           // no wait
		nil,             // args
	)
}
