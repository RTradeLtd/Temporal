package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/RTradeLtd/rtfs"

	"github.com/RTradeLtd/Temporal/tns"

	"github.com/RTradeLtd/Temporal/api"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/cmd"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
)

var (
	// Version denotes the tag of this build
	Version string

	certFile = filepath.Join(os.Getenv("HOME"), "/certificates/api.pem")
	keyFile  = filepath.Join(os.Getenv("HOME"), "/certificates/api.key")
	tCfg     config.TemporalConfig
)

var commands = map[string]cmd.Cmd{
	"tns": {
		Blurb:         "run a tns daemon or client",
		Description:   "allows running a tns daemon to manage a zone, or a client to query a dameon",
		ChildRequired: true,
		Children: map[string]cmd.Cmd{
			"daemon": {
				Blurb:       "run tns daemon",
				Description: "runs a tns daemon and zone manager",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					keystore, err := rtfs.NewKeystoreManager(cfg.IPFS.KeystorePath)
					if err != nil {
						log.Fatal(err)
					}
					zoneManagerPK, err := keystore.GetPrivateKeyByName(cfg.TNS.ZoneManagerKeyName)
					if err != nil {
						log.Fatal(err)
					}
					zonePK, err := keystore.GetPrivateKeyByName(cfg.TNS.ZoneManagerKeyName)
					if err != nil {
						log.Fatal(err)
					}
					managerOpts := tns.ManagerOpts{
						ManagerPK: zoneManagerPK,
						ZonePK:    zonePK,
						ZoneName:  cfg.TNS.ZoneName,
					}
					dbm, err := database.Initialize(&cfg, database.Options{})
					if err != nil {
						log.Fatal(err)
					}
					manager, err := tns.GenerateTNSManager(&managerOpts, dbm.DB)
					if err != nil {
						log.Fatal(err)
					}
					if err = manager.MakeHost(manager.PrivateKey, nil); err != nil {
						log.Fatal(err)
					}
					defer manager.Host.Close()
					manager.RunTNSDaemon()
					lim := len(manager.Host.Addrs())
					count := 0
					for count < lim {
						fmt.Println(manager.ReachableAddress(count))
						count++
					}
					select {}
				},
			},
			"client": {
				Blurb:       "run tns client",
				Description: "runs a tns client to make libp2p connections to a tns daemon",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					peerAddr := args["peerAddr"]
					if peerAddr == "" {
						log.Fatal("peerAddr argument is empty")
					}
					client, err := tns.GenerateTNSClient(true, nil)
					if err != nil {
						log.Fatal(err)
					}
					if err = client.MakeHost(client.PrivateKey, nil); err != nil {
						log.Fatal(err)
					}
					defer client.Host.Close()
					pid, err := client.AddPeerToPeerStore(peerAddr)
					if err != nil {
						log.Fatal(err)
					}
					if _, err = client.QueryTNS(pid, "echo", nil); err != nil {
						log.Fatal(err)
					}
				},
			},
		},
	},
	"api": {
		Blurb:       "start Temporal api server",
		Description: "Start the API service used to interact with Temporal. Run with DEBUG=true to enable debug messages.",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			service, err := api.Initialize(&cfg, os.Getenv("DEBUG") == "true")
			if err != nil {
				log.Fatal(err)
			}
			defer service.Close()

			port := os.Getenv("API_PORT")
			if port == "" {
				port = "6767"
			}
			addr := fmt.Sprintf("%s:%s", args["listenAddress"], port)
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
	"queue": {
		Blurb:         "execute commands for various queues",
		Description:   "Interact with Temporal's various queue APIs",
		ChildRequired: true,
		Children: map[string]cmd.Cmd{
			"ipfs": {
				Blurb:         "IPFS queue sub commands",
				Description:   "Used to launch the various queues that interact with IPFS",
				ChildRequired: true,
				Children: map[string]cmd.Cmd{
					"ipns-entry": {
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
					"pin": {
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
					"file": {
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
					"key-creation": {
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
					"cluster": {
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
			"dfa": {
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
			"email-send": {
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
			"tns": {
				Blurb:         "run tns queues",
				Description:   "Allows running the various tns queue services",
				ChildRequired: true,
				Children: map[string]cmd.Cmd{
					"zone-creation": {
						Blurb:       "Zone creation queue",
						Description: "Listens to requests to create TNS zones",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							mqConnectionURL := cfg.RabbitMQ.URL
							qm, err := queue.Initialize(queue.ZoneCreationQueue, mqConnectionURL, false, true)
							if err != nil {
								log.Fatal(err)
							}
							if err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg); err != nil {
								log.Fatal(err)
							}
						},
					},
					"record-creation": {
						Blurb:       "record creation queue",
						Description: "Listens to requests to create TNS records",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							mqConnectionURL := cfg.RabbitMQ.URL
							qm, err := queue.Initialize(queue.RecordCreationQueue, mqConnectionURL, false, true)
							if err != nil {
								log.Fatal(err)
							}
							if err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg); err != nil {
								log.Fatal(err)
							}
						},
					},
				},
			},
		},
	},
	"migrate": {
		Blurb:       "run database migrations",
		Description: "Runs our initial database migrations, creating missing tables, etc..",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if _, err := database.Initialize(&cfg, database.Options{
				RunMigrations: true,
			}); err != nil {
				log.Fatal(err)
			}
		},
	},
	"migrate-insecure": {
		Hidden:      true,
		Blurb:       "run database migrations without SSL",
		Description: "Runs our initial database migrations, creating missing tables, etc.. without SSL",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if _, err := database.Initialize(&cfg, database.Options{
				RunMigrations:  true,
				SSLModeDisable: true,
			}); err != nil {
				log.Fatal(err)
			}
		},
	},
	"init": {
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
	"user": {
		Hidden:      true,
		Blurb:       "create a user",
		Description: "Create a Temporal user. Provide args as username, password, email. Do not use in production.",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if len(os.Args) < 5 {
				log.Fatal("insufficient fields provided")
			}
			d, err := database.Initialize(&cfg, database.Options{
				SSLModeDisable: true,
			})
			if err != nil {
				log.Fatal(err)
			}
			if _, err := models.NewUserManager(d.DB).NewUserAccount(
				os.Args[2], os.Args[3], os.Args[4], false,
			); err != nil {
				log.Fatal(err)
			}
		},
	},
	"admin": {
		Hidden:      true,
		Blurb:       "assign user as an admin",
		Description: "Assign an existing Temporal user as an administrator.",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if len(os.Args) < 3 {
				log.Fatal("no user provided")
			}
			d, err := database.Initialize(&cfg, database.Options{
				SSLModeDisable: true,
			})
			if err != nil {
				log.Fatal(err)
			}
			found, err := models.NewUserManager(d.DB).ToggleAdmin(os.Args[2])
			if err != nil {
				log.Fatal(err)
			}
			if !found {
				log.Fatalf("user %s not found", os.Args[2])
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
	var (
		peerAddr string
		isTns    bool
	)
	// check for tns client operation and load peer addr
	for _, v := range os.Args {
		if v == "tns" {
			isTns = true
		}
		if isTns && v == "client" {
			peerAddr = os.Getenv("PEER_ADDR")
			if peerAddr == "" {
				log.Fatal("PEER_ADDR env var is empty")
			}
		}
	}
	if isTns && peerAddr != "" {
		flags["peerAddr"] = peerAddr
	}
	fmt.Println(tCfg.APIKeys.ChainRider)
	// execute
	os.Exit(temporal.Run(*tCfg, flags, os.Args[1:]))
}
