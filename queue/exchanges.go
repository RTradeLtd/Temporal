package queue

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
func (qm *Manager) DeclareIPFSPinExchange() error {
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
func (qm *Manager) DeclareIPFSKeyExchange() error {
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
