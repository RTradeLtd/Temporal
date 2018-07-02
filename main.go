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
	tCfg := config.LoadConfig(configDag)
	certFilePath := tCfg.API.Connection.Certificates.CertPath
	keyFilePath := tCfg.API.Connection.Certificates.KeyPath
	listenAddress := tCfg.API.Connection.ListenAddress
	jwtKey := tCfg.API.JwtKey
	rabbitMQConnectionURL := tCfg.RabbitMQ.URL
	dbPass := tCfg.Database.Password
	dbURL := tCfg.Database.URL
	dbUser := tCfg.Database.Username
	ethKeyFilePath := tCfg.Ethereum.Account.KeyFile
	ethKeyPass := tCfg.Ethereum.Account.KeyPass
	awsKey := tCfg.AWS.KeyID
	awsSecret := tCfg.AWS.Secret
	switch os.Args[1] {
	case "api":
		router := api.Setup(jwtKey, rabbitMQConnectionURL, dbPass, dbURL, ethKeyFilePath, ethKeyPass, listenAddress, dbUser, awsKey, awsSecret)
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
		err = qm.ConsumeMessage("", dbPass, dbURL, "", "", dbUser)
		if err != nil {
			log.Fatal(err)
		}
	case "queue-dfa":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, "", "", dbUser)
		if err != nil {
			log.Fatal(err)
		}
	case "ipfs-cluster-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.IpfsClusterQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, "", "", dbUser)
		if err != nil {
			log.Fatal(err)
		}
	case "payment-register-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.PaymentRegisterQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, ethKeyFilePath, ethKeyPass, dbUser)
		if err != nil {
			log.Fatal(err)
		}
	case "payment-received-queue":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.PaymentReceivedQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, "", "", dbUser)
		if err != nil {
			log.Fatal(err)
		}
	case "pin-payment-request":
		mqConnectionURL := tCfg.RabbitMQ.URL
		qm, err := queue.Initialize(queue.PinPaymentRequestQueue, mqConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
		err = qm.ConsumeMessage("", dbPass, dbURL, ethKeyFilePath, ethKeyPass, dbUser)
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
