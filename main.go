package main

import (
	"fmt"
	"log"
	"os"

	"github.com/RTradeLtd/Temporal/api"
	"github.com/RTradeLtd/Temporal/api/rtfs"
	"github.com/RTradeLtd/Temporal/cli"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/queue"
	ipfsQ "github.com/RTradeLtd/Temporal/queue/ipfs"
	"github.com/RTradeLtd/Temporal/rtswarm"
	"github.com/RTradeLtd/Temporal/server"
)

var certFile = "/home/solidity/certificates/api.crt"
var keyFile = "/home/solidity/certificates/api.key"

func main() {
	if len(os.Args) > 2 || len(os.Args) < 2 {
		fmt.Println("incorrect invocation")
		fmt.Println("./Temporal [api | swarm | queue-dpa | queue-dfa | ipfs-cluster-queue | migrate]")
		fmt.Println("api: run the api, used to interact with temporal")
		fmt.Println("swarm: run the ethereum swarm mode of tempora")
		fmt.Println("queue-dpa: listen to pin requests, and store them in the database")
		fmt.Println("queue-dfa: listen to file add requests, and add to the database")
		fmt.Println("ipfs-cluster-queue: listen to cluster pin pubsub topic")
		fmt.Println("migrate: migrate the database")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "api":
		listenAddress := os.Getenv("LISTEN_ADDRESS")
		if listenAddress == "" {
			fmt.Println("invalid address")
			fmt.Println("Please set LISTEN_ADDRESS env to a valid ip address")
			os.Exit(1)
		}
		router := api.Setup()
		router.RunTLS(fmt.Sprintf("%s:6767", listenAddress), certFile, keyFile)
	case "swarm":
		sm, err := rtswarm.NewSwarmManager()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", sm)
	case "queue-dpa":
		qm, err := queue.Initialize(queue.DatabasePinAddQueue)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("")
		if err != nil {
			log.Fatal(err)
		}
	case "queue-dfa":
		qm, err := queue.Initialize(queue.DatabaseFileAddQueue)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("")
		if err != nil {
			log.Fatal(err)
		}
	case "ipfs-cluster-queue":
		psm, err := ipfsQ.Initialize(rtfs.ClusterPubSubTopic)
		if err != nil {
			log.Fatal(err)
		}
		psm.ParseClusterPinTopic()
	case "payment-register-queue":
		qm, err := queue.Initialize(queue.PaymentRegisterQueue)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("")
		if err != nil {
			log.Fatal(err)
		}
	case "migrate":
		dbm := database.Initialize()
		dbm.RunMigrations()
	case "contract-backend":
		manager := server.Initialize()
		fmt.Println(manager)
	case "shell":
		cli.Initialize()
	case "lookup-address":
		db := database.OpenDBConnection()
		um := models.NewUserManager(db)
		mdl := um.FindByAddress("0xbF43d80dA01332b28cEE39644E8e08AD02a289F5")
		fmt.Println(mdl)
		db.Close()
	default:
		fmt.Println("noop")
	}

}
