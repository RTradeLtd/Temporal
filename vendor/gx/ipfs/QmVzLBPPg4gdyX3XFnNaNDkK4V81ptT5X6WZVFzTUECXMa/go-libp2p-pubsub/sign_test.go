package pubsub

import (
	"testing"

	pb "gx/ipfs/QmVzLBPPg4gdyX3XFnNaNDkK4V81ptT5X6WZVFzTUECXMa/go-libp2p-pubsub/pb"

	crypto "gx/ipfs/QmTW4SdgBWq9GjsBsHeUx8WuGxzhgzAf88UMH2w62PC8yK/go-libp2p-crypto"
	peer "gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer"
)

func TestSigning(t *testing.T) {
	privk, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		t.Fatal(err)
	}
	testSignVerify(t, privk)

	privk, _, err = crypto.GenerateKeyPair(crypto.Ed25519, 0)
	if err != nil {
		t.Fatal(err)
	}
	testSignVerify(t, privk)
}

func testSignVerify(t *testing.T, privk crypto.PrivKey) {
	id, err := peer.IDFromPublicKey(privk.GetPublic())
	if err != nil {
		t.Fatal(err)
	}
	m := pb.Message{
		Data:     []byte("abc"),
		TopicIDs: []string{"foo"},
		From:     []byte(id),
		Seqno:    []byte("123"),
	}
	signMessage(id, privk, &m)
	err = verifyMessageSignature(&m)
	if err != nil {
		t.Fatal(err)
	}
}
