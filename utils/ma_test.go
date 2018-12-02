package utils_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
)

var testIpfsMultiAddrString = "/ip4/192.168.1.242/tcp/4001/ipfs/QmXivHtDyAe8nS7cbQiS7ri9haUM2wGvbinjKws3a4EstT"
var testP2PMultiAddrString = "/ip4/192.168.1.242/tcp/4001/p2p/QmXivHtDyAe8nS7cbQiS7ri9haUM2wGvbinjKws3a4EstT"
var testInvalidIpfsMultiAddr = "/onion/erhkddypoy6qml6h:4003"
var testPeerID = "QmXivHtDyAe8nS7cbQiS7ri9haUM2wGvbinjKws3a4EstT"

func TestMultiAddrValidator(t *testing.T) {
	if _, err := utils.GenerateMultiAddrFromString("notarealmultiaddr"); err == nil {
		t.Fatal("error expected")
	}
	addr, err := utils.GenerateMultiAddrFromString(testInvalidIpfsMultiAddr)
	if err != nil {
		t.Fatal(err)
	}
	if valid, err := utils.ParseMultiAddrForIPFSPeer(addr); err != nil {
		t.Fatal(err)
	} else if valid {
		t.Fatal("address should not be a valid ipfs address")
	}
	addr, err = utils.GenerateMultiAddrFromString(testIpfsMultiAddrString)
	if err != nil {
		t.Fatal(err)
	}
	if addr == nil {
		t.Fatal("addr is nil when it shouldn't be")
	}
	if valid, err := utils.ParseMultiAddrForIPFSPeer(addr); err != nil {
		t.Fatal(err)
	} else if !valid {
		t.Fatal("address should be a valid ipfs address")
	}
	peerID, err := utils.ParsePeerIDFromIPFSMultiAddr(addr)
	if err != nil {
		t.Fatal(err)
	}
	if peerID != testPeerID {
		t.Fatal("recovered peer id does not match")
	}
	addr, err = utils.GenerateMultiAddrFromString(testP2PMultiAddrString)
	if err != nil {
		t.Fatal(err)
	}
	peerID, err = utils.ParsePeerIDFromIPFSMultiAddr(addr)
	if err != nil {
		t.Fatal(err)
	}
	if peerID != testPeerID {
		t.Fatal("recovered peer id does not match")
	}
}
