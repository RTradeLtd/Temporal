package utils_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
)

func TestGenerateIpfsMultiHashForFile(t *testing.T) {
	defer func() {
		if err := os.RemoveAll("temp"); err != nil {
			t.Fatal(err)
		}
	}()

	// setup
	if err := os.Mkdir("temp", os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile("temp/test.txt", []byte{}, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	// test
	reader, err := os.Open("temp/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	hash, err := utils.GenerateIpfsMultiHashForFile(reader)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("recovered hash is ", hash)
}
