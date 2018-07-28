package main

import (
	"fmt"
	"log"
	"os"

	//_ "./docs"
	"github.com/RTradeLtd/Temporal/api"
	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtswarm"
)

var certFile = "/home/solidity/certificates/api.pem"
var keyFile = "/home/solidity/certificates/api.key"
var tCfg config.TemporalConfig

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
	configDag := os.Getenv("CONFIG_DAG")
	if configDag == "" {
		log.Fatal("CONFIG_DAG is not set")
	}
	tCfg, err := config.LoadConfig(configDag)
	if err != nil {
		log.Fatal(err)
	}
	certFilePath := tCfg.API.Connection.Certificates.CertPath
	keyFilePath := tCfg.API.Connection.Certificates.KeyPath
	listenAddress := tCfg.API.Connection.ListenAddress
	dbPass := tCfg.Database.Password
	dbURL := tCfg.Database.URL
	dbUser := tCfg.Database.Username
	ethKeyFilePath := tCfg.Ethereum.Account.KeyFile
	ethKeyPass := tCfg.Ethereum.Account.KeyPass
	switch os.Args[1] {
	case "api":
		router := api.Setup(tCfg)
		router.RunTLS(fmt.Sprintf("%s:6767", listenAddress), certFilePath, keyFilePath)
	case "swarm":
		sm, err := rtswarm.NewSwarmManager()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", sm)
	case "queue-dpa":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.DatabasePinAddQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, "", "", dbUser, tCfg)
		if err != nil {
			log.Fatal(err)
		}
	case "queue-dfa":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, "", "", dbUser, tCfg)
		if err != nil {
			log.Fatal(err)
		}
	case "ipfs-pin-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.IpfsPinQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, ethKeyFilePath, ethKeyPass, dbUser, tCfg)
		if err != nil {
			log.Fatal(err)
		}
	case "ipfs-file-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.IpfsFileQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, ethKeyFilePath, ethKeyPass, dbUser, tCfg)
		if err != nil {
			log.Fatal(err)
		}
	case "pin-payment-confirmation-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.PinPaymentConfirmationQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, ethKeyFilePath, ethKeyPass, dbUser, tCfg)
		if err != nil {
			log.Fatal(err)
		}
	case "pin-payment-submission-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.PinPaymentSubmissionQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, ethKeyFilePath, ethKeyPass, dbUser, tCfg)
		if err != nil {
			log.Fatal(err)
		}
	case "email-send-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.EmailSendQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, ethKeyFilePath, ethKeyPass, dbUser, tCfg)
		if err != nil {
			log.Fatal(err)
		}
	case "ipns-entry-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.IpnsEntryQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, ethKeyFilePath, ethKeyPass, dbUser, tCfg)
		if err != nil {
			log.Fatal(err)
		}
	case "ipfs-pin-removal-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.IpfsPinRemovalQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, ethKeyFilePath, ethKeyPass, dbUser, tCfg)
		if err != nil {
			log.Fatal(err)
		}
	case "migrate":
		dbm, err := database.Initialize(dbPass, dbURL, dbUser)
		if err != nil {
			log.Fatal(err)
		}
		dbm.RunMigrations()
	default:
		fmt.Println("noop")
	}

}
