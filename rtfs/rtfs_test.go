package rtfs_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfs"
)

const testPIN = "QmNZiPk974vDsPmQii3YbrMKfi12KTSNM7XMiYyiea4VYZ"

func TestInitialize(t *testing.T) {
	im := rtfs.Initialize("")
	nodeInfo, err := im.Shell.ID()
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}
	fmt.Println(nodeInfo)
}
func TestPin(t *testing.T) {
	im := rtfs.Initialize("")
	err := im.Pin(testPIN)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("pin test successful")
}
