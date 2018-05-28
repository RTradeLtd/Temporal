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
	tCfg := LoadConfig(configDag)

	switch os.Args[1] {
	case "config-test":
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("enter dag cid for the config file")
		scanner.Scan()
		configCid := scanner.Text()
		config := LoadConfig(configCid)
		fmt.Printf("%+v\n", config)
	case "api":
		certFilePath := tCfg.API.Connection.Certificates.CertPath
		keyFilePath := tCfg.API.Connection.Certificates.KeyPath
		listenAddress := tCfg.API.Connection.ListenAddress
		adminUser := tCfg.API.Admin.Username
		adminPass := tCfg.API.Admin.Password
		jwtKey := tCfg.API.JwtKey
		rollbarToken := tCfg.API.RollbarToken
		rabbitMQConnectionURL := tCfg.RabbitMQ.URL
		router := api.Setup(adminUser, adminPass, jwtKey, rollbarToken, rabbitMQConnectionURL)
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
		err = qm.ConsumeMessage("")
		if err != nil {
			log.Fatal(err)
		}
	case "queue-dfa":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqConnectionURL)
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
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.PaymentRegisterQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("")
		if err != nil {
			log.Fatal(err)
		}
	case "payment-received-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.PaymentReceivedQueue, mqConnectionURL)
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
		manager := server.Initialize(false)
		fmt.Println(manager)
	case "cli":
		mqConnectionURL := tCfg.RabbitMQ.URL
		cli.Initialize(mqConnectionURL)
	case "lookup-address":
		db := database.OpenDBConnection()
		um := models.NewUserManager(db)
		mdl := um.FindByAddress("0xbF43d80dA01332b28cEE39644E8e08AD02a289F5")
		fmt.Println(mdl)
		db.Close()
	case "watch-payments":
		mqConnectionURL := tCfg.RabbitMQ.URL
		sm := server.Initialize(true)
		sm.WaitForAndProcessPaymentsReceivedEvent(mqConnectionURL)
	default:
		fmt.Println("noop")
	}

}
