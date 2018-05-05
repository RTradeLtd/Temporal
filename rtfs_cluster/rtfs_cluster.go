package rtfs_cluster

import (
	"context"
	"fmt"
	"log"

	ipfsc "github.com/ipfs/ipfs-cluster"
	host "github.com/libp2p/go-libp2p-host"
)

func L() {
	cfg := ipfsc.Config{}
	cfg.Default()
	fmt.Println(cfg)
}

func parseJsonConfig(rawJson []byte) {
	cfg := ipfsc.Config{}
	cfg.LoadJSON(rawJson)
}

func BuildClusterHost() (*host.Host, *ipfsc.Config) {
	cfg := ipfsc.Config{}
	cfg.Default()
	host, err := ipfsc.NewClusterHost(context.TODO(), &cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &host, &cfg
}

/*
func BuildCluster() {
	host, cfg := BuildClusterHost()
	ipfsc.NewCluster(host, cfg)
}*/
