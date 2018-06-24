package utils_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
)

var testIpfsMultiAddrString = "/ip4/192.168.1.242/tcp/4001/ipfs/QmXivHtDyAe8nS7cbQiS7ri9haUM2wGvbinjKws3a4EstT"
var testP2PMultiAddrString = "/ip4/192.168.1.242/tcp/4001/ipfs/QmXivHtDyAe8nS7cbQiS7ri9haUM2wGvbinjKws3a4EstT"

func TestMultiAddrValidator(t *testing.T) {
	addr, err := utils.GenerateMultiAddrFromString(testIpfsMultiAddrString)
	if err != nil {
		t.Fatal(err)
	}

	_, err = utils.ParseMultiAddrForBootstrap(addr)
	if err != nil {
		t.Fatal(err)
	}

	addr, err = utils.GenerateMultiAddrFromString(testP2PMultiAddrString)
	if err != nil {
		t.Fatal(err)
	}

	_, err = utils.ParseMultiAddrForBootstrap(addr)
	if err != nil {
		t.Fatal(err)
	}

	pretty, err := utils.ParsePeerIDFromIPFSMultiAddr(addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(pretty)
}
