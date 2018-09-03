package rtfsp_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfsp"
)

const (
	ipfsConfigPath          = "/ipfs/config"
	testIpfsMultiAddrString = "/ip4/192.168.1.242/tcp/4001/ipfs/QmXivHtDyAe8nS7cbQiS7ri9haUM2wGvbinjKws3a4EstT"
	expectedPeerID          = "QmXivHtDyAe8nS7cbQiS7ri9haUM2wGvbinjKws3a4EstT"
)

func TestConfigGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	var pcm rtfsp.PrivateConfigManager
	err := pcm.ParseConfigAndWrite()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrivateConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

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

	_, err = pcm.GenerateBootstrapPeer(testIpfsMultiAddrString)
	if err != nil {
		t.Fatal(err)
	}
	bpCfg, err := pcm.ConfigureBootstrap(testIpfsMultiAddrString)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(bpCfg)
	fmt.Println(pcm.Config.Bootstrap)
}
