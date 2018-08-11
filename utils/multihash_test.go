package utils_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
)

var testFile = "/tmp/test.txt"

func TestGenerateIpfsMultiHashForFile(t *testing.T) {
	reader, err := os.Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := utils.GenerateIpfsMultiHashForFile(reader)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("recovered hash is ", hash)
}
