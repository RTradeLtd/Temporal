package p2pd

import (
	"context"
	"errors"
	"math/rand"
	"time"

	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	inet "gx/ipfs/QmNgLg1NTw37iWbYPKcyK85YJ9Whs1MkPtJwhfqbNYAyKg/go-libp2p-net"
	pstore "gx/ipfs/QmPiemjiKBC9VA7vZF82m4x1oygtg2c2YVqag8PX7dN1BD/go-libp2p-peerstore"
)

var BootstrapPeers = []string{
	"/dnsaddr/bootstrap.libp2p.io/ipfs/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
	"/dnsaddr/bootstrap.libp2p.io/ipfs/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
	"/dnsaddr/bootstrap.libp2p.io/ipfs/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
	"/dnsaddr/bootstrap.libp2p.io/ipfs/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",            // mars.i.ipfs.io
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",           // pluto.i.ipfs.io
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",           // saturn.i.ipfs.io
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",             // venus.i.ipfs.io
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",            // earth.i.ipfs.io
	"/ip6/2604:a880:1:20::203:d001/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",  // pluto.i.ipfs.io
	"/ip6/2400:6180:0:d0::151:6001/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",  // saturn.i.ipfs.io
	"/ip6/2604:a880:800:10::4a:5001/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64", // venus.i.ipfs.io
	"/ip6/2a03:b0c0:0:1010::23:1001/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd", // earth.i.ipfs.io
}

const BootstrapConnections = 4

func bootstrapPeerInfo() ([]*pstore.PeerInfo, error) {
	pis := make([]*pstore.PeerInfo, len(BootstrapPeers))
	for x, p := range BootstrapPeers {
		a, err := ma.NewMultiaddr(p)
		if err != nil {
			return nil, err
		}

		pi, err := pstore.InfoFromP2pAddr(a)
		if err != nil {
			return nil, err
		}

		pis[x] = pi
	}

	return pis, nil
}

func shufflePeerInfos(peers []*pstore.PeerInfo) {
	for i := range peers {
		j := rand.Intn(i + 1)
		peers[i], peers[j] = peers[j], peers[i]
	}
}

func (d *Daemon) Bootstrap() error {
	pis, err := bootstrapPeerInfo()
	if err != nil {
		return err
	}

	for _, pi := range pis {
		d.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, pstore.PermanentAddrTTL)
	}

	count := d.connectBootstrapPeers(pis, BootstrapConnections)
	if count == 0 {
		return errors.New("Failed to connect to bootstrap peers")
	}

	go d.keepBootstrapConnections(pis)

	if d.dht != nil {
		return d.dht.Bootstrap(d.ctx)
	}

	return nil
}

func (d *Daemon) connectBootstrapPeers(pis []*pstore.PeerInfo, toconnect int) int {
	count := 0

	shufflePeerInfos(pis)

	ctx, cancel := context.WithTimeout(d.ctx, 60*time.Second)
	defer cancel()

	for _, pi := range pis {
		if d.host.Network().Connectedness(pi.ID) == inet.Connected {
			continue
		}
		err := d.host.Connect(ctx, *pi)
		if err != nil {
			log.Debugf("Error connecting to bootstrap peer %s: %s", pi.ID, err.Error())
		} else {
			d.host.ConnManager().TagPeer(pi.ID, "bootstrap", 1)
			count++
			toconnect--
		}
		if toconnect == 0 {
			break
		}
	}

	return count

}

func (d *Daemon) keepBootstrapConnections(pis []*pstore.PeerInfo) {
	ticker := time.NewTicker(15 * time.Minute)
	for {
		<-ticker.C

		conns := d.host.Network().Conns()
		if len(conns) >= BootstrapConnections {
			continue
		}

		toconnect := BootstrapConnections - len(conns)
		d.connectBootstrapPeers(pis, toconnect)
	}
}
