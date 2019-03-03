package p2pd

import (
	"context"
	"errors"
	"math/rand"
	"time"

	inet "gx/ipfs/QmY3ArotKMKaL7YGfbQfyDrib6RVraLqZYWXZvVgZktBxp/go-libp2p-net"
	pstore "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore"
	dht "gx/ipfs/QmdR6WN3TUEAVQ9KWE2UiFJikWTbUvgBJay6mjB4yUJebq/go-libp2p-kad-dht"
)

var BootstrapPeers = dht.DefaultBootstrapPeers

const BootstrapConnections = 4

func bootstrapPeerInfo() ([]*pstore.PeerInfo, error) {
	pis := make([]*pstore.PeerInfo, 0, len(BootstrapPeers))
	for _, a := range BootstrapPeers {
		pi, err := pstore.InfoFromP2pAddr(a)
		if err != nil {
			return nil, err
		}
		pis = append(pis, pi)
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
