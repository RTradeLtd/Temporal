package test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gx/ipfs/QmPVkJMTeRC6iBByPWdrRkD3BE5UXsj5HPzb4kPqL186mS/testify/require"

	p2pd "gx/ipfs/QmRpsHkKwSXnbFRaQUhAY27WtnUAC2m8tAPfws9Lk72T4h/go-libp2p-daemon"
	"gx/ipfs/QmRpsHkKwSXnbFRaQUhAY27WtnUAC2m8tAPfws9Lk72T4h/go-libp2p-daemon/p2pclient"
	pb "gx/ipfs/QmRpsHkKwSXnbFRaQUhAY27WtnUAC2m8tAPfws9Lk72T4h/go-libp2p-daemon/pb"
	crypto "gx/ipfs/QmTW4SdgBWq9GjsBsHeUx8WuGxzhgzAf88UMH2w62PC8yK/go-libp2p-crypto"
	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	cid "gx/ipfs/QmTbxNB1NwDesLmKTscr4udL2tVP7MaxvXnD1D9yX7g3PN/go-cid"
	peer "gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer"
	peertest "gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer/test"
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
	daemon, err := p2pd.NewDaemon(ctx, daemonAddr, false, false)
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
	dmaddr, cmaddr, dirCloser := getEndpointsMaker(t)(t)
	daemon, closeDaemon := createDaemon(t, dmaddr)
	client, closeClient := createClient(t, daemon.Listener().Multiaddr(), cmaddr)

	closer := func() {
		closeDaemon()
		closeClient()
		dirCloser()
	}
	return daemon, client, closer
}

type makeEndpoints func(t *testing.T) (daemon, client ma.Multiaddr, cleanup func())

func makeTcpLocalhostEndpoints(t *testing.T) (daemon, client ma.Multiaddr, cleanup func()) {
	daemon, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	require.NoError(t, err)
	client, err = ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	require.NoError(t, err)
	cleanup = func() {}
	return
}

func makeUnixEndpoints(t *testing.T) (daemon, client ma.Multiaddr, cleanup func()) {
	daemonPath, clientPath, cleanup := createTempDir(t)
	daemon, err := ma.NewComponent("unix", daemonPath)
	require.NoError(t, err)
	client, err = ma.NewComponent("unix", clientPath)
	require.NoError(t, err)
	return
}

func getEndpointsMaker(t *testing.T) makeEndpoints {
	if runtime.GOOS == "windows" {
		return makeTcpLocalhostEndpoints
	} else {
		return makeUnixEndpoints
	}
}

func createMockDaemonClientPair(t *testing.T) (*mockDaemon, *p2pclient.Client, func()) {
	dmaddr, cmaddr, cleanup := getEndpointsMaker(t)(t)

	daemon := newMockDaemon(t, dmaddr, cmaddr)
	client, clientCloser := createClient(t, daemon.listener.Multiaddr(), cmaddr)
	return daemon, client, func() {
		daemon.Close()
		clientCloser()
		cleanup()
	}
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
