package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/RTradeLtd/Temporal/mini"

	"github.com/RTradeLtd/Temporal/v2"
	"github.com/RTradeLtd/Temporal/api/v3"
	"github.com/RTradeLtd/Temporal/grpc-clients"
	"github.com/RTradeLtd/Temporal/log"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/cmd"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
	pbOrch "github.com/RTradeLtd/grpc/ipfs-orchestrator"
	pbLens "github.com/RTradeLtd/grpc/lens"
	pbSigner "github.com/RTradeLtd/grpc/temporal"
	"github.com/RTradeLtd/kaas"
	"github.com/jinzhu/gorm"
)

var (
	// Version denotes the tag of this build
	Version string

	closeMessage = "press CTRL+C to stop processing and close queue resources"
	certFile     = filepath.Join(os.Getenv("HOME"), "/certificates/api.pem")
	keyFile      = filepath.Join(os.Getenv("HOME"), "/certificates/api.key")
	tCfg         config.TemporalConfig
	db           *gorm.DB
	ctx          context.Context
	cancel       context.CancelFunc
	orch         pbOrch.ServiceClient
	lens         pbLens.IndexerAPIClient
	signer       pbSigner.SignerClient
	err          error
	logFilePath  = "/var/log/temporal/"
	dev          bool
)

var commands = map[string]cmd.Cmd{
	"api": {
		Blurb:       "start Temporal api server",
		Description: "Start the API service used to interact with Temporal. Run with DEBUG=true to enable debug messages.",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if cfg.LogDir == "" {
				logFilePath = logFilePath + "api_service.log"
			} else {
				logFilePath = cfg.LogDir + "api_service.log"
			}
			logger, err := log.NewLogger(logFilePath, dev)
			if err != nil {
				fmt.Println("failed to start logger ", err)
				os.Exit(1)
			}
			logger = logger.With("version", args["version"])
			service, err := v2.Initialize(&cfg, args["version"], os.Getenv("DEBUG") == "true", logger, lens, orch, signer)
			if err != nil {
				logger.Fatal(err)
			}

			port := os.Getenv("API_PORT")
			if port == "" {
				port = "6767"
			}
			quitChannel := make(chan os.Signal)
			signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
			go func() {
				fmt.Println(closeMessage)
				<-quitChannel
				service.Close()
				cancel()
			}()
			addr := fmt.Sprintf("%s:%s", args["listenAddress"], port)
			if args["certFilePath"] == "" || args["keyFilePath"] == "" {
				fmt.Println("TLS config incomplete - starting API service without TLS...")
				err = service.ListenAndServe(ctx, addr, nil)
			} else {
				fmt.Println("Starting API service with TLS...")
				err = service.ListenAndServe(ctx, addr, &v2.TLSConfig{
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
							if cfg.LogDir == "" {
								logFilePath = logFilePath + "ipns_consumer.log"
							} else {
								logFilePath = cfg.LogDir + "ipns_consumer.log"
							}
							logger, err := log.NewLogger(logFilePath, dev)
							if err != nil {
								fmt.Println("failed to start logger ", err)
								os.Exit(1)
							}
							qm, err := queue.New(queue.IpnsEntryQueue, cfg.RabbitMQ.URL, false, logger)
							if err != nil {
								fmt.Println("failed to start queue", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							waitGroup.Add(1)
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							if err := qm.ConsumeMessages(ctx, waitGroup, db, &cfg); err != nil {
								fmt.Println("failed to consume messages", err)
								os.Exit(1)
							}
							waitGroup.Wait()
						},
					},
					"pin": {
						Blurb:       "Pin addition queue",
						Description: "Listens to pin requests",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							if cfg.LogDir == "" {
								logFilePath = logFilePath + "pin_consumer.log"
							} else {
								logFilePath = cfg.LogDir + "pin_consumer.log"
							}
							logger, err := log.NewLogger(logFilePath, dev)
							if err != nil {
								fmt.Println("failed to start logger ", err)
								os.Exit(1)
							}
							qm, err := queue.New(queue.IpfsPinQueue, cfg.RabbitMQ.URL, false, logger)
							if err != nil {
								fmt.Println("failed to start queue", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							waitGroup.Add(1)
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							if err := qm.ConsumeMessages(ctx, waitGroup, db, &cfg); err != nil {
								fmt.Println("failed to consume messages", err)
								os.Exit(1)
							}
							waitGroup.Wait()
						},
					},
					"file": {
						Blurb:       "File upload queue",
						Description: "Listens to file upload requests. Only applies to advanced uploads",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							if cfg.LogDir == "" {
								logFilePath = logFilePath + "file_consumer.log"
							} else {
								logFilePath = cfg.LogDir + "file_consumer.log"
							}
							logger, err := log.NewLogger(logFilePath, dev)
							if err != nil {
								fmt.Println("failed to start logger ", err)
								os.Exit(1)
							}
							qm, err := queue.New(queue.IpfsFileQueue, cfg.RabbitMQ.URL, false, logger)
							if err != nil {
								fmt.Println("failed to start queue", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							waitGroup.Add(1)
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							if err := qm.ConsumeMessages(ctx, waitGroup, db, &cfg); err != nil {
								fmt.Println("failed to consume messages", err)
								os.Exit(1)
							}
							waitGroup.Wait()
						},
					},
					"key-creation": {
						Blurb:       "Key creation queue",
						Description: fmt.Sprintf("Listen to key creation requests.\nMessages to this queue are broadcasted to all nodes"),
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							if cfg.LogDir == "" {
								logFilePath = logFilePath + "key_consumer.log"
							} else {
								logFilePath = cfg.LogDir + "key_consumer.log"
							}
							logger, err := log.NewLogger(logFilePath, dev)
							if err != nil {
								fmt.Println("failed to start logger ", err)
								os.Exit(1)
							}
							qm, err := queue.New(queue.IpfsKeyCreationQueue, cfg.RabbitMQ.URL, false, logger)
							if err != nil {
								fmt.Println("failed to start queue", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							waitGroup.Add(1)
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							if err := qm.ConsumeMessages(ctx, waitGroup, db, &cfg); err != nil {
								fmt.Println("failed to consume messages", err)
								os.Exit(1)
							}
							waitGroup.Wait()
						},
					},
					"cluster": {
						Blurb:       "Cluster pin queue",
						Description: "Listens to requests to pin content to the cluster",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							if cfg.LogDir == "" {
								logFilePath = logFilePath + "cluster_pin_consumer.log"
							} else {
								logFilePath = cfg.LogDir + "cluster_pin_consumer.log"
							}
							logger, err := log.NewLogger(logFilePath, dev)
							if err != nil {
								fmt.Println("failed to start logger ", err)
								os.Exit(1)
							}
							qm, err := queue.New(queue.IpfsClusterPinQueue, cfg.RabbitMQ.URL, false, logger)
							if err != nil {
								fmt.Println("failed to start queue", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							waitGroup.Add(1)
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							if err := qm.ConsumeMessages(ctx, waitGroup, db, &cfg); err != nil {
								fmt.Println("failed to consume messages", err)
								os.Exit(1)
							}
							waitGroup.Wait()
						},
					},
				},
			},
			"dfa": {
				Blurb:       "Database file add queue",
				Description: "Listens to file uploads requests. Only applies to simple upload route",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					if cfg.LogDir == "" {
						logFilePath = logFilePath + "dfa_consumer.log"
					} else {
						logFilePath = cfg.LogDir + "dfa_consumer.log"
					}
					logger, err := log.NewLogger(logFilePath, dev)
					if err != nil {
						fmt.Println("failed to start logger ", err)
						os.Exit(1)
					}
					qm, err := queue.New(queue.DatabaseFileAddQueue, cfg.RabbitMQ.URL, false, logger)
					if err != nil {
						fmt.Println("failed to start queue", err)
						os.Exit(1)
					}
					quitChannel := make(chan os.Signal)
					signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
					waitGroup := &sync.WaitGroup{}
					waitGroup.Add(1)
					go func() {
						fmt.Println(closeMessage)
						<-quitChannel
						cancel()
					}()
					if err := qm.ConsumeMessages(ctx, waitGroup, db, &cfg); err != nil {
						fmt.Println("failed to consume messages", err)
						os.Exit(1)
					}
					waitGroup.Wait()
				},
			},
			"email-send": {
				Blurb:       "Email send queue",
				Description: "Listens to requests to send emails",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					if cfg.LogDir == "" {
						logFilePath = logFilePath + "email_consumer.log"
					} else {
						logFilePath = cfg.LogDir + "email_consumer.log"
					}
					logger, err := log.NewLogger(logFilePath, dev)
					if err != nil {
						fmt.Println("failed to start logger ", err)
						os.Exit(1)
					}
					qm, err := queue.New(queue.EmailSendQueue, cfg.RabbitMQ.URL, false, logger)
					if err != nil {
						fmt.Println("failed to start queue", err)
						os.Exit(1)
					}
					quitChannel := make(chan os.Signal)
					signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
					waitGroup := &sync.WaitGroup{}
					waitGroup.Add(1)
					go func() {
						fmt.Println(closeMessage)
						<-quitChannel
						cancel()
					}()
					if err := qm.ConsumeMessages(ctx, waitGroup, db, &cfg); err != nil {
						fmt.Println("failed to consume messages", err)
						os.Exit(1)
					}
					waitGroup.Wait()
				},
			},
		},
	},
	"krab": {
		Blurb:       "runs the krab service",
		Description: "Runs the krab grpc server, allowing for secure private key management",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if err := kaas.NewServer(cfg.Endpoints.Krab.URL, "tcp", &cfg); err != nil {
				fmt.Println("failed to start krab server", err)
				os.Exit(1)
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
				fmt.Println("failed to perform insecure migration", err)
				os.Exit(1)
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
				fmt.Println("CONFIG_DAG is not set")
				os.Exit(1)
			}
			if err := config.GenerateConfig(configDag); err != nil {
				fmt.Println("failed to generate default config template", err)
				os.Exit(1)
			}
		},
	},
	"user": {
		Hidden:      true,
		Blurb:       "create a user",
		Description: "Create a Temporal user. Provide args as username, password, email. Do not use in production.",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if len(args) < 3 {
				fmt.Println("insufficient arguments provided")
				os.Exit(1)
			}
			d, err := database.Initialize(&cfg, database.Options{
				SSLModeDisable: true,
			})
			if err != nil {
				fmt.Println("failed to initialize database", err)
				os.Exit(1)
			}
			if _, err := models.NewUserManager(d.DB).NewUserAccount(
				args["user"], args["pass"], args["email"],
			); err != nil {
				fmt.Println("failed to create user account", err)
				os.Exit(1)
			}
		},
	},
	"admin": {
		Hidden:      true,
		Blurb:       "assign user as an admin",
		Description: "Assign an existing Temporal user as an administrator.",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if args["dbAdmin"] == "" {
				fmt.Println("dbAdmin flag not provided")
				os.Exit(1)
			}
			d, err := database.Initialize(&cfg, database.Options{
				SSLModeDisable: true,
			})
			if err != nil {
				fmt.Println("failed to initialize database", err)
				os.Exit(1)
			}
			found, err := models.NewUserManager(d.DB).ToggleAdmin(args["dbAdmin"])
			if err != nil {
				fmt.Println("failed to tag user as admin", err)
				os.Exit(1)
			}
			if !found {
				fmt.Println("failed to find user", err)
				os.Exit(1)
			}
		},
	},
	"v3": {
		Blurb:         "experimental Temporal V3 API",
		Hidden:        true,
		PreRun:        true,
		ChildRequired: true,
		Children: map[string]cmd.Cmd{
			"proxy": {
				Blurb: "run a RESTful proxy for the Temporal V3 API",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					if len(os.Args) < 4 {
						fmt.Println("no address provided")
						os.Exit(1)
					}
					if err := v3.REST(os.Args[3]); err != nil {
						fmt.Println("error occurred during proxy initialization", err)
						os.Exit(1)
					}
				},
			},
			"server": {
				Blurb: "run the Temporal V3 API server",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					if len(os.Args) < 4 {
						fmt.Println("no address provided")
						os.Exit(1)
					}
					s := v3.New()
					if err := s.Run(context.Background(), os.Args[3]); err != nil {
						fmt.Println("error starting v3 API server", err)
						os.Exit(1)
					}
				},
			},
		},
	},
	"make-bucket": {
		Hidden:      true,
		Blurb:       "used to create a minio bucket, run against localhost only",
		Description: "Allows the creation of buckets with minio, useful for the initial setup of temporal",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			mm, err := mini.NewMinioManager(
				cfg.MINIO.Connection.IP+":"+cfg.MINIO.Connection.Port,
				cfg.MINIO.AccessKey, cfg.MINIO.SecretKey, false)
			if err != nil {
				fmt.Println("failed to initialize minio manager")
				os.Exit(1)
			}
			if err = mm.MakeBucket(args); err != nil {
				fmt.Println("failed to create bucket")
				os.Exit(1)
			}
		},
	},
	"migrate": {
		Blurb:       "run database migrations",
		Description: "Runs our initial database migrations, creating missing tables, etc..",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if _, err := database.Initialize(&cfg, database.Options{
				RunMigrations: true,
			}); err != nil {
				fmt.Println("failed to perform secure migration", err)
				os.Exit(1)
			}
		},
	},
}

func main() {
	if Version == "" {
		Version = "latest"
	}

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
	logger, err := log.NewLogger("stdout", false)
	if err != nil {
		fmt.Println("failed to initialize logger")
		os.Exit(1)
	}
	// load config
	configDag := os.Getenv("CONFIG_DAG")
	if configDag == "" {
		logger.Fatal("CONFIG_DAG is not set")
	}
	tCfg, err := config.LoadConfig(configDag)
	if err != nil {
		logger.Fatal(err)
	}
	if initDB := os.Getenv("INIT_DB"); strings.ToLower(initDB) == "true" {
		sslDisabled := os.Getenv("SSL_MODE_DISABLE") == "true"
		db, err = database.OpenDBConnection(database.DBOptions{
			User:           tCfg.Database.Username,
			Password:       tCfg.Database.Password,
			Address:        tCfg.Database.URL,
			Port:           tCfg.Database.Port,
			SSLModeDisable: sslDisabled,
		})
		if err != nil {
			logger.Fatal(err)
		}
	}
	// initialize global context
	ctx, cancel = context.WithCancel(context.Background())
	// load arguments
	flags := map[string]string{
		"configDag":     configDag,
		"certFilePath":  tCfg.API.Connection.Certificates.CertPath,
		"keyFilePath":   tCfg.API.Connection.Certificates.KeyPath,
		"listenAddress": tCfg.API.Connection.ListenAddress,
		"dbPass":        tCfg.Database.Password,
		"dbURL":         tCfg.Database.URL,
		"dbUser":        tCfg.Database.Username,
		"version":       Version,
	}
	switch os.Args[1] {
	case "user":
		flags["user"] = os.Args[2]
		flags["pass"] = os.Args[3]
		flags["email"] = os.Args[4]
	case "admin":
		flags["dbAdmin"] = os.Args[2]
	case "api":
		lensClient, err := clients.NewLensClient(tCfg.Endpoints)
		if err != nil {
			logger.Fatal(err)
		}
		defer lensClient.Close()
		lens = lensClient
		orchClient, err := clients.NewOcrhestratorClient(tCfg.Orchestrator)
		if err != nil {
			logger.Fatal(err)
		}
		defer orchClient.Close()
		orch = orchClient
		signerClient, err := clients.NewSignerClient(tCfg, os.Getenv("SSL_MODE_DISABLE") == "true")
		if err != nil {
			logger.Fatal(err)
		}
		defer signerClient.Close()
		signer = signerClient
	case "make-bucket", "make-bucket-insecure":
		flags["name"] = os.Args[2]
	}
	fmt.Println(tCfg.APIKeys.ChainRider)
	// execute
	os.Exit(temporal.Run(*tCfg, flags, os.Args[1:]))
}
