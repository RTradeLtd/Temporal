package utils_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/utils"
)

const (
	testHash = "QmdowUuRF4YEJFJvw2TDiECVEMfq89fNVHTXqdN3Z6JM8j"
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
				t.Fatal(err)
			}
			fmt.Println("cost: ", cost)
		})
	}
}
