package autonat

import (
	"context"
	"fmt"

	pb "gx/ipfs/QmQXeTmQcPFnf3ZAvik2qgKxWNoZ27aKGcW8hUBwBrTxT1/go-libp2p-autonat/pb"

	inet "gx/ipfs/QmPtFaR7BWHLAjSwLh9kXcyrgTzDpuhcWLkx8ioa9RMYnx/go-libp2p-net"
	ma "gx/ipfs/QmRKLtwMw131aK7ugC3G7ybpumMz78YrJe5dzneyindvG1/go-multiaddr"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	pstore "gx/ipfs/QmZ9zH2FnLcxv1xyzFeUpDUeo55xEhZQHgveZijcxr7TLj/go-libp2p-peerstore"
	ggio "gx/ipfs/QmdxUuburamoF6zF9qjeQC4WYcWGbWuRmdLacMEsW8ioD8/gogo-protobuf/io"
	host "gx/ipfs/QmfD51tKgJiTMnW9JEiDiPwsCY4mqUoxkhKhBfyW12spTC/go-libp2p-host"
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
