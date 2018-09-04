package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	//_ "./docs"
	"github.com/RTradeLtd/Temporal/api"
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

type cmd struct {
	blurb  string
	action func(config.TemporalConfig, map[string]string)
	hidden bool
}

var commands = map[string]cmd{
	"api": cmd{
		blurb: "start the api, used to interact with Temporal",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"queue-dfa": cmd{
		blurb: "listen to file add requests, and add to the database",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"ipfs-pin-queue": cmd{
		blurb: "",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"ipfs-file-queue": cmd{
		blurb: "",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"pin-payment-confirmation-queue": cmd{
		blurb: "",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"pin-payment-submission-queue": cmd{
		blurb: "",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"email-send-queue": cmd{
		blurb: "",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"ipns-entry-queue": cmd{
		blurb: "",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"ipfs-pin-removal-queue": cmd{
		blurb: "",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"ipfs-key-creation-queue": cmd{
		blurb: "",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"calculcate-config-checksum": cmd{
		blurb: "",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"ipfs-cluster-queue": cmd{
		blurb: "listen to cluster pin pubsub topic",
		action: func(cfg config.TemporalConfig, args map[string]string) {
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
	"migrate": cmd{
		blurb: "run database migrations",
		action: func(cfg config.TemporalConfig, args map[string]string) {
			dbm, err := database.Initialize(&cfg, false)
			if err != nil {
				log.Fatal(err)
			}
			dbm.RunMigrations()
		},
	},
}

func main() {
	// guard against invalid arg count
	if len(os.Args) > 2 || len(os.Args) < 2 {
		printNoOp(os.Args, commands)
		os.Exit(1)
	}

	// build-in invocations
	switch os.Args[1] {
	case "help":
		printHelp(commands)
		os.Exit(0)
	case "version":
		println("temporal " + Version)
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
	args := map[string]string{
		"configDag":     configDag,
		"certFilePath":  tCfg.API.Connection.Certificates.CertPath,
		"keyFilePath":   tCfg.API.Connection.Certificates.KeyPath,
		"listenAddress": tCfg.API.Connection.ListenAddress,

		"dbPass": tCfg.Database.Password,
		"dbURL":  tCfg.Database.URL,
		"dbUser": tCfg.Database.Username,
	}

	// find and execute command
	invocation, ok := commands[os.Args[1]]
	if !ok {
		printNoOp(os.Args, commands)
		os.Exit(1)
	}
	if invocation.action == nil {
		fmt.Printf("no action found for '%s'\n", strings.Join(os.Args[:], " "))
		os.Exit(1)
	}
	invocation.action(*tCfg, args)
}
