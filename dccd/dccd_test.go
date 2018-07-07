package dccd_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/dccd"
)

var testHash = "Qmbu7x6gJbsKDcseQv66pSbUcAA3Au6f7MfTYVXwvBxN2K"

func TestDCCDComplete(t *testing.T) {
	manager := dccd.NewDCCDManager("https://ipfs.io")
	resp, err := manager.Shell.Cat(testHash)
	if err != nil {
		t.Fatal(err)
	}
	err = resp.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDCCD(t *testing.T) {
	manager := dccd.NewDCCDManager("")
	// Parse gateway array
	manager.ParseGateways()
	dispersals, err := manager.DisperseContent(testHash)
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
