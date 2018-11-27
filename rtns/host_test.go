package rtns_test

import (
	"context"
	"fmt"
	"testing"

	ci "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	peer "gx/ipfs/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"

	"github.com/RTradeLtd/Temporal/rtns"
)

const (
	testPath      = "/ipfs/QmNdm1ZyLX7hBVTDYhfiZ6oVjQHdEkN1VxV5rfJDHBVZyH"
	testSwarmADDR = "/ip4/0.0.0.0/tcp/4002"
)

func TestPublisher_Gen(t *testing.T) {
	publisher, err := rtns.NewPublisher(nil, testSwarmADDR)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := publisher.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	// create our private key
	pk, _, err := ci.GenerateKeyPair(ci.Ed25519, 256)
	if err != nil {
		t.Fatal(err)
	}
	if pid, err := peer.IDFromPrivateKey(pk); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println("id to check ", pid.Pretty())
	}
	if err := publisher.Publish(context.Background(), pk, testPath); err != nil {
		t.Fatal(err)
	}
}

func TestPublisher_NoGen(t *testing.T) {
	// create our private key
	pk, _, err := ci.GenerateKeyPair(ci.Ed25519, 256)
	if err != nil {
		t.Fatal(err)
	}
	publisher, err := rtns.NewPublisher(&rtns.Opts{PK: pk}, testSwarmADDR)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := publisher.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	// create our private key
	pk, _, err = ci.GenerateKeyPair(ci.Ed25519, 256)
	if err != nil {
		t.Fatal(err)
	}
	if pid, err := peer.IDFromPrivateKey(pk); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println("id to check ", pid.Pretty())
	}
	if err := publisher.Publish(context.Background(), pk, testPath); err != nil {
		t.Fatal(err)
	}
}
