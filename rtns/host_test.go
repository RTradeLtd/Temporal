package rtns_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	ci "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	peer "gx/ipfs/QmTRhk7cgjUf2gfQ3p2M9KPECNZEW9XUrmHcFCgog4cPgB/go-libp2p-peer"

	"github.com/RTradeLtd/Temporal/rtns"
)

type contextKey string

const (
	ipnsPublishTTL contextKey = "ipns-publish-ttl"
	testPath                  = "/ipfs/QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
	testSwarmADDR             = "/ip4/0.0.0.0/tcp/4002"
)

func TestPublisher_Success(t *testing.T) {
	// create our private key
	pk, _, err := ci.GenerateKeyPair(ci.Ed25519, 256)
	if err != nil {
		t.Fatal(err)
	}
	publisher, err := rtns.NewPublisher(pk, false, testSwarmADDR)
	if err != nil {
		t.Fatal(err)
	}
	// sleep giving time for our node to discover some peers
	time.Sleep(time.Second * 15)
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
	// set the lifetime
	eol := time.Now().Add(time.Minute * 10)
	ctx := context.WithValue(context.Background(), ipnsPublishTTL, time.Minute*10)
	if err := publisher.PublishWithEOL(ctx, pk, testPath, eol); err != nil {
		t.Fatal(err)
	}
}

func TestPublisher_Failure(t *testing.T) {
	// create our private key
	pk, _, err := ci.GenerateKeyPair(ci.Ed25519, 256)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rtns.NewPublisher(pk, false, "notarealaddress"); err == nil {
		t.Fatal("expected error")
	}
}
