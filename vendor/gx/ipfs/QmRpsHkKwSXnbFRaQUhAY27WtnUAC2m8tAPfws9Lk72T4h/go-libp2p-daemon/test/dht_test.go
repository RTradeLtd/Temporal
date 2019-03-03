package test

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"gx/ipfs/QmRpsHkKwSXnbFRaQUhAY27WtnUAC2m8tAPfws9Lk72T4h/go-libp2p-daemon/p2pclient"
	pb "gx/ipfs/QmRpsHkKwSXnbFRaQUhAY27WtnUAC2m8tAPfws9Lk72T4h/go-libp2p-daemon/pb"
	crypto "gx/ipfs/QmTW4SdgBWq9GjsBsHeUx8WuGxzhgzAf88UMH2w62PC8yK/go-libp2p-crypto"
	peer "gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer"
)

func clientRequestAsync(t *testing.T, client *p2pclient.Client, method string, arg interface{}) interface{} {
	argv := reflect.ValueOf(arg)
	methodv := reflect.ValueOf(client).MethodByName(method)
	elemtype := methodv.Type().Out(0)
	streaming := false
	if elemtype.Kind() == reflect.Chan {
		streaming = true
		elemtype = elemtype.Elem()
	}
	chantype := reflect.ChanOf(reflect.BothDir, elemtype)
	outcv := reflect.MakeChan(chantype, 10)
	go func() {
		defer outcv.Close()
		args := []reflect.Value{
			reflect.ValueOf(context.Background()),
			argv,
		}
		if !streaming {
			args = args[1:]
		}
		res := methodv.Call(args)
		if err, ok := res[1].Interface().(error); ok {
			t.Fatalf("request failed: %s", err)
		}

		if !streaming {
			outcv.Send(res[0])
			return
		}

		for {
			v, ok := res[0].Recv()
			if !ok {
				break
			}
			outcv.Send(v)
		}
	}()

	return outcv.Interface()
}

func TestDHTFindPeer(t *testing.T) {
	daemon, client, closer := createMockDaemonClientPair(t)
	defer closer()
	id := randPeerID(t)

	infoc := clientRequestAsync(t, client, "FindPeer", id).(chan p2pclient.PeerInfo)
	conn := daemon.ExpectConn(t)
	conn.ExpectDHTRequestType(t, pb.DHTRequest_FIND_PEER)
	findPeerResponse := wrapDhtResponse(peerInfoResponse(t, id))
	conn.SendMessage(t, findPeerResponse)
	select {
	case info := <-infoc:
		if info.ID != id {
			t.Fatalf("id %s didn't match expected %s", info.ID, id)
		}
		if len(info.Addrs) != 1 {
			t.Fatalf("expected 1 address, got %d", len(info.Addrs))
		}
		if !bytes.Equal(info.Addrs[0].Bytes(), findPeerResponse.Dht.Peer.Addrs[0]) {
			t.Fatal("address didn't match expected")
		}
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for peer info")
	}
}

func TestDHTGetPublicKey(t *testing.T) {
	daemon, client, closer := createMockDaemonClientPair(t)
	defer closer()
	id := randPeerID(t)
	key := randPubKey(t)

	keyc := clientRequestAsync(t, client, "GetPublicKey", id).(chan crypto.PubKey)
	conn := daemon.ExpectConn(t)
	conn.ExpectDHTRequestType(t, pb.DHTRequest_GET_PUBLIC_KEY)
	keybytes, err := key.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	getKeyResponse := wrapDhtResponse(valueResponse(keybytes))
	conn.SendMessage(t, getKeyResponse)
	select {
	case reskey := <-keyc:
		if !key.Equals(reskey) {
			t.Fatal("keys did not match")
		}
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for peer info")
	}
}

func TestDHTGetValue(t *testing.T) {
	daemon, client, closer := createMockDaemonClientPair(t)
	defer closer()
	key := randBytes(t)
	value := randBytes(t)

	valuec := clientRequestAsync(t, client, "GetValue", key).(chan []byte)
	conn := daemon.ExpectConn(t)
	conn.ExpectDHTRequestType(t, pb.DHTRequest_GET_VALUE)
	getKeyResponse := wrapDhtResponse(valueResponse(value))
	conn.SendMessage(t, getKeyResponse)
	select {
	case resvalue := <-valuec:
		if !bytes.Equal(resvalue, value) {
			t.Fatal("value did not match")
		}
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for peer info")
	}
}

func TestDHTPutValue(t *testing.T) {
	daemon, client, closer := createMockDaemonClientPair(t)
	defer closer()
	key := randBytes(t)
	value := randBytes(t)

	donec := make(chan struct{})
	go func() {
		err := client.PutValue(key, value)
		if err != nil {
			t.Fatal(err)
		}
		donec <- struct{}{}
	}()

	conn := daemon.ExpectConn(t)
	conn.ExpectDHTRequestType(t, pb.DHTRequest_PUT_VALUE)
	putValueResponse := wrapDhtResponse(nil)
	conn.SendMessage(t, putValueResponse)
	select {
	case <-donec:
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for response")
	}
}

func TestDHTProvide(t *testing.T) {
	daemon, client, closer := createMockDaemonClientPair(t)
	defer closer()
	cid := randCid(t)
	donec := make(chan struct{})
	go func() {
		err := client.Provide(cid)
		if err != nil {
			t.Fatal(err)
		}
		donec <- struct{}{}
	}()

	conn := daemon.ExpectConn(t)
	conn.ExpectDHTRequestType(t, pb.DHTRequest_PROVIDE)
	provideResponse := wrapDhtResponse(nil)
	conn.SendMessage(t, provideResponse)
	select {
	case <-donec:
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for response")
	}
}

func TestDHTFindPeersConnectedToPeer(t *testing.T) {
	daemon, client, closer := createMockDaemonClientPair(t)
	defer closer()
	ids := randPeerIDs(t, 3)

	infoc := clientRequestAsync(t, client, "FindPeersConnectedToPeer", ids[0]).(chan p2pclient.PeerInfo)

	conn := daemon.ExpectConn(t)
	req := conn.ExpectDHTRequestType(t, pb.DHTRequest_FIND_PEERS_CONNECTED_TO_PEER)
	if !bytes.Equal(req.GetPeer(), []byte(ids[0])) {
		t.Fatal("request id didn't match expected id")
	}

	resps := make([]*pb.DHTResponse, 2)
	for i := 1; i < 3; i++ {
		resps[i-1] = peerInfoResponse(t, ids[i])
	}
	conn.SendStreamAsync(t, resps)

	i := 0
	for range infoc {
		i++
	}
	if i != 2 {
		t.Fatalf("expected 2 responses, got %d", i)
	}
}

func TestDHTFindProviders(t *testing.T) {
	daemon, client, closer := createMockDaemonClientPair(t)
	defer closer()
	ids := randPeerIDs(t, 3)

	contentID := randCid(t)
	infoc := clientRequestAsync(t, client, "FindProviders", contentID).(chan p2pclient.PeerInfo)

	conn := daemon.ExpectConn(t)
	req := conn.ExpectDHTRequestType(t, pb.DHTRequest_FIND_PROVIDERS)
	if !bytes.Equal(req.GetCid(), contentID.Bytes()) {
		t.Fatal("request cid didn't match expected cid")
	}

	resps := make([]*pb.DHTResponse, 2)
	for i := 1; i < 3; i++ {
		resps[i-1] = peerInfoResponse(t, ids[i])
	}
	conn.SendStreamAsync(t, resps)

	i := 0
	for range infoc {
		i++
	}
	if i != 2 {
		t.Fatalf("expected 2 responses, got %d", i)
	}
}

func TestDHTGetClosestPeers(t *testing.T) {
	daemon, client, closer := createMockDaemonClientPair(t)
	defer closer()
	ids := randPeerIDs(t, 2)
	key := randBytes(t)

	idc := clientRequestAsync(t, client, "GetClosestPeers", key).(chan peer.ID)

	conn := daemon.ExpectConn(t)
	req := conn.ExpectDHTRequestType(t, pb.DHTRequest_GET_CLOSEST_PEERS)
	if !bytes.Equal(req.GetKey(), key) {
		t.Fatal("request key didn't match expected key")
	}
	fmt.Println("we good")

	resps := make([]*pb.DHTResponse, 2)
	for i, id := range ids {
		resps[i] = peerIDResponse(t, id)
	}
	conn.SendStreamAsync(t, resps)

	i := 0
	for range idc {
		i++
	}
	if i != 2 {
		t.Fatalf("expected 2 responses, got %d", i)
	}
}

func TestDHTSearchValue(t *testing.T) {
	daemon, client, closer := createMockDaemonClientPair(t)
	defer closer()
	key := randBytes(t)
	values := make([][]byte, 2)
	for i := range values {
		values[i] = randBytes(t)
	}

	valuec := clientRequestAsync(t, client, "SearchValue", key).(chan []byte)
	conn := daemon.ExpectConn(t)
	conn.ExpectDHTRequestType(t, pb.DHTRequest_SEARCH_VALUE)
	resps := make([]*pb.DHTResponse, 2)
	for i, value := range values {
		resps[i] = valueResponse(value)
	}
	conn.SendStreamAsync(t, resps)
	expiry := time.Now().Add(testTimeout)
	for i := 0; i < 2; i++ {
		select {
		case resvalue := <-valuec:
			if !bytes.Equal(resvalue, values[i]) {
				t.Fatalf("value %d did not match", i)
			}
		case <-time.After(time.Until(expiry)):
			t.Fatal("timed out waiting for values")
		}
	}
}
