package rtfscluster_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfscluster"
	gocid "github.com/ipfs/go-cid"
)

const (
	testPIN = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
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
}

func TestInitialize_Failure(t *testing.T) {
	if _, err := rtfscluster.Initialize("10.255.255.255", "9094"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDecodeHashString(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		hash string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Success", args{testPIN}, false},
		{"Failure", args{"notahash"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := cm.DecodeHashString(tt.args.hash); (err != nil) != tt.wantErr {
				t.Fatalf("DecodeHashString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
	type args struct {
		cid gocid.Cid
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Success", args{decoded}, false},
		{"Failure", args{gocid.Cid{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cm.Pin(tt.args.cid); (err != nil) != tt.wantErr {
				t.Fatalf("Pin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
		t.Fatal("no cid statuses found")
	}
}

func TestGetStatusForCidLocally(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		hash string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantStatus bool
	}{
		{"Success", args{testPIN}, false, true},
		{"Failure", args{"blah"}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := cm.GetStatusForCidLocally(tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetStatusForCidLocally() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (status != nil) != tt.wantStatus {
				t.Fatalf("GetStatusForCidLocally() status = %v, wantStatus %v", status, tt.wantStatus)
			}
		})
	}
}

func TestGetStatusForCidGlobally(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		hash string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantStatus bool
	}{
		{"Success", args{testPIN}, false, true},
		{"Failure", args{"blah"}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := cm.GetStatusForCidGlobally(tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetStatusForCidGlobally() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (status != nil) != tt.wantStatus {
				t.Fatalf("GetStatusForCidGlobally() status = %v, wantStatus %v", status, tt.wantStatus)
			}
		})
	}
}

func TestListPeers(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	peers, err := cm.ListPeers()
	if err != nil {
		t.Fatal(err)
	}
	if len(peers) == 0 {
		t.Fatal("no pers found")
	}
}

func TestRemovePinFromCluster(t *testing.T) {
	cm, err := rtfscluster.Initialize(nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		hash string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Success", args{testPIN}, false},
		{"Failure", args{"blah"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cm.RemovePinFromCluster(tt.args.hash); (err != nil) != tt.wantErr {
				t.Fatalf("RemovePinFromCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
			// trigger a pin so we can run more tests
			if !tt.wantErr {
				decoded, err := cm.DecodeHashString(tt.args.hash)
				if err != nil {
					t.Fatal(err)
				}
				if err := cm.Pin(decoded); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
