package libp2pquic

import (
	"net"

	manet "gx/ipfs/QmV6FjemM1K8oXjrvuq3wuVWWoU2TLDPmNnKrxHzY3v6Ai/go-multiaddr-net"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
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
