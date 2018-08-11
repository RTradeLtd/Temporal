package utils_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
)

var testFile = "/tmp/test.txt"

func TestGenerateIpfsMultiHashForFile(t *testing.T) {
	hash, err := utils.GenerateIpfsMultiHashForFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("recovered hash is ", hash)
}
