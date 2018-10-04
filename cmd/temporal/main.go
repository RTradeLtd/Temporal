package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/RTradeLtd/Temporal/api"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/cmd"
	"github.com/RTradeLtd/config"
)

var (
	// Version denotes the tag of this build
	Version string

	certFile = filepath.Join(os.Getenv("HOME"), "/certificates/api.pem")
	keyFile  = filepath.Join(os.Getenv("HOME"), "/certificates/api.key")
	tCfg     config.TemporalConfig
)

var commands = map[string]cmd.Cmd{
	"api": cmd.Cmd{
		Blurb:       "start Temporal api server",
		Description: "Start the API service used to interact with Temporal. Run with DEBUG=true to enable debug messages.",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			service, err := api.Initialize(&cfg, os.Getenv("DEBUG") == "true")
			if err != nil {
				log.Fatal(err)
			}

			addr := fmt.Sprintf("%s:6767", args["listenAddress"])
			if args["certFilePath"] == "" || args["keyFilePath"] == "" {
				fmt.Println("TLS config incomplete - starting API service without TLS...")
				err = service.ListenAndServe(addr, nil)
			} else {
				fmt.Println("Starting API service with TLS...")
				err = service.ListenAndServe(addr, &api.TLSConfig{
					CertFile: args["certFilePath"],
					KeyFile:  args["keyFilePath"],
				})
			}
			if err != nil {
				fmt.Printf("API service execution failed: %s\n", err.Error())
				fmt.Println("Refer to the logs for more details")
			}
		},
	},
	"queue": cmd.Cmd{
		Blurb:         "execute commands for various queues",
		Description:   "Interact with Temporal's various queue APIs",
		ChildRequired: true,
		Children: map[string]cmd.Cmd{
			"ipfs": cmd.Cmd{
				Blurb:         "IPFS queue sub commands",
				Description:   "Used to launch the various queues that interact with IPFS",
				ChildRequired: true,
				Children: map[string]cmd.Cmd{
					"ipns-entry": cmd.Cmd{
						Blurb:       "IPNS entry creation queue",
						Description: "Listens to requests to create IPNS records",
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
					"pin": cmd.Cmd{
						Blurb:       "Pin addition queue",
						Description: "Listens to pin requests",
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
					"file": cmd.Cmd{
						Blurb:       "File upload queue",
						Description: "Listens to file upload requests. Only applies to advanced uploads",
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
					"key-creation": cmd.Cmd{
						Blurb:       "Key creation queue",
						Description: fmt.Sprintf("Listen to key creation requests.\nMessages to this queue are broadcasted to all nodes"),
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
					"cluster": cmd.Cmd{
						Blurb:       "Cluster pin queue",
						Description: "Listens to requests to pin content to the cluster",
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
				},
			},
			"dfa": cmd.Cmd{
				Blurb:       "Database file add queue",
				Description: "Listens to file uploads requests. Only applies to simple upload route",
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
			"email-send": cmd.Cmd{
				Blurb:       "Email send queue",
				Description: "Listens to requests to send emails",
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
		},
	},
	"migrate": cmd.Cmd{
		Blurb:       "run database migrations",
		Description: "Runs our initial database migrations, creating missing tables, etc..",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if _, err := database.Initialize(&cfg, database.DatabaseOptions{
				RunMigrations: true,
			}); err != nil {
				log.Fatal(err)
			}
		},
	},
	"migrate-insecure": cmd.Cmd{
		Hidden:      true,
		Blurb:       "run database migrations without SSL",
		Description: "Runs our initial database migrations, creating missing tables, etc.. without SSL",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if _, err := database.Initialize(&cfg, database.DatabaseOptions{
				RunMigrations:  true,
				SSLModeDisable: true,
			}); err != nil {
				log.Fatal(err)
			}
		},
	},
	"init": cmd.Cmd{
		PreRun:      true,
		Blurb:       "initialize blank Temporal configuration",
		Description: "Initializes a blank Temporal configuration template at CONFIG_DAG.",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			configDag := os.Getenv("CONFIG_DAG")
			if configDag == "" {
				log.Fatal("CONFIG_DAG is not set")
			}
			if err := config.GenerateConfig(configDag); err != nil {
				log.Fatal(err)
			}
		},
	},
}

func main() {
	// create app
	temporal := cmd.New(commands, cmd.Config{
		Name:     "Temporal",
		ExecName: "temporal",
		Version:  Version,
		Desc:     "Temporal is an easy-to-use interface into distributed and decentralized storage technologies for personal and enterprise use cases.",
	})

	// run no-config commands, exit if command was run
	if exit := temporal.PreRun(os.Args[1:]); exit == cmd.CodeOK {
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
