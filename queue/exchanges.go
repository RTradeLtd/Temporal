package queue

const (
	// PinExchange is the name of the fanout exchange for regular ipfs pins
	PinExchange = "ipfs-pin"
	// ClusterPinExchange is the name of the fanout exchange for cluster ipfs pins
	ClusterPinExchange = "ipfs-cluster-pin"
	// FileExchange is the name of the fanout exchange for regular ipfs files
	FileExchange = "ipfs-file"
)

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

// DeclareIPFSClusterPinExchange is used to declare the exchange used to handle ipfs cluster pins
func (qm *QueueManager) DeclareIPFSClusterPinExchange() error {
	return qm.Channel.ExchangeDeclare(
		ClusterPinExchange, // name
		"fanout",           // type
		true,               // durable
		false,              // auto-delete
		false,              // internal
		false,              // no wait
		nil,                // args
	)
}

// DeclareIPFSFileExchange is sued to declare the exchange used to handle ipfs files
func (qm *QueueManager) DeclareIPFSFileExchange() error {
	return qm.Channel.ExchangeDeclare(
		FileExchange, // name
		"fanout",     // type
		true,         // durable
		false,        // auto-delete
		false,        // internal
		false,        // no wait
		nil,          // args
	)
}
