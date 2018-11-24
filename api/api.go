// Package api is the main package for Temporal's
// http api
package api

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/RTradeLtd/rtfs/krab"

	"github.com/RTradeLtd/ChainRider-Go/dash"
	clients "github.com/RTradeLtd/Temporal/grpc-clients"
	"github.com/RTradeLtd/rtfs"

	limit "github.com/aviddiviner/gin-limit"
	helmet "github.com/danielkov/gin-helmet"
	"github.com/sirupsen/logrus"

	"github.com/RTradeLtd/config"
	xss "github.com/dvwright/xss-mw"
	stats "github.com/semihalev/gin-stats"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	"github.com/RTradeLtd/Temporal/api/middleware"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	xssMdlwr xss.XssMw
	dev      = true
)

// API is our API service
type API struct {
	ipfs    rtfs.Manager
	keys    *rtfs.KeystoreManager
	r       *gin.Engine
	cfg     *config.TemporalConfig
	dbm     *database.Manager
	um      *models.UserManager
	im      *models.IpnsManager
	pm      *models.PaymentManager
	dm      *models.DropManager
	ue      *models.EncryptedUploadManager
	zm      *models.ZoneManager
	rm      *models.RecordManager
	nm      *models.IPFSNetworkManager
	l       *log.Logger
	signer  *clients.SignerClient
	orch    *clients.IPFSOrchestratorClient
	lc      *clients.LensClient
	dc      *dash.Client
	service string
}

// Initialize is used ot initialize our API service. debug = true is useful
// for debugging database issues.
func Initialize(cfg *config.TemporalConfig, debug bool) (*API, error) {
	var (
		err    error
		router = gin.Default()
		p      = ginprometheus.NewPrometheus("gin")
	)

	// set up prometheus monitoring
	p.SetListenAddress(fmt.Sprintf("%s:6768", cfg.API.Connection.ListenAddress))
	p.Use(router)

	// open log file
	logfile, err := os.OpenFile(cfg.API.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %s", err)
	}
	kb, err := krab.NewKrab(krab.Opts{Passphrase: cfg.IPFS.KrabPassword, DSPath: cfg.IPFS.KeystorePath})
	if err != nil {
		return nil, err
	}
	keystore, err := rtfs.NewKeystoreManager(kb)
	if err != nil {
		return nil, err
	}
	im, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		keystore,
		time.Minute*10,
	)
	if err != nil {
		return nil, err
	}
	// set up API struct
	api, err := new(cfg, router, im, keystore, debug, logfile)
	if err != nil {
		return nil, err
	}

	// init routes
	api.setupRoutes()
	api.LogInfo("api initialization successful")

	// return our configured API service
	return api, nil
}

func new(cfg *config.TemporalConfig, router *gin.Engine, ipfs rtfs.Manager, keystore *rtfs.KeystoreManager, debug bool, out io.Writer) (*API, error) {
	var (
		logger = log.New()
		dbm    *database.Manager
		err    error
	)

	// set up logger
	logger.Out = out
	logger.Info("logger initialized")

	// enable debug mode if requested
	if debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	// set up database manager
	dbm, err = database.Initialize(cfg, database.Options{LogMode: debug})
	if err != nil {
		logger.Warnf("failed to connect to database: %s", err.Error())
		logger.Warnf("failed to connect to database with secure connection - attempting insecure connection...")
		dbm, err = database.Initialize(cfg, database.Options{
			LogMode:        debug,
			SSLModeDisable: true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database with insecure connection: %s", err.Error())
		}
		logger.Warnf("insecure database connection established")
	} else {
		logger.Info("secure database connection established")
	}

	signer, err := clients.NewSignerClient(cfg, os.Getenv("MODE") == "development")
	if err != nil {
		return nil, err
	}

	orch, err := clients.NewOcrhestratorClient(cfg.Orchestrator, os.Getenv("MODE") == "development")
	if err != nil {
		return nil, err
	}
	var networkVersion string
	if dev {
		networkVersion = "testnet"
	} else {
		networkVersion = "main"
	}
	dc, err := dash.NewClient(&dash.ConfigOpts{
		APIVersion:      "v1",
		DigitalCurrency: "dash",
		//TODO: change to main before production release
		Blockchain: networkVersion,
		Token:      cfg.APIKeys.ChainRider,
	})
	if err != nil {
		return nil, err
	}
	lensClient, err := clients.NewLensClient(cfg.Endpoints)
	if err != nil {
		return nil, err
	}
	return &API{
		ipfs:    ipfs,
		keys:    keystore,
		cfg:     cfg,
		service: "api",
		r:       router,
		l:       logger,
		dbm:     dbm,
		um:      models.NewUserManager(dbm.DB),
		im:      models.NewIPNSManager(dbm.DB),
		pm:      models.NewPaymentManager(dbm.DB),
		dm:      models.NewDropManager(dbm.DB),
		ue:      models.NewEncryptedUploadManager(dbm.DB),
		lc:      lensClient,
		signer:  signer,
		orch:    orch,
		dc:      dc,
		zm:      models.NewZoneManager(dbm.DB),
		rm:      models.NewRecordManager(dbm.DB),
		nm:      models.NewHostedIPFSNetworkManager(dbm.DB),
	}, nil
}

// Close releases API resources
func (api *API) Close() {
	api.lc.Close()
	api.signer.Close()
	api.orch.Close()
}

// TLSConfig is used to enable TLS on the API service
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// ListenAndServe spins up the API server
func (api *API) ListenAndServe(addr string, tls *TLSConfig) error {
	if tls != nil {
		return api.r.RunTLS(addr, tls.CertFile, tls.KeyFile)
	}
	return api.r.Run(addr)
}

// setupRoutes is used to setup all of our api routes
func (api *API) setupRoutes() error {
	// setup the connection limit
	connLimit, err := strconv.Atoi(api.cfg.API.Connection.Limit)
	if err != nil {
		return err
	}
	// set up defaults
	api.r.Use(
		xssMdlwr.RemoveXss(),
		limit.MaxAllowed(connLimit),
		helmet.NoSniff(),
		middleware.CORSMiddleware(),
		stats.RequestStats())

	// set up middleware
	ginjwt := middleware.JwtConfigGenerate(api.cfg.API.JwtKey, api.dbm.DB, api.l)
	authware := []gin.HandlerFunc{
		ginjwt.MiddlewareFunc(),
		middleware.APIRestrictionMiddleware(api.dbm.DB),
	}

	// V1 API
	v1 := api.r.Group("/api/v1")

	// system checks used to verify the integrity of our services
	systemChecks := v1.Group("/systems")
	{
		systemChecks.GET("/check", api.SystemsCheck)
	}

	// authless account recovery routes
	forgot := v1.Group("/forgot")
	{
		forgot.POST("/username", api.forgotUserName)
		forgot.POST("/password", api.resetPassword)
	}

	// authentication
	auth := v1.Group("/auth")
	{
		auth.POST("/register", api.registerUserAccount)
		auth.POST("/login", ginjwt.LoginHandler)
	}

	// statistics
	statistics := v1.Group("/statistics").Use(authware...)
	{
		statistics.GET("/stats", api.getStats)
	}

	// lens search engine
	lens := v1.Group("/lens")
	{
		// allow anyone to index
		lens.POST("/index")
		// only allow registered users to search
		lens.POST("/search", api.submitSearchRequest).Use(authware...)
	}

	// payments
	payments := v1.Group("/payments", authware...)
	{
		dash := payments.Group("/create")
		{
			dash.POST("/dash", api.CreateDashPayment)
		}
		payments.POST("/request", api.RequestSignedPaymentMessage)
		payments.POST("/confirm", api.ConfirmPayment)
		deposit := payments.Group("/deposit")
		{
			deposit.GET("/address/:type", api.GetDepositAddress)
		}
	}

	// accounts
	account := v1.Group("/account", authware...)
	{
		account.POST("/rekt", api.selfRekt)
		token := account.Group("/token")
		{
			token.GET("/username", api.getUserFromToken)
		}
		airdrop := account.Group("/airdrop")
		{
			airdrop.POST("/register", api.registerAirDrop)
		}
		password := account.Group("/password")
		{
			password.POST("/change", api.changeAccountPassword)
		}
		key := account.Group("/key")
		{
			key.GET("/export/:name", api.exportKey)
			ipfs := key.Group("/ipfs")
			{
				ipfs.GET("/get", api.getIPFSKeyNamesForAuthUser)
				ipfs.POST("/new", api.createIPFSKey)
			}
		}
		credits := account.Group("/credits")
		{
			credits.GET("/available", api.getCredits)
		}
		email := account.Group("/email")
		{
			email.POST("/forgot", api.forgotEmail)
			token := email.Group("/token")
			{
				token.GET("/get", api.getEmailVerificationToken)
				token.POST("/verify", api.verifyEmailAddress)
			}
		}
	}

	// ipfs routes
	ipfs := v1.Group("/ipfs", authware...)
	{
		// public ipfs routes
		public := ipfs.Group("/public")
		{
			// pinning routes
			pin := public.Group("/pin")
			{
				pin.POST("/:hash", api.pinHashLocally)
				pin.GET("/check/:hash", api.checkLocalNodeForPin)
			}
			// file upload routes
			file := public.Group("/file")
			{
				file.POST("/add", api.addFileLocally)
				file.POST("/add/advanced", api.addFileLocallyAdvanced)
			}
			// pubsub routes
			pubsub := public.Group("/pubsub")
			{
				pubsub.POST("/publish/:topic", api.ipfsPubSubPublish)
			}
			// general routes
			public.GET("/stat/:key", api.getObjectStatForIpfs)
			public.GET("/dag/:hash", api.getDagObject)
			public.POST("/download/:hash", api.downloadContentHash)
		}

		// private ipfs routes
		private := ipfs.Group("/private")
		{
			// network management routes
			network := private.Group("/network")
			{
				network.GET("/:name", api.getIPFSPrivateNetworkByName)
				network.POST("/new", api.createHostedIPFSNetworkEntryInDatabase)
				network.POST("/stop", api.stopIPFSPrivateNetwork)
				network.POST("/start", api.startIPFSPrivateNetwork)
				network.DELETE("/remove", api.removeIPFSPrivateNetwork)
			}
			// pinning routes
			pin := private.Group("/pin")
			{
				pin.POST("/:hash", api.pinHashLocally)
				pin.GET("/check/:hash", api.checkLocalNodeForPin)
			}
			// file upload routes
			file := private.Group("/file")
			{
				file.POST("/add", api.addFileLocally)
				file.POST("/add/advanced", api.addFileLocallyAdvanced)
			}
			// pubsub routes
			pubsub := private.Group("/pubsub")
			{
				pubsub.POST("/publish/:topic", api.ipfsPubSubPublishToHostedIPFSNetwork)
			}
			// object stat route
			ipfs.POST("/stat/:key", api.getObjectStatForIpfsForHostedIPFSNetwork)
			// ipns routes
			ipns := private.Group("/ipns")
			{
				ipns.POST("/publish/details", api.publishDetailedIPNSToHostedIPFSNetwork)
			}
			// general routes
			private.GET("/networks", api.getAuthorizedPrivateNetworks)
			private.GET("/uploads/:network_name", api.getUploadsByNetworkName)
			private.POST("/download/:hash", api.downloadContentHashForPrivateNetwork)
		}
		// utility routes
		utils := ipfs.Group("/utils")
		{
			utils.POST("laser", api.BeamContent)
		}
		// ipfs cluster routes
		cluster := ipfs.Group("/cluster")
		{
			// sync control routes
			sync := cluster.Group("/sync")
			{
				errors := sync.Group("/errors")
				{
					errors.POST("/local", api.syncClusterErrorsLocally) // admin locked
				}
			}
			// status routes
			status := cluster.Group("/status")
			{
				// pin status route
				pin := status.Group("/pin")
				{
					pin.GET("/local/:hash", api.getLocalStatusForClusterPin)   // admin locked
					pin.GET("/global/:hash", api.getGlobalStatusForClusterPin) // admin locked
				}
				// local cluster status route
				status.GET("/local", api.fetchLocalClusterStatus)
			}
			// general routes
			cluster.POST("/pin/:hash", api.pinHashToCluster) // admin locked
		}
	}

	// ipns
	ipns := v1.Group("/ipns", authware...)
	{
		ipns.POST("/publish/details", api.publishToIPNSDetails)
		ipns.GET("/records", api.getIPNSRecordsPublishedByUser)
	}

	// database
	database := v1.Group("/database", authware...)
	{
		database.GET("/uploads", api.getUploadsFromDatabase)  // admin locked
		database.GET("/uploads/:user", api.getUploadsForUser) // partial admin locked
	}

	// frontend
	frontend := v1.Group("/frontend", authware...)
	{
		uploads := frontend.Group("/uploads")
		{
			uploads.GET("/encrypted", api.getEncryptedUploadsForUser)
		}
		cost := frontend.Group("/cost")
		{
			calculate := cost.Group("/calculate")
			{
				calculate.GET("/:hash/:holdtime", api.calculatePinCost)
				calculate.POST("/file", api.calculateFileCost)
			}
		}
	}

	// admin
	admin := v1.Group("/admin", authware...)
	{
		admin.POST("/utils/file-size-check", CalculateFileSize)
		mini := admin.Group("/mini")
		{
			mini.POST("/create/bucket", api.makeBucket)
		}
	}

	api.LogInfo("Routes initialized")
	return nil
}
