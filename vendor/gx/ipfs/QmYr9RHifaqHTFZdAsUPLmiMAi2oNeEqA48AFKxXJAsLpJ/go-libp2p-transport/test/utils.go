package utils

import (
	"reflect"
	"runtime"
	"testing"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	tpt "gx/ipfs/QmYr9RHifaqHTFZdAsUPLmiMAi2oNeEqA48AFKxXJAsLpJ/go-libp2p-transport"
)

var Subtests = []func(t *testing.T, ta, tb tpt.Transport, maddr ma.Multiaddr, peerA peer.ID){
	SubtestProtocols,
	SubtestBasic,
	SubtestCancel,
	SubtestPingPong,

	// Stolen from the stream muxer test suite.
	SubtestStress1Conn1Stream1Msg,
	SubtestStress1Conn1Stream100Msg,
	SubtestStress1Conn100Stream100Msg,
	SubtestStress50Conn10Stream50Msg,
	SubtestStress1Conn1000Stream10Msg,
	SubtestStress1Conn100Stream100Msg10MB,
	SubtestStreamOpenStress,
	SubtestStreamReset,
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func SubtestTransport(t *testing.T, ta, tb tpt.Transport, addr string, peerA peer.ID) {
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range Subtests {
		t.Run(getFunctionName(f), func(t *testing.T) {
			f(t, ta, tb, maddr, peerA)
		})
	}
}
