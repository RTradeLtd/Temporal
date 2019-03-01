package p2pclient

import (
	"errors"
	"sync"

	pb "gx/ipfs/QmRpsHkKwSXnbFRaQUhAY27WtnUAC2m8tAPfws9Lk72T4h/go-libp2p-daemon/pb"
	multiaddr "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	peer "gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer"
	logging "gx/ipfs/QmbkT7eMTyXfpeyB3ZMxxcxg7XH8t6uXp49jqzz4HB7BGF/go-log"
	manet "gx/ipfs/Qmc85NSvmSG4Frn9Vb2cBc1rMyULH6D3TNVEfCzSKoUpip/go-multiaddr-net"
	ggio "gx/ipfs/QmddjPSGZb3ieihSseFeCfVRpZzcqczPNsD2DvarSwnjJB/gogo-protobuf/io"
)

var log = logging.Logger("p2pclient")

// MessageSizeMax is cribbed from github.com/libp2p/go-libp2p-net
const MessageSizeMax = 1 << 22 // 4 MB

// Client is the struct that manages a connection to a libp2p daemon.
type Client struct {
	controlMaddr multiaddr.Multiaddr
	listenMaddr  multiaddr.Multiaddr
	listener     manet.Listener

	mhandlers sync.Mutex
	handlers  map[string]StreamHandlerFunc
}

// NewClient creates a new libp2p daemon client, connecting to a daemon
// listening on a multi-addr at controlMaddr, and establishing an inbound
// listening multi-address at listenMaddr
func NewClient(controlMaddr, listenMaddr multiaddr.Multiaddr) (*Client, error) {
	client := &Client{
		controlMaddr: controlMaddr,
		handlers:     make(map[string]StreamHandlerFunc),
	}

	if err := client.listen(listenMaddr); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) newControlConn() (manet.Conn, error) {
	return manet.Dial(c.controlMaddr)
}

// Identify queries the daemon for its peer ID and listen addresses.
func (c *Client) Identify() (peer.ID, []multiaddr.Multiaddr, error) {
	control, err := c.newControlConn()
	if err != nil {
		return peer.ID(""), nil, err
	}
	defer control.Close()
	r := ggio.NewDelimitedReader(control, MessageSizeMax)
	w := ggio.NewDelimitedWriter(control)

	req := &pb.Request{Type: pb.Request_IDENTIFY.Enum()}
	if err = w.WriteMsg(req); err != nil {
		return peer.ID(""), nil, err
	}

	res := &pb.Response{}
	if err = r.ReadMsg(res); err != nil {
		return peer.ID(""), nil, err
	}

	if reserr := res.GetError(); reserr != nil {
		return peer.ID(""), nil, errors.New(reserr.GetMsg())
	}

	idres := res.GetIdentify()
	id, err := peer.IDFromBytes(idres.Id)
	if err != nil {
		return peer.ID(""), nil, err
	}
	addrs := make([]multiaddr.Multiaddr, 0, len(idres.Addrs))
	for i, addrbytes := range idres.Addrs {
		addr, err := multiaddr.NewMultiaddrBytes(addrbytes)
		if err != nil {
			log.Errorf("failed to parse multiaddr in position %d in response to identify request", i)
			continue
		}
		addrs = append(addrs, addr)
	}

	return id, addrs, nil
}

// Connect establishes a connection to a peer after populating the Peerstore
// entry for said peer with a list of addresses.
func (c *Client) Connect(p peer.ID, addrs []multiaddr.Multiaddr) error {
	control, err := c.newControlConn()
	if err != nil {
		return err
	}
	defer control.Close()
	r := ggio.NewDelimitedReader(control, MessageSizeMax)
	w := ggio.NewDelimitedWriter(control)

	addrbytes := make([][]byte, len(addrs))
	for i, addr := range addrs {
		addrbytes[i] = addr.Bytes()
	}

	req := &pb.Request{
		Type: pb.Request_CONNECT.Enum(),
		Connect: &pb.ConnectRequest{
			Peer:  []byte(p),
			Addrs: addrbytes,
		},
	}

	if err := w.WriteMsg(req); err != nil {
		return err
	}

	res := &pb.Response{}
	if err := r.ReadMsg(res); err != nil {
		return err
	}

	if err := res.GetError(); err != nil {
		return errors.New(err.GetMsg())
	}

	return nil
}
