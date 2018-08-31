package routinghelpers

import (
	"context"
	"errors"
	"strings"
	"sync"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	routing "gx/ipfs/QmS4niovD1U6pRjUBXivr1zvvLBqiTKbERjFo994JU7oQS/go-libp2p-routing"
	ropts "gx/ipfs/QmS4niovD1U6pRjUBXivr1zvvLBqiTKbERjFo994JU7oQS/go-libp2p-routing/options"
	cid "gx/ipfs/QmZFbDTY9jfSBms2MchvYM9oYRbAF19K7Pby47yDBfpPrb/go-cid"
	pstore "gx/ipfs/QmeKD8YT7887Xu6Z86iZmpYNxrLogJexqxEugSmaf14k64/go-libp2p-peerstore"
)

type dummyValueStore sync.Map

func (d *dummyValueStore) PutValue(ctx context.Context, key string, value []byte, opts ...ropts.Option) error {
	if strings.HasPrefix(key, "/notsupported/") {
		return routing.ErrNotSupported
	}
	if strings.HasPrefix(key, "/error/") {
		return errors.New(key[len("/error/"):])
	}
	if strings.HasPrefix(key, "/stall/") {
		<-ctx.Done()
		return ctx.Err()
	}
	(*sync.Map)(d).Store(key, value)
	return nil
}

func (d *dummyValueStore) GetValue(ctx context.Context, key string, opts ...ropts.Option) ([]byte, error) {
	if strings.HasPrefix(key, "/error/") {
		return nil, errors.New(key[len("/error/"):])
	}
	if strings.HasPrefix(key, "/stall/") {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if v, ok := (*sync.Map)(d).Load(key); ok {
		return v.([]byte), nil
	}
	return nil, routing.ErrNotFound
}

type dummyProvider map[string][]peer.ID

func (d dummyProvider) FindProvidersAsync(ctx context.Context, c *cid.Cid, count int) <-chan pstore.PeerInfo {
	peers := d[c.KeyString()]
	if len(peers) > count {
		peers = peers[:count]
	}
	out := make(chan pstore.PeerInfo)
	go func() {
		defer close(out)
		for _, p := range peers {
			if p == "stall" {
				<-ctx.Done()
				return
			}
			select {
			case out <- pstore.PeerInfo{ID: p}:
			case <-ctx.Done():
			}
		}
	}()
	return out
}

func (d dummyProvider) Provide(ctx context.Context, c *cid.Cid, local bool) error {
	return routing.ErrNotSupported
}

type cbProvider func(c *cid.Cid, local bool) error

func (d cbProvider) Provide(ctx context.Context, c *cid.Cid, local bool) error {
	return d(c, local)
}

func (d cbProvider) FindProvidersAsync(ctx context.Context, c *cid.Cid, count int) <-chan pstore.PeerInfo {
	ch := make(chan pstore.PeerInfo)
	close(ch)
	return ch
}

type dummyPeerRouter map[peer.ID]struct{}

func (d dummyPeerRouter) FindPeer(ctx context.Context, p peer.ID) (pstore.PeerInfo, error) {
	if _, ok := d[p]; ok {
		return pstore.PeerInfo{ID: p}, nil
	}
	return pstore.PeerInfo{}, routing.ErrNotFound
}
