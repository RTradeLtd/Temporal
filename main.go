package main

import (
	"fmt"
	"log"
	"os"

	"github.com/RTradeLtd/RTC-IPFS/rtfs"
)

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
}
