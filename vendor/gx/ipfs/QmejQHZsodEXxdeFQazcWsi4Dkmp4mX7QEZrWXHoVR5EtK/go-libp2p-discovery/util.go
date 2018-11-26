package discovery

import (
	"context"
	"time"

	pstore "gx/ipfs/QmQAGG1zxfePqj2t7bLxyN8AFccZ889DDR9Gn8kVLDrGZo/go-libp2p-peerstore"
	logging "gx/ipfs/QmcuXC5cxs79ro2cUuHs4HQ2bkDLJUYokwL8aivcX6HW3C/go-log"
)

var log = logging.Logger("discovery")

// FindPeers is a utility function that synchonously collects peers from a Discoverer
func FindPeers(ctx context.Context, d Discoverer, ns string, limit int) ([]pstore.PeerInfo, error) {
	res := make([]pstore.PeerInfo, 0, limit)

	ch, err := d.FindPeers(ctx, ns, Limit(limit))
	if err != nil {
		return nil, err
	}

	for pi := range ch {
		res = append(res, pi)
	}

	return res, nil
}

// Advertise is a utility function that persistently advertises a service through an Advertiser
func Advertise(ctx context.Context, a Advertiser, ns string) {
	go func() {
		for {
			ttl, err := a.Advertise(ctx, ns)
			if err != nil {
				log.Debugf("Error advertising %s: %s", ns, err.Error())
				if ctx.Err() != nil {
					return
				}

				select {
				case <-time.After(2 * time.Minute):
					continue
				case <-ctx.Done():
					return
				}
			}

			wait := 7 * ttl / 8
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return
			}
		}
	}()
}
