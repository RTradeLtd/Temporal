package config

import (
	"context"
	"testing"

	swarmt "gx/ipfs/QmPWNZRUybw3nwJH3mpkrwB97YEQmXRkzvyh34rpJiih6Q/go-libp2p-swarm/testing"
	bhost "gx/ipfs/QmQiaskfWpdRJ4x2spEQjPFTUkEB87KDYu91qnNYBqvvcX/go-libp2p/p2p/host/basic"
	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	host "gx/ipfs/QmRRCrNRs4qxotXx7WJT6SpCvSNEhXvyBcVjXY2K71pcjE/go-libp2p-host"
	mux "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
	yamux "gx/ipfs/QmcsgrV3nCAKjiHKZhKVXWc4oY3WBECJCqahXEMpHeMrev/go-smux-yamux"
)

func TestMuxerSimple(t *testing.T) {
	// single
	_, err := MuxerConstructor(func(_ peer.ID) mux.Transport { return nil })
	if err != nil {
		t.Fatal(err)
	}
}

func TestMuxerByValue(t *testing.T) {
	_, err := MuxerConstructor(yamux.DefaultTransport)
	if err != nil {
		t.Fatal(err)
	}
}
func TestMuxerDuplicate(t *testing.T) {
	_, err := MuxerConstructor(func(_ peer.ID, _ peer.ID) mux.Transport { return nil })
	if err != nil {
		t.Fatal(err)
	}
}

func TestMuxerError(t *testing.T) {
	_, err := MuxerConstructor(func() (mux.Transport, error) { return nil, nil })
	if err != nil {
		t.Fatal(err)
	}
}

func TestMuxerBadTypes(t *testing.T) {
	for i, f := range []interface{}{
		func() error { return nil },
		func() string { return "" },
		func() {},
		func(string) mux.Transport { return nil },
		func(string) (mux.Transport, error) { return nil, nil },
		nil,
		"testing",
	} {

		if _, err := MuxerConstructor(f); err == nil {
			t.Fatalf("constructor %d with type %T should have failed", i, f)
		}
	}
}

func TestCatchDuplicateTransportsMuxer(t *testing.T) {
	ctx := context.Background()
	h := bhost.New(swarmt.GenSwarm(t, ctx))
	yamuxMuxer, err := MuxerConstructor(yamux.DefaultTransport)
	if err != nil {
		t.Fatal(err)
	}

	var tests = map[string]struct {
		h             host.Host
		transports    []MsMuxC
		expectedError string
	}{
		"no duplicate transports": {
			h:             h,
			transports:    []MsMuxC{MsMuxC{yamuxMuxer, "yamux"}},
			expectedError: "",
		},
		"duplicate transports": {
			h: h,
			transports: []MsMuxC{
				MsMuxC{yamuxMuxer, "yamux"},
				MsMuxC{yamuxMuxer, "yamux"},
			},
			expectedError: "duplicate muxer transport: yamux",
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			_, err = makeMuxer(test.h, test.transports)
			if err != nil {
				if err.Error() != test.expectedError {
					t.Errorf(
						"\nexpected: [%v]\nactual:   [%v]\n",
						test.expectedError,
						err,
					)
				}
			}
		})
	}
}
