package loggables

import (
	"errors"
	ic "gx/ipfs/QmTW4SdgBWq9GjsBsHeUx8WuGxzhgzAf88UMH2w62PC8yK/go-libp2p-crypto"
	peer "gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer"

	"crypto/rand"
	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	"net"
	"testing"
)

func TestNetConn(t *testing.T) {
	connA, _ := net.Pipe()
	loggable := NetConn(connA)
	if _, ok := loggable.Loggable()["localAddr"]; !ok {
		t.Fatalf("loggable missing localAddr")
	}
	if _, ok := loggable.Loggable()["remoteAddr"]; !ok {
		t.Fatalf("loggable missing remoteAddr")
	}
}

func TestError(t *testing.T) {
	loggable := Error(errors.New("test"))
	if _, ok := loggable.Loggable()["error"]; !ok {
		t.Fatalf("loggable missing error")
	}
}

func TestUuid(t *testing.T) {
	loggable := Uuid("testKey")
	if _, ok := loggable.Loggable()["testKey"]; !ok {
		t.Fatalf("loggable does not contain expected key")
	}
}

func TestDial(t *testing.T) {
	leftPriv, _, _ := ic.GenerateECDSAKeyPair(rand.Reader)
	lid, err := peer.IDFromPrivateKey(leftPriv)
	if err != nil {
		t.Fatalf("failed to create leftId: %s", err.Error())
	}
	rightPriv, _, _ := ic.GenerateECDSAKeyPair(rand.Reader)
	rid, err := peer.IDFromPrivateKey(rightPriv)
	if err != nil {
		t.Fatalf("failed to create rightId: %s", err.Error())
	}
	laddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/80")
	if err != nil {
		t.Fatalf("failed to create laddr: %s", err.Error())
	}
	raddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/81")
	if err != nil {
		t.Fatalf("failed to create raddr: %s", err.Error())
	}

	metadata := Dial("test", lid, rid, laddr, raddr)
	loggable := metadata.Loggable()

	expected := [][]string{
		{"localPeer", lid.Pretty()},
		{"localAddr", laddr.String()},
		{"remotePeer", rid.Pretty()},
		{"remoteAddr", raddr.String()},
	}
	for _, tuple := range expected {
		if actual, ok := loggable[tuple[0]]; !ok || actual != tuple[1] {
			t.Fatalf("Expected %s but got %s", tuple[1], actual)
		}
	}

}
