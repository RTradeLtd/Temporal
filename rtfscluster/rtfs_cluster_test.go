package rtfscluster_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfscluster"
)

const (
	testPIN = "QmNZiPk974vDsPmQii3YbrMKfi12KTSNM7XMiYyiea4VYZ"
	// ip address of our first ipfs and ipfs cluster node as per our makefile
	nodeOneAPIAddr = "192.168.1.101"
	// ip address of our second ipfs and ipfs cluster node as per our makefile
	nodeTwoAPIAddr = "192.168.2.101"
	// this is the port of the IPFS Cluster API
	nodePort = "9094"
)

func TestInitialize(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	id, err := cm.Client.ID()
	if err != nil {
		t.Fatal(err)
	}
	if id.Version == "" {
		t.Fatal("version is empty string when it shouldn't be")
	}
	fmt.Println(id)
}

func TestParseLocalStatusAllAndSync(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	syncedCids, err := cm.ParseLocalStatusAllAndSync()
	if err != nil {
		t.Fatal(err)
	}
	if len(syncedCids) > 0 {
		fmt.Println("uh oh cluster errors detected")
		fmt.Println("this isn't indicative of a test failure")
		fmt.Println("but that the cluster is experiencing some issues")
	} else {
		fmt.Println("yay no cluster errors detected")
	}
}

func TestClusterPin(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := cm.DecodeHashString(testPIN)
	if err != nil {
		t.Fatal(err)
	}
	if err = cm.Pin(decoded); err != nil {
		t.Fatal(err)
	}
}

func TestFetchLocalStatus(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	cidStatuses, err := cm.FetchLocalStatus()
	if err != nil {
		t.Fatal(err)
	}
	if len(cidStatuses) == 0 {
		fmt.Println("no cids detected")
	}
	fmt.Println(cidStatuses)
}

func TestGetStatusForCidLocally(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}

	status, err := cm.GetStatusForCidLocally(testPIN)
	if err != nil {
		t.Fatal(err)
	}

	if status == nil {
		t.Fatal("status is nil when it shouldn't be")
	}
}

func TestGetStatusForCidGlobally(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	status, err := cm.GetStatusForCidGlobally(testPIN)
	if err != nil {
		t.Fatal(err)
	}

	if status == nil {
		t.Fatal("status is nil when it shouldn't be")
	}
}

func TestListPeers(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = cm.ListPeers(); err != nil {
		t.Fatal(err)
	}
}
func TestRemovePinFromCluster(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	if err = cm.RemovePinFromCluster(testPIN); err != nil {
		t.Fatal(err)
	}
}
