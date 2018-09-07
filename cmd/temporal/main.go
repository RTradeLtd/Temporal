package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	//_ "./docs"
	"github.com/RTradeLtd/Temporal/api"
	"github.com/RTradeLtd/Temporal/cmd/temporal/app"
	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
)

var (
	// Version denotes the tag of this build
	Version string

	certFile = "/home/solidity/certificates/api.pem"
	keyFile  = "/home/solidity/certificates/api.key"
	tCfg     config.TemporalConfig
)

var commands = map[string]app.Cmd{
	"api": app.Cmd{
		Blurb: "start the api, used to interact with Temporal",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			api, err := api.Initialize(&cfg, true)
			if err != nil {
				log.Fatal(err)
			}
			api.Logger.Info("API service initialized")
			err = api.Router.RunTLS(
				fmt.Sprintf("%s:6767", args["listenAddress"]),
				args["certFilePath"],
				args["keyFilePath"])
			if err != nil {
				msg := fmt.Sprintf("API service execution failed due to the following error: %s", err.Error())
				api.Logger.Fatal(msg)
				fmt.Printf("API execution failed for error %s\nSee logs for more details", err.Error())
			}
		},
	},
	"queue": app.Cmd{
		Blurb:       "execute commands for various queues",
		Description: "Interact with Temporal's various queue APIs",
		Children: map[string]app.Cmd{
			"dfa": app.Cmd{
				Blurb: "listen to file add requests, and add to the database",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqConnectionURL, false, true)
					if err != nil {
						log.Fatal(err)
					}
					err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			"ipfs-pin": app.Cmd{
				Blurb: "",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.Initialize(queue.IpfsPinQueue, mqConnectionURL, false, true)
					if err != nil {
						log.Fatal(err)
					}
					err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			"ipfs-file": app.Cmd{
				Blurb: "",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.Initialize(queue.IpfsFileQueue, mqConnectionURL, false, true)
					if err != nil {
						log.Fatal(err)
					}
					err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			"ipfs-pin-removal": app.Cmd{
				Blurb: "",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.Initialize(queue.IpfsPinRemovalQueue, mqConnectionURL, false, true)
					if err != nil {
						log.Fatal(err)
					}
					err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			"ipfs-key-creation": app.Cmd{
				Blurb: "",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.Initialize(queue.IpfsKeyCreationQueue, mqConnectionURL, false, true)
					if err != nil {
						log.Fatal(err)
					}
					err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			"ipfs-cluster": app.Cmd{
				Blurb: "listen to cluster pin pubsub topic",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.Initialize(queue.IpfsClusterPinQueue, mqConnectionURL, false, true)
					if err != nil {
						log.Fatal(err)
					}
					err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			"pin-payment-confirmation": app.Cmd{
				Blurb: "",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.Initialize(queue.PinPaymentConfirmationQueue, mqConnectionURL, false, true)
					if err != nil {
						log.Fatal(err)
					}
					err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			"pin-payment-submission": app.Cmd{
				Blurb: "",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.Initialize(queue.PinPaymentSubmissionQueue, mqConnectionURL, false, true)
					if err != nil {
						log.Fatal(err)
					}
					err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			"email-send": app.Cmd{
				Blurb: "",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.Initialize(queue.EmailSendQueue, mqConnectionURL, false, true)
					if err != nil {
						log.Fatal(err)
					}
					err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			"ipns-entry": app.Cmd{
				Blurb: "",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.Initialize(queue.IpnsEntryQueue, mqConnectionURL, false, true)
					if err != nil {
						log.Fatal(err)
					}
					err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
		},
	},
	"calculcate-config-checksum": app.Cmd{
		Blurb: "",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			fileBytes, err := ioutil.ReadFile(args["configDag"])
			if err != nil {
				log.Fatal(err)
			}
			hash, err := utils.CalculateConfigFileChecksum(fileBytes)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("calculated config file checksum is %s\n", hash)
			// TODO: hardcode the checksum value so we can do a check here
		},
	},
	"migrate": app.Cmd{
		Blurb: "run database migrations",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			dbm, err := database.Initialize(&cfg, false)
			if err != nil {
				log.Fatal(err)
			}
			dbm.RunMigrations()
		},
	},
}

func main() {
	// create app
	temporal := app.New(commands, app.Config{
		Name:     "Temporal",
		ExecName: "temporal",
		Version:  Version,
		Desc:     "Temporal is an easy-to-use interface into distributed and decentralized storage technologies for personal and enterprise use cases.",
	})

	// run no-config commands
	exit := temporal.PreRun(os.Args[1:])
	if exit == app.CodeOK {
		os.Exit(0)
	}

	// load config
	configDag := os.Getenv("CONFIG_DAG")
	if configDag == "" {
		log.Fatal("CONFIG_DAG is not set")
	}
	tCfg, err := config.LoadConfig(configDag)
	if err != nil {
		log.Fatal(err)
	}

	// load arguments
	flags := map[string]string{
		"configDag":     configDag,
		"certFilePath":  tCfg.API.Connection.Certificates.CertPath,
		"keyFilePath":   tCfg.API.Connection.Certificates.KeyPath,
		"listenAddress": tCfg.API.Connection.ListenAddress,

		"dbPass": tCfg.Database.Password,
		"dbURL":  tCfg.Database.URL,
		"dbUser": tCfg.Database.Username,
	}

	// execute
	os.Exit(temporal.Run(*tCfg, flags, os.Args[1:]))
}
