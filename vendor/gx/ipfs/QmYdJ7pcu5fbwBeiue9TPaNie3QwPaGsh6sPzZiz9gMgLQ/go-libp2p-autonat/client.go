package autonat

import (
	"context"
	"fmt"

	pb "gx/ipfs/QmYdJ7pcu5fbwBeiue9TPaNie3QwPaGsh6sPzZiz9gMgLQ/go-libp2p-autonat/pb"

	pstore "gx/ipfs/QmQAGG1zxfePqj2t7bLxyN8AFccZ889DDR9Gn8kVLDrGZo/go-libp2p-peerstore"
	ma "gx/ipfs/QmRKLtwMw131aK7ugC3G7ybpumMz78YrJe5dzneyindvG1/go-multiaddr"
	host "gx/ipfs/QmahxMNoNuSsgQefo9rkpcfRFmQrMN6Q99aztKXf63K7YJ/go-libp2p-host"
	peer "gx/ipfs/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"
	ggio "gx/ipfs/QmdxUuburamoF6zF9qjeQC4WYcWGbWuRmdLacMEsW8ioD8/gogo-protobuf/io"
	inet "gx/ipfs/QmenvQQy4bFGSiHJUGupVmCRHfetg5rH3vTp9Z2f6v2KXR/go-libp2p-net"
)

// AutoNATClient is a stateless client interface to AutoNAT peers
type AutoNATClient interface {
	// DialBack requests from a peer providing AutoNAT services to test dial back
	// and report the address on a successful connection.
	DialBack(ctx context.Context, p peer.ID) (ma.Multiaddr, error)
}

// AutoNATError is the class of errors signalled by AutoNAT services
type AutoNATError struct {
	Status pb.Message_ResponseStatus
	Text   string
}

// GetAddrs is a function that returns the addresses to dial back
type GetAddrs func() []ma.Multiaddr

// NewAutoNATClient creates a fresh instance of an AutoNATClient
// If getAddrs is nil, h.Addrs will be used
func NewAutoNATClient(h host.Host, getAddrs GetAddrs) AutoNATClient {
	if getAddrs == nil {
		getAddrs = h.Addrs
	}
	return &client{h: h, getAddrs: getAddrs}
}

type client struct {
	h        host.Host
	getAddrs GetAddrs
}

func (c *client) DialBack(ctx context.Context, p peer.ID) (ma.Multiaddr, error) {
	s, err := c.h.NewStream(ctx, p, AutoNATProto)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	r := ggio.NewDelimitedReader(s, inet.MessageSizeMax)
	w := ggio.NewDelimitedWriter(s)

	req := newDialMessage(pstore.PeerInfo{ID: c.h.ID(), Addrs: c.getAddrs()})
	err = w.WriteMsg(req)
	if err != nil {
		return nil, err
	}

	var res pb.Message
	err = r.ReadMsg(&res)
	if err != nil {
		return nil, err
	}

	if res.GetType() != pb.Message_DIAL_RESPONSE {
		return nil, fmt.Errorf("Unexpected response: %s", res.GetType().String())
	}

	status := res.GetDialResponse().GetStatus()
	switch status {
	case pb.Message_OK:
		addr := res.GetDialResponse().GetAddr()
		return ma.NewMultiaddrBytes(addr)

	default:
		return nil, AutoNATError{Status: status, Text: res.GetDialResponse().GetStatusText()}
	}
}

func (e AutoNATError) Error() string {
	return fmt.Sprintf("AutoNAT error: %s (%s)", e.Text, e.Status.String())
}

func (e AutoNATError) IsDialError() bool {
	return e.Status == pb.Message_E_DIAL_ERROR
}

func (e AutoNATError) IsDialRefused() bool {
	return e.Status == pb.Message_E_DIAL_REFUSED
}

// IsDialError returns true if the AutoNAT peer signalled an error dialing back
func IsDialError(e error) bool {
	ae, ok := e.(AutoNATError)
	return ok && ae.IsDialError()
}

// IsDialRefused returns true if the AutoNAT peer signalled refusal to dial back
func IsDialRefused(e error) bool {
	ae, ok := e.(AutoNATError)
	return ok && ae.IsDialRefused()
}
