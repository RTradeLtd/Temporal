package routinghelpers

import (
	"context"
	"testing"

	peert "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer/test"
	routing "gx/ipfs/QmS4niovD1U6pRjUBXivr1zvvLBqiTKbERjFo994JU7oQS/go-libp2p-routing"
)

func TestGetPublicKey(t *testing.T) {
	d := Parallel{
		Parallel{
			&Compose{
				ValueStore: &LimitedValueStore{
					ValueStore: new(dummyValueStore),
					Namespaces: []string{"other"},
				},
			},
		},
		Tiered{
			&Compose{
				ValueStore: &LimitedValueStore{
					ValueStore: new(dummyValueStore),
					Namespaces: []string{"pk"},
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
	}

	pid, _ := peert.RandPeerID()

	ctx := context.Background()
	if _, err := d.GetPublicKey(ctx, pid); err != routing.ErrNotFound {
		t.Fatal(err)
	}
}
