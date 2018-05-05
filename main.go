package main

import (
	"github.com/RTradeLtd/RTC-IPFS/api"
	"github.com/RTradeLtd/RTC-IPFS/database"
)

func main() {
	database.RunMigrations()
	router := api.Setup()
	router.Run(":6767")
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
