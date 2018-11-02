package rtfs_cluster_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfs_cluster"
)

const (
	testPIN        = "QmNZiPk974vDsPmQii3YbrMKfi12KTSNM7XMiYyiea4VYZ"
	nodeOneAPIAddr = "192.168.1.101"
	nodeTwoAPIAddr = "192.168.2.101"
	nodePort       = "9094"
)

func TestInitialize(t *testing.T) {
	cm, err := rtfs_cluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	id, err := cm.Client.ID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(id)
}

func TestParseLocalStatusAllAndSync(t *testing.T) {
	cm, err := rtfs_cluster.Initialize(nodeOneAPIAddr, nodePort)
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
	cm, err := rtfs_cluster.Initialize(nodeOneAPIAddr, nodePort)
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

func TestRemovePinFromCluster(t *testing.T) {
	cm, err := rtfs_cluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	if err = cm.RemovePinFromCluster(testPIN); err != nil {
		t.Fatal(err)
	}
}

func TestFetchLocalStatus(t *testing.T) {
	cm, err := rtfs_cluster.Initialize(nodeOneAPIAddr, nodePort)
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
