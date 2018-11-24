package relay

import (
	"net"

	pb "gx/ipfs/QmWxCQ66YzAb24Py5D3qeDJcVTF4aW9AeAmppQ7FhQZxgZ/go-libp2p-circuit/pb"

	manet "gx/ipfs/QmQVUtnrNGtCRkCMpXgpApfzQjc8FDaDVxHqWH8cnZQeh5/go-multiaddr-net"
	ma "gx/ipfs/QmRKLtwMw131aK7ugC3G7ybpumMz78YrJe5dzneyindvG1/go-multiaddr"
)

var _ manet.Listener = (*RelayListener)(nil)

type RelayListener Relay

func (l *RelayListener) Relay() *Relay {
	return (*Relay)(l)
}

func (r *Relay) Listener() *RelayListener {
	// TODO: Only allow one!
	return (*RelayListener)(r)
}

func (l *RelayListener) Accept() (manet.Conn, error) {
	select {
	case c := <-l.incoming:
		err := l.Relay().writeResponse(c.Stream, pb.CircuitRelay_SUCCESS)
		if err != nil {
			log.Debugf("error writing relay response: %s", err.Error())
			c.Stream.Reset()
			return nil, err
		}

		// TODO: Pretty print.
		log.Infof("accepted relay connection: %s", c)

		return c, nil
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	}
}

func (l *RelayListener) Addr() net.Addr {
	return &NetAddr{
		Relay:  "any",
		Remote: "any",
	}
}

func (l *RelayListener) Multiaddr() ma.Multiaddr {
	return ma.Cast(ma.CodeToVarint(P_CIRCUIT))
}

func (l *RelayListener) Close() error {
	// TODO: noop?
	return nil
}
