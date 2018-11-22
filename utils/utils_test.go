package utils_test

import (
	"fmt"
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
	manager, err := rtfs.NewManager(nodeOneAPIAddr, nil, time.Minute*10)
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
			if cost <= float64(0) {
				t.Fatal("invalid size returned")
			}
			fmt.Println("cost: ", cost)
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
			if cost <= float64(0) {
				t.Fatal("invalid size returned")
			}
			fmt.Println("cost: ", cost)
		})
	}
}

func TestUtils_CalculateFileSizeInGigaBytes(t *testing.T) {
	size := utils.BytesToGigaBytes(testSize)
	if size != 0.12341962847858667 {
		t.Fatal("failed to calculate correct size")
	}
}

func TestUtils_CalculateAPICallCost(t *testing.T) {
	type args struct {
		callType string
		private  bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"ipns", args{"ipns", false}},
		{"pubsub", args{"pubsub", false}},
		{"ed25519", args{"ed25519", false}},
		{"rsa", args{"rsa", false}},
		{"dlink", args{"dlink", false}},
		{"invalid", args{"invalid", false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, err := utils.CalculateAPICallCost(tt.args.callType, tt.args.private)
			if err != nil && tt.name != "invalid" {
				t.Fatal(err)
			}
			if cost == 0 && tt.name != "invalid" {
				t.Fatal("invalid cost returned")
			}
			if err == nil && tt.name == "invalid" {
				t.Fatal("failed to recognize invalid test name")
			}
		})
	}
}
