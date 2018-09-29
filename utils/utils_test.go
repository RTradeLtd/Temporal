package utils_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/utils"
)

const (
	testHash = "QmdowUuRF4YEJFJvw2TDiECVEMfq89fNVHTXqdN3Z6JM8j"
	testSize = int64(132520817)
)

func TestUtils_CalculatePinCost(t *testing.T) {
	manager, err := rtfs.Initialize("", "")
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
				manager.Shell,
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
