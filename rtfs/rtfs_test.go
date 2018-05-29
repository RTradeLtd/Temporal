package rtfs_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfs"
)

func TestInitialize(t *testing.T) {
	im := rtfs.Initialize("")
	nodeInfo, err := im.Shell.ID()
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}
	fmt.Println(nodeInfo)
}
