package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	//_ "./docs"
	"github.com/RTradeLtd/Temporal/api"
	"github.com/RTradeLtd/Temporal/cli"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/queue"
	ipfsQ "github.com/RTradeLtd/Temporal/queue/ipfs"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/rtswarm"
	"github.com/RTradeLtd/Temporal/server"
)

var certFile = "/home/solidity/certificates/api.pem"
var keyFile = "/home/solidity/certificates/api.key"
var tCfg TemporalConfig

func main() {
	if len(os.Args) > 2 || len(os.Args) < 2 {
		fmt.Println("incorrect invocation")
		fmt.Println("./Temporal [api | swarm | queue-dpa | queue-dfa | ipfs-cluster-queue | migrate payment-register-queue | payment-received-queue]")
		fmt.Println("api: run the api, used to interact with temporal")
		fmt.Println("swarm: run the ethereum swarm mode of tempora")
		fmt.Println("queue-dpa: listen to pin requests, and store them in the database")
		fmt.Println("queue-dfa: listen to file add requests, and add to the database")
		fmt.Println("ipfs-cluster-queue: listen to cluster pin pubsub topic")
		fmt.Println("payment-register-queue: listen to the payemnt registered events sent to rabbitmq")
		fmt.Println("payment-received-queue: listen to the paymeny received events sent to rabbitmq")
		fmt.Println("migrate: migrate the database")
		os.Exit(1)
	}
	configDag := os.Getenv("CONFIG_DAG")
	if configDag == "" {
		log.Fatal("CONFIG_DAG is not set")
	}
	tCfg := LoadConfig(configDag)
	certFilePath := tCfg.API.Connection.Certificates.CertPath
	keyFilePath := tCfg.API.Connection.Certificates.KeyPath
	listenAddress := tCfg.API.Connection.ListenAddress
	adminUser := tCfg.API.Admin.Username
	adminPass := tCfg.API.Admin.Password
	jwtKey := tCfg.API.JwtKey
	rollbarToken := tCfg.API.RollbarToken
	rabbitMQConnectionURL := tCfg.RabbitMQ.URL
	dbPass := tCfg.Database.Password
	dbURL := tCfg.Database.URL
	ethKeyFilePath := tCfg.Ethereum.Account.KeyFile
	ethKeyPass := tCfg.Ethereum.Account.KeyPass
	switch os.Args[1] {
	case "config-test":
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("enter dag cid for the config file")
		scanner.Scan()
		configCid := scanner.Text()
		config := LoadConfig(configCid)
		fmt.Printf("%+v\n", config)
	case "api":
		router := api.Setup(adminUser, adminPass, jwtKey, rollbarToken, rabbitMQConnectionURL, dbPass, dbURL, ethKeyFilePath, ethKeyPass)
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
		err = qm.ConsumeMessage("", dbPass, dbURL)
		if err != nil {
			log.Fatal(err)
		}
	case "queue-dfa":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL)
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
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.PaymentRegisterQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL)
		if err != nil {
			log.Fatal(err)
		}
	case "payment-received-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.PaymentReceivedQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL)
		if err != nil {
			log.Fatal(err)
		}
	case "migrate":
		dbm := database.Initialize(dbPass, dbURL)
		dbm.RunMigrations()
	case "contract-backend":
		manager := server.Initialize(false, ethKeyFilePath, ethKeyPass)
		fmt.Println(manager)
	case "cli":
		mqConnectionURL := tCfg.RabbitMQ.URL
		cli.Initialize(mqConnectionURL, ethKeyFilePath, ethKeyPass)
	case "lookup-address":
		db := database.OpenDBConnection(dbPass, dbURL)
		um := models.NewUserManager(db)
		mdl := um.FindByAddress("0xbF43d80dA01332b28cEE39644E8e08AD02a289F5")
		fmt.Println(mdl)
		db.Close()
	case "watch-payments":
		mqConnectionURL := tCfg.RabbitMQ.URL
		sm := server.Initialize(true, ethKeyFilePath, ethKeyPass)
		sm.WaitForAndProcessPaymentsReceivedEvent(mqConnectionURL)
	default:
		fmt.Println("noop")
	}

}
