package dccd_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/dccd"
)

var testHash = "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"

func TestDCCD(t *testing.T) {
	manager := dccd.NewDCCDManager("")
	// Parse gateway array
	manager.ParseGateways()
	dispersals, err := manager.DisperseContent(testHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dispersals)
}
