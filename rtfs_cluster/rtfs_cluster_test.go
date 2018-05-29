package rtfs_cluster_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfs_cluster"
)

const testPIN = "QmNZiPk974vDsPmQii3YbrMKfi12KTSNM7XMiYyiea4VYZ"

func TestInitialize(t *testing.T) {
	cm := rtfs_cluster.Initialize()
	id, err := cm.Client.ID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(id)
}

func TestParseLocalStatusAllAndSync(t *testing.T) {
	cm := rtfs_cluster.Initialize()
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
	cm := rtfs_cluster.Initialize()
	decoded := cm.DecodeHashString(testPIN)
	err := cm.Pin(decoded)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemovePinFromCluster(t *testing.T) {
	cm := rtfs_cluster.Initialize()
	err := cm.RemovePinFromCluster(testPIN)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFetchLocalStatus(t *testing.T) {
	cm := rtfs_cluster.Initialize()
	cidStatuses, err := cm.FetchLocalStatus()
	if err != nil {
		t.Fatal(err)
	}
	if len(cidStatuses) == 0 {
		fmt.Println("no cids detected")
	}
	fmt.Println(cidStatuses)
}
