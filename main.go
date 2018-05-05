package main

import (
	"fmt"
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
		rtfs_cluster.BuildClusterHost()
	case "api":
		database.RunMigrations()
		router := api.Setup()
		router.Run(":6767")
	default:
		fmt.Println("idiot")
	}

}

/*
func main() {
	if len(os.Args) > 2 || len(os.Args) < 2 {
		log.Fatal("not enough arguments")
	}

	m := rtfs.Initialize()

	fileToAdd := os.Args[1]

	fileReader, err := os.Open(fileToAdd)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := m.Shell.Add(fileReader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
}*/
