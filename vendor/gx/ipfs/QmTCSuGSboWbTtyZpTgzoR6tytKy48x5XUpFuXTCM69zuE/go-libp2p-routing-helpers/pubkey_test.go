package routinghelpers

import (
	"context"
	"testing"

	peert "github.com/libp2p/go-libp2p-peer/test"
	routing "github.com/libp2p/go-libp2p-routing"
)

func TestGetPublicKey(t *testing.T) {
	d := Parallel{
		Routers: []routing.IpfsRouting{
			Parallel{
				Routers: []routing.IpfsRouting{
					&Compose{
						ValueStore: &LimitedValueStore{
							ValueStore: new(dummyValueStore),
							Namespaces: []string{"other"},
						},
					},
				},
			},
			Tiered{
				Routers: []routing.IpfsRouting{
					&Compose{
						ValueStore: &LimitedValueStore{
							ValueStore: new(dummyValueStore),
							Namespaces: []string{"pk"},
						},
					},
				},
			},
			&Compose{
				ValueStore: &LimitedValueStore{
					ValueStore: new(dummyValueStore),
					Namespaces: []string{"other", "pk"},
				},
			},
			&struct{ Compose }{Compose{ValueStore: &LimitedValueStore{ValueStore: Null{}}}},
			&struct{ Compose }{},
		},
	}

	pid, _ := peert.RandPeerID()

	ctx := context.Background()
	if _, err := d.GetPublicKey(ctx, pid); err != routing.ErrNotFound {
		t.Fatal(err)
	}
}
