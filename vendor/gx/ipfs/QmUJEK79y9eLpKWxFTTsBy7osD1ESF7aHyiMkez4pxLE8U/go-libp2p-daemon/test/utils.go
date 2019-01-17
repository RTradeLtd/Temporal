package test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	crypto "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	p2pd "gx/ipfs/QmUJEK79y9eLpKWxFTTsBy7osD1ESF7aHyiMkez4pxLE8U/go-libp2p-daemon"
	"gx/ipfs/QmUJEK79y9eLpKWxFTTsBy7osD1ESF7aHyiMkez4pxLE8U/go-libp2p-daemon/p2pclient"
	pb "gx/ipfs/QmUJEK79y9eLpKWxFTTsBy7osD1ESF7aHyiMkez4pxLE8U/go-libp2p-daemon/pb"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	peertest "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer/test"
	mh "gx/ipfs/QmerPMzPk1mJVowm8KgmoknWa4yCYvvugMPsgWmDNUvDLW/go-multihash"
)

func createTempDir(t *testing.T) (string, string, func()) {
	root := os.TempDir()
	dir, err := ioutil.TempDir(root, "p2pd")
	if err != nil {
		t.Fatalf("creating temp dir: %s", err)
	}
	daemonPath := filepath.Join(dir, "daemon.sock")
	clientPath := filepath.Join(dir, "client.sock")
	closer := func() {
		os.RemoveAll(dir)
	}
	return daemonPath, clientPath, closer
}

func createDaemon(t *testing.T, daemonAddr ma.Multiaddr) (*p2pd.Daemon, func()) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	daemon, err := p2pd.NewDaemon(ctx, daemonAddr)
	if err != nil {
		t.Fatal(err)
	}
	return daemon, cancelCtx
}

func createClient(t *testing.T, daemonAddr ma.Multiaddr, clientAddr ma.Multiaddr) (*p2pclient.Client, func()) {
	client, err := p2pclient.NewClient(daemonAddr, clientAddr)
	if err != nil {
		t.Fatal(err)
	}
	closer := func() {
		client.Close()
	}
	return client, closer
}

func createDaemonClientPair(t *testing.T) (*p2pd.Daemon, *p2pclient.Client, func()) {
	daemonPath, clientPath, dirCloser := createTempDir(t)
	dmaddr, err := ma.NewComponent("unix", daemonPath)
	if err != nil {
		t.Fatal(err)
	}
	cmaddr, err := ma.NewComponent("unix", clientPath)
	if err != nil {
		t.Fatal(err)
	}
	daemon, closeDaemon := createDaemon(t, dmaddr)
	client, closeClient := createClient(t, dmaddr, cmaddr)

	closer := func() {
		closeDaemon()
		closeClient()
		dirCloser()
	}
	return daemon, client, closer
}

func createMockDaemonClientPair(t *testing.T) (*mockdaemon, *p2pclient.Client, func()) {
	daemonPath, clientPath, dirCloser := createTempDir(t)
	dmaddr, err := ma.NewComponent("unix", daemonPath)
	if err != nil {
		t.Fatal(err)
	}
	cmaddr, err := ma.NewComponent("unix", clientPath)
	if err != nil {
		t.Fatal(err)
	}
	client, clientCloser := createClient(t, dmaddr, cmaddr)
	daemon := newMockDaemon(t, dmaddr, cmaddr)
	closer := func() {
		daemon.Close()
		clientCloser()
		dirCloser()
	}
	return daemon, client, closer
}

func randPeerID(t *testing.T) peer.ID {
	id, err := peertest.RandPeerID()
	if err != nil {
		t.Fatalf("peer id: %s", err)
	}
	return id
}

func randPeerIDs(t *testing.T, n int) []peer.ID {
	ids := make([]peer.ID, n)
	for i := 0; i < n; i++ {
		ids[i] = randPeerID(t)
	}
	return ids
}

func randCid(t *testing.T) cid.Cid {
	buf := make([]byte, 10)
	rand.Read(buf)
	hash, err := mh.Sum(buf, mh.SHA2_256, -1)
	if err != nil {
		t.Fatalf("creating hash for cid: %s", err)
	}
	id := cid.NewCidV1(cid.Raw, hash)
	if err != nil {
		t.Fatalf("creating cid: %s", err)
	}
	return id
}

func randCids(t *testing.T, n int) []cid.Cid {
	ids := make([]cid.Cid, n)
	for i := 0; i < n; i++ {
		ids[i] = randCid(t)
	}
	return ids
}

func randBytes(t *testing.T) []byte {
	buf := make([]byte, 10)
	rand.Read(buf)
	return buf
}

func randString(t *testing.T) string {
	buf := make([]byte, 10)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}

func randStrings(t *testing.T, n int) []string {
	out := make([]string, n)
	for i := 0; i < n; i++ {
		buf := make([]byte, 10)
		rand.Read(buf)
		out[i] = hex.EncodeToString(buf)
	}
	return out
}

func randPubKey(t *testing.T) crypto.PubKey {
	_, pub, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatalf("generating pubkey: %s", err)
	}
	return pub
}

func wrapDhtResponse(dht *pb.DHTResponse) *pb.Response {
	return &pb.Response{
		Type: pb.Response_OK.Enum(),
		Dht:  dht,
	}
}

func peerInfoResponse(t *testing.T, id peer.ID) *pb.DHTResponse {
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/p2p-circuit/p2p/%s", id.Pretty()))
	if err != nil {
		t.Fatal(err)
	}
	return &pb.DHTResponse{
		Type: pb.DHTResponse_VALUE.Enum(),
		Peer: &pb.PeerInfo{
			Id:    []byte(id),
			Addrs: [][]byte{addr.Bytes()},
		},
	}
}

func peerIDResponse(t *testing.T, id peer.ID) *pb.DHTResponse {
	return &pb.DHTResponse{
		Type:  pb.DHTResponse_VALUE.Enum(),
		Value: []byte(id),
	}
}

func valueResponse(buf []byte) *pb.DHTResponse {
	return &pb.DHTResponse{
		Type:  pb.DHTResponse_VALUE.Enum(),
		Value: buf,
	}
}
