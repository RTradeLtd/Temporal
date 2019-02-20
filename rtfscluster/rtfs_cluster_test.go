package rtfscluster_test

import (
	"context"
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
	cm, err := rtfscluster.Initialize(context.Background(), nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	id, err := cm.Client.ID(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if id.Version == "" {
		t.Fatal("version is empty string when it shouldn't be")
	}
}

func TestInitialize_Failure(t *testing.T) {
	if _, err := rtfscluster.Initialize(context.Background(), "10.255.255.255", "9094"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDecodeHashString(t *testing.T) {
	cm, err := rtfscluster.Initialize(context.Background(), nodeOneAPIAddr, nodePort)
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
	cm, err := rtfscluster.Initialize(context.Background(), nodeOneAPIAddr, nodePort)
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
			if err := cm.Pin(context.Background(), tt.args.cid); (err != nil) != tt.wantErr {
				t.Fatalf("Pin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListPeers(t *testing.T) {
	cm, err := rtfscluster.Initialize(context.Background(), nodeOneAPIAddr, nodePort)
	if err != nil {
		t.Fatal(err)
	}
	peers, err := cm.ListPeers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(peers) == 0 {
		t.Fatal("no pers found")
	}
}
