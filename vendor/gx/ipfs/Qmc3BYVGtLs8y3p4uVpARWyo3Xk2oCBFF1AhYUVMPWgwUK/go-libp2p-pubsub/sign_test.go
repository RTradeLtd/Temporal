package pubsub

import (
	"testing"

	pb "gx/ipfs/Qmc3BYVGtLs8y3p4uVpARWyo3Xk2oCBFF1AhYUVMPWgwUK/go-libp2p-pubsub/pb"

	crypto "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	peer "gx/ipfs/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"
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
