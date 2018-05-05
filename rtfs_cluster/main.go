package rtfs_cluster

import (
	"fmt"

	ipfsc "github.com/ipfs/ipfs-cluster"
)

func L() {
	cfg := ipfsc.Config{}
	cfg.Default()
	fmt.Println(cfg)
}
