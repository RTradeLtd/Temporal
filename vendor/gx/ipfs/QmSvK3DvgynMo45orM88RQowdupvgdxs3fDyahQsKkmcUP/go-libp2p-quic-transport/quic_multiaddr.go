package libp2pquic

import (
	"net"

	manet "gx/ipfs/QmQVUtnrNGtCRkCMpXgpApfzQjc8FDaDVxHqWH8cnZQeh5/go-multiaddr-net"
	ma "gx/ipfs/QmRKLtwMw131aK7ugC3G7ybpumMz78YrJe5dzneyindvG1/go-multiaddr"
)

var quicMA ma.Multiaddr

func init() {
	var err error
	quicMA, err = ma.NewMultiaddr("/quic")
	if err != nil {
		panic(err)
	}
}

func toQuicMultiaddr(na net.Addr) (ma.Multiaddr, error) {
	udpMA, err := manet.FromNetAddr(na)
	if err != nil {
		return nil, err
	}
	return udpMA.Encapsulate(quicMA), nil
}

func fromQuicMultiaddr(addr ma.Multiaddr) (net.Addr, error) {
	return manet.ToNetAddr(addr.Decapsulate(quicMA))
}
