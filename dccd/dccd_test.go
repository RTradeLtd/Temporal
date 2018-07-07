package dccd_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/dccd"
)

var testHash = "Qmbu7x6gJbsKDcseQv66pSbUcAA3Au6f7MfTYVXwvBxN2K"

func TestDispersal(t *testing.T) {
	manager := dccd.NewDCCDManager("")
	resp, err := manager.DisperseContentWithShell(testHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}

func TestDCCD(t *testing.T) {
	manager := dccd.NewDCCDManager("")
	// Parse gateway array
	manager.ParseGateways()
	dispersals, err := manager.DisperseContentWithShell(testHash)
	if err != nil {
		t.Fatal(err)
	}
	var failureCount int
	var successCount int
	for k, v := range dispersals {
		if v == false {
			fmt.Printf("dispersal for %s failed\n", k)
			failureCount++
			continue
		}
		fmt.Printf("dispersal for %s passed", k)
		successCount++
	}
	fmt.Println("Number of failures", failureCount)
	fmt.Println("Number of successes", successCount)
}
