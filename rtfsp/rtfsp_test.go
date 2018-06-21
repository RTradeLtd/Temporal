package rtfsp_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/rtfsp"
)

var ipfsConfigPath = "/ipfs/config"
var testIpfsMultiAddrString = "/ip4/192.168.1.242/tcp/4001/ipfs/QmXivHtDyAe8nS7cbQiS7ri9haUM2wGvbinjKws3a4EstT"
var expectedPeerID = "QmXivHtDyAe8nS7cbQiS7ri9haUM2wGvbinjKws3a4EstT"

func TestPrivateConfig(t *testing.T) {
	pcm, err := rtfsp.GenerateConfigManager(ipfsConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	addr, err := pcm.GenerateIPFSMultiAddr(testIpfsMultiAddrString)
	if err != nil {
		t.Fatal(err)
	}
	peerID := addr.ID().Pretty()
	if peerID != expectedPeerID {
		t.Fatal("unexpected peer id returned")
	}
}
