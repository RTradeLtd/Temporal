package main

import (
	"fmt"
	"log"
	"os"

	"github.com/RTradeLtd/RTC-IPFS/api"
	"github.com/RTradeLtd/RTC-IPFS/database"
	"github.com/RTradeLtd/RTC-IPFS/rtfs_cluster"
)

func main() {
	if len(os.Args) > 2 || len(os.Args) < 2 {
		fmt.Println("idiot")
	}
	switch os.Args[1] {
	case "cluster":
		cm := rtfs_cluster.Initialize()
		cm.GenRestAPIConfig()
		cm.GenClient()
		cm.ParseLocalStatusAllAndSync()
		cid := cm.DecodeHashString("QmXXSSQpbYhGRMPqqZ4gF1SjqBkBjpnb44JuR1frwL1RiA")
		err := cm.Pin(cid)
		if err != nil {
			log.Fatal(err)
		}
	case "api":
		database.RunMigrations()
		router := api.Setup()
		router.Run(":6767")
	default:
		fmt.Println("idiot")
	}

}
