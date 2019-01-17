package utils_test

import (
	"testing"
	"time"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/RTradeLtd/rtfs"
)

const (
	testHash       = "QmNZiPk974vDsPmQii3YbrMKfi12KTSNM7XMiYyiea4VYZ"
	nodeOneAPIAddr = "192.168.1.101:5001"
	testSize       = int64(132520817)
)

func TestUtils_CalculatePinCost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}
	manager, err := rtfs.NewManager(nodeOneAPIAddr, time.Minute*10)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		hash    string
		months  int64
		private bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"Public", args{testHash, int64(10), false}},
		{"Private", args{testHash, int64(10), true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, err := utils.CalculatePinCost(
				tt.args.hash,
				tt.args.months,
				manager,
				tt.args.private,
			)
			if err != nil {
				t.Fatal(err)
			}
			if cost == 0 {
				t.Fatal("invalid size returned")
			}
		})
	}
}

func TestUtils_CalculateFileCost(t *testing.T) {
	type args struct {
		size    int64
		months  int64
		private bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"Public", args{testSize, int64(10), false}},
		{"Private", args{testSize, int64(10), true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := utils.CalculateFileCost(
				tt.args.months,
				tt.args.size,
				tt.args.private,
			)
			if cost == 0 {
				t.Fatal("invalid size returned")
			}
		})
	}
}

func TestUtils_CalculateAPICallCost(t *testing.T) {
	type args struct {
		callType string
		private  bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ipns-public", args{"ipns", false}, false},
		{"ipns-private", args{"ipns", true}, false},
		{"pubsub-public", args{"pubsub", false}, false},
		{"pubsub-private", args{"pubsub", true}, false},
		{"ed25519-public", args{"ed25519", false}, false},
		{"ed25519-private", args{"ed25519", true}, false},
		{"rsa-public", args{"rsa", false}, false},
		{"rsa-private", args{"rsa", true}, false},
		{"invalid", args{"invalid", false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, err := utils.CalculateAPICallCost(tt.args.callType, tt.args.private)
			if (err != nil) != tt.wantErr {
				t.Fatal(err)
			}
			if cost == 0 && tt.name != "invalid" {
				t.Fatal("invalid cost returned")
			}
		})
	}
}

func TestUtils_FloatToBigInt(t *testing.T) {
	want := int64(500000000000000000)
	bigInt := utils.FloatToBigInt(0.5)
	if bigInt.Int64() != want {
		t.Fatal("failed to properly calculate big int")
	}
}
