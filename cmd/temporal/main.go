package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/bobheadxi/zapx"
	"go.uber.org/zap"

	v2 "github.com/RTradeLtd/Temporal/api/v2"
	clients "github.com/RTradeLtd/Temporal/grpc-clients"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/cmd/v2"
	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2"
	"github.com/RTradeLtd/database/v2/models"
	pbLens "github.com/RTradeLtd/grpc/lensv2"
	pbOrch "github.com/RTradeLtd/grpc/nexus"
	pbSigner "github.com/RTradeLtd/grpc/pay"
	"github.com/RTradeLtd/kaas/v2"
	pbBchWallet "github.com/gcash/bchwallet/rpc/walletrpc"
	"github.com/jinzhu/gorm"
)

// Version denotes the tag of this build
var Version string

const (
	closeMessage   = "press CTRL+C to stop processing and close queue resources"
	defaultLogPath = "/var/log/temporal/"
)

// globals
var (
	ctx       context.Context
	cancel    context.CancelFunc
	orch      pbOrch.ServiceClient
	lens      pbLens.LensV2Client
	signer    pbSigner.SignerClient
	bchWallet pbBchWallet.WalletServiceClient
)

// command-line flags
var (
	devMode    *bool
	debug      *bool
	configPath *string
	dbNoSSL    *bool
	dbMigrate  *bool
	apiPort    *string
)

func baseFlagSet() *flag.FlagSet {
	var f = flag.NewFlagSet("", flag.ExitOnError)

	// basic flags
	devMode = f.Bool("dev", false,
		"toggle dev mode")
	debug = f.Bool("debug", false,
		"toggle debug mode")
	configPath = f.String("config", os.Getenv("CONFIG_DAG"),
		"path to Temporal configuration")

	// db configuration
	dbNoSSL = f.Bool("db.no_ssl", false,
		"toggle SSL connection with database")
	dbMigrate = f.Bool("db.migrate", false,
		"toggle whether a database migration should occur")

	// api configuration
	apiPort = f.String("api.port", "6767",
		"set port to expose API on")

	return f
}

func logPath(base, file string) (logPath string) {
	if base == "" {
		logPath = filepath.Join(base, file)
	} else {
		logPath = filepath.Join(base, file)
	}
	return
}

func newDB(cfg config.TemporalConfig, noSSL bool) (*gorm.DB, error) {
	dbm, err := database.New(&cfg, database.Options{SSLModeDisable: noSSL})
	if err != nil {
		return nil, err
	}
	return dbm.DB, nil
}

func initClients(l *zap.SugaredLogger, cfg *config.TemporalConfig) (closers []func()) {
	closers = make([]func(), 0)
	if lens == nil {
		client, err := clients.NewLensClient(cfg.Services)
		if err != nil {
			l.Fatal(err)
		}
		closers = append(closers, client.Close)
		lens = client
	}
	if orch == nil {
		client, err := clients.NewOcrhestratorClient(cfg.Nexus)
		if err != nil {
			l.Fatal(err)
		}
		closers = append(closers, client.Close)
		orch = client
	}
	if signer == nil {
		client, err := clients.NewSignerClient(cfg)
		if err != nil {
			l.Fatal(err)
		}
		closers = append(closers, client.Close)
		signer = client
	}
	if bchWallet == nil {
		client, err := clients.NewBchWalletClient(cfg.Services)
		if err != nil {
			l.Fatal(err)
		}
		closers = append(closers, client.Close)
		bchWallet = client
	}
	return
}

var commands = map[string]cmd.Cmd{
	"api": {
		Blurb:       "start Temporal api server",
		Description: "Start the API service used to interact with Temporal. Run with DEBUG=true to enable debug messages.",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			logger, err := zapx.New(logPath(cfg.LogDir, "api_service.log"), *devMode)
			if err != nil {
				fmt.Println("failed to start logger ", err)
				os.Exit(1)
			}
			l := logger.Sugar().With("version", args["version"])

			// init clients and clean up if necessary
			var closers = initClients(l, &cfg)
			if closers != nil {
				defer func() {
					for _, c := range closers {
						c()
					}
				}()
			}
			clients := v2.Clients{
				Lens:      lens,
				Orch:      orch,
				Signer:    signer,
				BchWallet: bchWallet,
			}
			// init api service
			service, err := v2.Initialize(
				ctx,
				&cfg,
				args["version"],
				v2.Options{DebugLogging: *debug, DevMode: *devMode},
				clients,
				l,
			)
			if err != nil {
				l.Fatal(err)
			}

			// set up clean interrupt
			quitChannel := make(chan os.Signal)
			signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
			go func() {
				fmt.Println(closeMessage)
				<-quitChannel
				cancel()
				service.Close()
			}()

			// go!
			var addr = fmt.Sprintf("%s:%s", args["listenAddress"], *apiPort)
			var (
				cert string
				key  string
			)
			if args["certFilePath"] == "" || args["keyFilePath"] == "" {
				fmt.Println("TLS config incomplete - starting API service without TLS...")
				err = service.ListenAndServe(ctx, addr, nil)
			} else {
				if cert, err = filepath.Abs(args["certFilePath"]); err != nil {
					fmt.Println("certFilePath:", err)
					os.Exit(1)
				}
				if key, err = filepath.Abs(args["keyFilePath"]); err != nil {
					fmt.Println("keyFilePath:", err)
					os.Exit(1)
				}
				fmt.Println("Starting API service with TLS...")
				err = service.ListenAndServe(ctx, addr, &v2.TLSConfig{
					CertFile: cert,
					KeyFile:  key,
				})
			}
			if err != nil {
				fmt.Printf("API service execution failed: %s\n", err.Error())
				fmt.Println("Refer to the logs for more details")
				os.Exit(1)
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
							logger, err := zapx.New(logPath(cfg.LogDir, "ipns_consumer.log"), *devMode)
							if err != nil {
								fmt.Println("failed to start logger ", err)
								os.Exit(1)
							}
							l := logger.Named("ipns_consumer").Sugar()
							db, err := newDB(cfg, *dbNoSSL)
							if err != nil {
								fmt.Println("failed to start db", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							for {
								qm, err := queue.New(queue.IpnsEntryQueue, cfg.RabbitMQ.URL, false, *devMode, &cfg, l)
								if err != nil {
									fmt.Println("failed to start queue", err)
									os.Exit(1)
								}
								waitGroup.Add(1)
								err = qm.ConsumeMessages(ctx, waitGroup, db, &cfg)
								if err != nil && err.Error() != queue.ErrReconnect {
									fmt.Println("failed to consume messages", err)
									os.Exit(1)
								} else if err != nil && err.Error() == queue.ErrReconnect {
									continue
								}
								// this will only be true if we had a graceful exit to the queue process, aka CTRL+C
								if err == nil {
									break
								}
							}
							waitGroup.Wait()
						},
					},
					"pin": {
						Blurb:       "Pin addition queue",
						Description: "Listens to pin requests",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							logger, err := zapx.New(logPath(cfg.LogDir, "pin_consumer.log"), *devMode)
							if err != nil {
								fmt.Println("failed to start logger ", err)
								os.Exit(1)
							}
							l := logger.Named("pin_consumer").Sugar()

							db, err := newDB(cfg, *dbNoSSL)
							if err != nil {
								fmt.Println("failed to start db", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							for {
								qm, err := queue.New(queue.IpfsPinQueue, cfg.RabbitMQ.URL, false, *devMode, &cfg, l)
								if err != nil {
									fmt.Println("failed to start queue", err)
									os.Exit(1)
								}
								waitGroup.Add(1)
								err = qm.ConsumeMessages(ctx, waitGroup, db, &cfg)
								if err != nil && err.Error() != queue.ErrReconnect {
									fmt.Println("failed to consume messages", err)
									os.Exit(1)
								} else if err != nil && err.Error() == queue.ErrReconnect {
									continue
								}
								// this will only be true if we had a graceful exit to the queue process, aka CTRL+C
								if err == nil {
									break
								}
							}
							waitGroup.Wait()
						},
					},
					"key-creation": {
						Blurb:       "Key creation queue",
						Description: fmt.Sprintf("Listen to key creation requests.\nMessages to this queue are broadcasted to all nodes"),
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							logger, err := zapx.New(logPath(cfg.LogDir, "key_consumer.log"), *devMode)
							if err != nil {
								fmt.Println("failed to start logger ", err)
								os.Exit(1)
							}
							l := logger.Named("key_consumer").Sugar()

							db, err := newDB(cfg, *dbNoSSL)
							if err != nil {
								fmt.Println("failed to start db", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							for {
								qm, err := queue.New(queue.IpfsKeyCreationQueue, cfg.RabbitMQ.URL, false, *devMode, &cfg, l)
								if err != nil {
									fmt.Println("failed to start queue", err)
									os.Exit(1)
								}
								waitGroup.Add(1)
								err = qm.ConsumeMessages(ctx, waitGroup, db, &cfg)
								if err != nil && err.Error() != queue.ErrReconnect {
									fmt.Println("failed to consume messages", err)
									os.Exit(1)
								} else if err != nil && err.Error() == queue.ErrReconnect {
									continue
								}
								// this will only be true if we had a graceful exit to the queue process, aka CTRL+C
								if err == nil {
									break
								}
							}
							waitGroup.Wait()
						},
					},
					"cluster": {
						Blurb:       "Cluster pin queue",
						Description: "Listens to requests to pin content to the cluster",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							logger, err := zapx.New(logPath(cfg.LogDir, "cluster_pin_consumer.log"), *devMode)
							if err != nil {
								fmt.Println("failed to start logger ", err)
								os.Exit(1)
							}
							l := logger.Named("cluster_pin_consumer").Sugar()

							db, err := newDB(cfg, *dbNoSSL)
							if err != nil {
								fmt.Println("failed to start db", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							for {
								qm, err := queue.New(queue.IpfsClusterPinQueue, cfg.RabbitMQ.URL, false, *devMode, &cfg, l)
								if err != nil {
									fmt.Println("failed to start queue", err)
									os.Exit(1)
								}
								waitGroup.Add(1)
								err = qm.ConsumeMessages(ctx, waitGroup, db, &cfg)
								if err != nil && err.Error() != queue.ErrReconnect {
									fmt.Println("failed to consume messages", err)
									os.Exit(1)
								} else if err != nil && err.Error() == queue.ErrReconnect {
									continue
								}
								// this will only be true if we had a graceful exit to the queue process, aka CTRL+C
								if err == nil {
									break
								}
							}
							waitGroup.Wait()
						},
					},
				},
			},
			"email-send": {
				Blurb:       "Email send queue",
				Description: "Listens to requests to send emails",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					logger, err := zapx.New(logPath(cfg.LogDir, "email_consumer.log"), *devMode)
					if err != nil {
						fmt.Println("failed to start logger ", err)
						os.Exit(1)
					}
					l := logger.Named("email_consumer").Sugar()

					db, err := newDB(cfg, *dbNoSSL)
					if err != nil {
						fmt.Println("failed to start db", err)
						os.Exit(1)
					}
					quitChannel := make(chan os.Signal)
					signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
					waitGroup := &sync.WaitGroup{}
					go func() {
						fmt.Println(closeMessage)
						<-quitChannel
						cancel()
					}()
					for {
						qm, err := queue.New(queue.EmailSendQueue, cfg.RabbitMQ.URL, false, *devMode, &cfg, l)
						if err != nil {
							fmt.Println("failed to start queue", err)
							os.Exit(1)
						}
						waitGroup.Add(1)
						err = qm.ConsumeMessages(ctx, waitGroup, db, &cfg)
						if err != nil && err.Error() != queue.ErrReconnect {
							fmt.Println("failed to consume messages", err)
							os.Exit(1)
						} else if err != nil && err.Error() == queue.ErrReconnect {
							continue
						}
						// this will only be true if we had a graceful exit to the queue process, aka CTRL+C
						if err == nil {
							break
						}
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
			if err := kaas.NewServer(cfg.Services.Krab.URL, "tcp", &cfg); err != nil {
				fmt.Println("failed to start krab server", err)
				os.Exit(1)
			}
		},
	},
	"init": {
		PreRun:      true,
		Blurb:       "initialize blank Temporal configuration",
		Description: "Initializes a blank Temporal configuration template at path provided by the '-config' flag",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			println("generating config at", *configPath)
			if err := config.GenerateConfig(*configPath); err != nil {
				fmt.Println("failed to generate default config template", err)
				os.Exit(1)
			}
		},
	},
	"user": {
		Hidden:      true,
		Blurb:       "create a user",
		Description: "Create a Temporal user. Provide args as username, password, email. Do not use in production.",
		Args:        []string{"user", "pass", "email"},
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			fmt.Printf("creating user '%s' (%s)...\n", args["user"], args["email"])
			d, err := database.New(&cfg, database.Options{
				SSLModeDisable: *dbNoSSL,
				RunMigrations:  *dbMigrate,
			})
			if err != nil {
				fmt.Println("failed to initialize database connection", err)
				os.Exit(1)
			}
			// create user account
			if _, err := models.NewUserManager(d.DB).NewUserAccount(
				args["user"], args["pass"], args["email"],
			); err != nil {
				fmt.Println("failed to create user account", err)
				os.Exit(1)
			}
			// add credits
			if _, err := models.NewUserManager(d.DB).AddCredits(args["user"], 99999999); err != nil {
				fmt.Println("failed to grant credits to user account", err)
				os.Exit(1)
			}
			// generate email activation token
			userModel, err := models.NewUserManager(d.DB).GenerateEmailVerificationToken(args["user"])
			if err != nil {
				fmt.Println("failed to generate email verification token", err)
				os.Exit(1)
			}
			// activate email
			if _, err := models.NewUserManager(d.DB).ValidateEmailVerificationToken(args["user"], userModel.EmailVerificationToken); err != nil {
				fmt.Println("failed to activate email", err)
				os.Exit(1)
			}
		},
	},
	"admin": {
		Hidden:      true,
		Blurb:       "assign user as an admin",
		Description: "Assign an existing Temporal user as an administrator.",
		Args:        []string{"dbAdmin"},
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if args["dbAdmin"] == "" {
				fmt.Println("dbAdmin flag not provided")
				os.Exit(1)
			}
			d, err := database.New(&cfg, database.Options{
				SSLModeDisable: *dbNoSSL,
				RunMigrations:  *dbMigrate,
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
	"migrate": {
		Blurb:       "run database migrations",
		Description: "Runs our initial database migrations, creating missing tables, etc. Not affected by --db.migrate",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if _, err := database.New(&cfg, database.Options{
				SSLModeDisable: *dbNoSSL,
				RunMigrations:  true,
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

	// initialize global context
	ctx, cancel = context.WithCancel(context.Background())

	// create app
	temporal := cmd.New(commands, cmd.Config{
		Name:     "Temporal",
		ExecName: "temporal",
		Version:  Version,
		Desc:     "Temporal is an easy-to-use interface into distributed and decentralized storage technologies for personal and enterprise use cases.",
		Options:  baseFlagSet(),
	})

	// run no-config commands, exit if command was run
	if exit := temporal.PreRun(nil, os.Args[1:]); exit == cmd.CodeOK {
		os.Exit(0)
	}

	// load config
	tCfg, err := config.LoadConfig(*configPath)
	if err != nil {
		println("failed to load config at", *configPath)
		os.Exit(1)
	}

	// load arguments
	flags := map[string]string{
		"certFilePath":  tCfg.API.Connection.Certificates.CertPath,
		"keyFilePath":   tCfg.API.Connection.Certificates.KeyPath,
		"listenAddress": tCfg.API.Connection.ListenAddress,
		"dbPass":        tCfg.Database.Password,
		"dbURL":         tCfg.Database.URL,
		"dbUser":        tCfg.Database.Username,
		"version":       Version,
	}

	// execute
	os.Exit(temporal.Run(*tCfg, flags, os.Args[1:]))
}
