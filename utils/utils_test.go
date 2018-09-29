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

	cost, err := utils.CalculatePinCost(testHash, 10, manager.Shell)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(cost)
}
