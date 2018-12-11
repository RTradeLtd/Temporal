package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/RTradeLtd/Temporal/grpc-clients"

	"github.com/RTradeLtd/Temporal/api/v2"
	"github.com/RTradeLtd/Temporal/api/v3"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/cmd"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/kaas"
	"github.com/jinzhu/gorm"
)

var (
	// Version denotes the tag of this build
	Version      string
	closeMessage = "press CTRL+C to stop processing and close queue resources"
	certFile     = filepath.Join(os.Getenv("HOME"), "/certificates/api.pem")
	keyFile      = filepath.Join(os.Getenv("HOME"), "/certificates/api.key")
	tCfg         config.TemporalConfig
	db           *gorm.DB
	ctx          context.Context
	cancel       context.CancelFunc
	logFilePath  = "/var/log/temporal/"
)

var commands = map[string]cmd.Cmd{
	"api": {
		Blurb:       "start Temporal api server",
		Description: "Start the API service used to interact with Temporal. Run with DEBUG=true to enable debug messages.",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			lc, err := clients.NewLensClient(cfg.Endpoints)
			if err != nil {
				log.Fatal(err)
			}
			defer lc.Close()
			service, err := v2.Initialize(&cfg, os.Getenv("DEBUG") == "true", lc)
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
				err = service.ListenAndServe(addr, &v2.TLSConfig{
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
							qm, err := queue.New(queue.IpnsEntryQueue, mqConnectionURL, false, logFilePath)
							if err != nil {
								log.Fatal(err)
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
								log.Fatal(err)
							}
							waitGroup.Wait()
						},
					},
					"pin": {
						Blurb:       "Pin addition queue",
						Description: "Listens to pin requests",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							mqConnectionURL := cfg.RabbitMQ.URL
							qm, err := queue.New(queue.IpfsPinQueue, mqConnectionURL, false, logFilePath)
							if err != nil {
								log.Fatal(err)
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
								log.Fatal(err)
							}
							waitGroup.Wait()
						},
					},
					"file": {
						Blurb:       "File upload queue",
						Description: "Listens to file upload requests. Only applies to advanced uploads",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							mqConnectionURL := cfg.RabbitMQ.URL
							qm, err := queue.New(queue.IpfsFileQueue, mqConnectionURL, false, logFilePath)
							if err != nil {
								log.Fatal(err)
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
								log.Fatal(err)
							}
							waitGroup.Wait()
						},
					},
					"key-creation": {
						Blurb:       "Key creation queue",
						Description: fmt.Sprintf("Listen to key creation requests.\nMessages to this queue are broadcasted to all nodes"),
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							mqConnectionURL := cfg.RabbitMQ.URL
							qm, err := queue.New(queue.IpfsKeyCreationQueue, mqConnectionURL, false, logFilePath)
							if err != nil {
								log.Fatal(err)
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
								log.Fatal(err)
							}
							waitGroup.Wait()
						},
					},
					"cluster": {
						Blurb:       "Cluster pin queue",
						Description: "Listens to requests to pin content to the cluster",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							mqConnectionURL := cfg.RabbitMQ.URL
							qm, err := queue.New(queue.IpfsClusterPinQueue, mqConnectionURL, false, logFilePath)
							if err != nil {
								log.Fatal(err)
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
								log.Fatal(err)
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
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.New(queue.DatabaseFileAddQueue, mqConnectionURL, false, logFilePath)
					if err != nil {
						log.Fatal(err)
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
						log.Fatal(err)
					}
					waitGroup.Wait()
				},
			},
			"email-send": {
				Blurb:       "Email send queue",
				Description: "Listens to requests to send emails",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					mqConnectionURL := cfg.RabbitMQ.URL
					qm, err := queue.New(queue.EmailSendQueue, mqConnectionURL, false, logFilePath)
					if err != nil {
						log.Fatal(err)
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
						log.Fatal(err)
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
				log.Fatal(err)
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
			if len(args) < 3 {
				log.Fatal("insufficient fields provided")
			}
			d, err := database.Initialize(&cfg, database.Options{
				SSLModeDisable: true,
			})
			if err != nil {
				log.Fatal(err)
			}
			if _, err := models.NewUserManager(d.DB).NewUserAccount(
				args["user"], args["pass"], args["email"], false,
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
			if args["dbAdmin"] == "" {
				log.Fatal("no user provided through command line argument")
			}
			d, err := database.Initialize(&cfg, database.Options{
				SSLModeDisable: true,
			})
			if err != nil {
				log.Fatal(err)
			}
			found, err := models.NewUserManager(d.DB).ToggleAdmin(args["dbAdmin"])
			if err != nil {
				log.Fatal(err)
			}
			if !found {
				log.Fatalf("user %s not found", args["dbAdmin"])
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
						log.Fatal("no address provided")
					}
					log.Fatal(v3.REST(os.Args[3]))
				},
			},
			"server": {
				Blurb: "run the Temporal V3 API server",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					if len(os.Args) < 4 {
						log.Fatal("no address provided")
					}
					s := v3.New()
					log.Fatal(s.Run(context.Background(), os.Args[3]))
				},
			},
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
			log.Fatal(err)
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
	}
	if os.Args[1] == "user" {
		flags["user"] = os.Args[2]
		flags["pass"] = os.Args[3]
		flags["email"] = os.Args[4]
	}
	if os.Args[1] == "admin" {
		flags["dbAdmin"] = os.Args[2]
	}
	fmt.Println(tCfg.APIKeys.ChainRider)
	// execute
	os.Exit(temporal.Run(*tCfg, flags, os.Args[1:]))
}
