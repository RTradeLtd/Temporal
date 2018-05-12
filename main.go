package main

import (
	"fmt"
	"log"
	"os"

	"github.com/RTradeLtd/Temporal/api"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"
	"github.com/RTradeLtd/Temporal/rtswarm"
)

func main() {
	if len(os.Args) > 2 || len(os.Args) < 2 {
		log.Fatal("idiot")
	}
	token := os.Getenv("ROLLBAR_TOKEN")
	if token == "" {
		os.Exit(1)
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
	case "swarm":
		sm, err := rtswarm.NewSwarmManager()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", sm)
	default:
		fmt.Println("idiot")
	}

}
