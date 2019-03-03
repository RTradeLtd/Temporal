package testbedi

import (
	"context"
	"io"
)

// NewNodeFunc constructs a node implementing the Core interface. It is provided
// a path to an already created directory `dir`, as well as a map of attributes
// which can be supplied to shape process execution.
// Examples of attributes include: which binary to use, docker image, cpu/ram
// limits, or any other information that may be required to property setup or
// manage the node.
type NewNodeFunc func(dir string, attrs map[string]string) (Core, error)

// GetAttrListFunc returns a list of attribute names that can be queried from
// the node. These attributes may include those can be set from the NewNodeFunc,
// or additional attributes at may only be available after initialization.
// Attributes returned should be queriable through the Attribute interface.
// Examples include: api address, peerid, cpu/ram limits, jitter.
type GetAttrListFunc func() []string

// GetAttrDescFunc returns the description of the attribute `attr`
type GetAttrDescFunc func(attr string) (string, error)

type Libp2p interface {
	// PeerID returns the peer id
	PeerID() (string, error)
	// APIAddr returns the multiaddr for the api
	APIAddr() (string, error)
	// SwarmAddrs returns the swarm addrs for the node
	SwarmAddrs() ([]string, error)
}

type Config interface {
	Core
	// Config returns the configuration of the node
	Config() (interface{}, error)
	// WriteConfig writes the configuration of the node
	WriteConfig(interface{}) error
}

// Attributes are ways to shape process execution and additional information that alters the
// environment the process executes in
type Attribute interface {
	Core
	// Attr returns the value of attr
	Attr(attr string) (string, error)
	// SetAttr sets the attr to val
	SetAttr(attr string, val string) error
	// GetAttrList returns a list of attrs that can be retrieved
	GetAttrList() []string
	// GetAttrDesc returns the description of attr
	GetAttrDesc(attr string) (string, error)
	/* Example:

	   * Network:
	   - Bandwidth
	   - Jitter
	   - Latency
	   - Packet_Loss

	   * CPU
	   - Limit

	   * RAM
	   - Limit

	*/
}

// Metrics are ways to gather information during process execution
type Metric interface {
	Core
	// Events returns reader for events
	Events() (io.ReadCloser, error)
	// StderrReader returns reader of stderr for the node
	StderrReader() (io.ReadCloser, error)
	// StdoutReader returns reader of stdout for the node
	StdoutReader() (io.ReadCloser, error)

	// Heartbeat returns key values pairs of a defined set of metrics
	Heartbeat() (map[string]string, error)
	// Metric returns metric value at key
	Metric(key string) (string, error)
	// GetMetricList returns list of metrics
	GetMetricList() []string
	// GetMetricDesc returns description of metrics
	GetMetricDesc(key string) (string, error)
	/* Examples:

	   * Filesystem:
	   - device_name
	   - swap
	   - mount_point
	   - total
	   - pct_used

	   * CPU
	   - cores
	   - iowait
	   - pct_used

	   * RAM
	   - total
	   - pct_used

	   * Network
	   - bwout
	   - bwin
	   - ping

	*/
}

// Core specifies the interface to a process controlled by iptb
type Core interface {
	Libp2p
	// Allows a node to run any initialization it may require
	// Ex: Installing additional dependencies / setting up configuration
	Init(ctx context.Context, args ...string) (Output, error)
	// Starts the node, wait can be used to delay the return till the node is ready
	// to accept commands
	Start(ctx context.Context, wait bool, args ...string) (Output, error)
	// Stops the node
	Stop(ctx context.Context) error
	// Runs a command in the context of the node
	RunCmd(ctx context.Context, stdin io.Reader, args ...string) (Output, error)
	// Connect the node to another
	Connect(ctx context.Context, n Core) error
	// Starts a shell in the context of the node
	Shell(ctx context.Context, ns []Core) error

	// Dir returns the iptb directory assigned to the node
	Dir() string
	// Type returns a string that identifies the implementation
	// Examples localipfs, dockeripfs, etc.
	Type() string

	String() string
}
