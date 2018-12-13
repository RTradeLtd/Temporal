// Package v2 is the main package for Temporal's
// http api
package v2

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/rtfscluster"

	"github.com/RTradeLtd/kaas"

	"github.com/RTradeLtd/ChainRider-Go/dash"
	"github.com/RTradeLtd/Temporal/queue"
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

	pbOrch "github.com/RTradeLtd/grpc/ipfs-orchestrator"
	pbLens "github.com/RTradeLtd/grpc/lens"
	pbSigner "github.com/RTradeLtd/grpc/temporal"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	xssMdlwr xss.XssMw
	dev      = true
)

const (
	realName = "temporal-realm"
)

// API is our API service
type API struct {
	ipfs        rtfs.Manager
	ipfsCluster *rtfscluster.ClusterManager
	keys        *kaas.Client
	r           *gin.Engine
	cfg         *config.TemporalConfig
	dbm         *database.Manager
	um          *models.UserManager
	im          *models.IpnsManager
	pm          *models.PaymentManager
	dm          *models.DropManager
	ue          *models.EncryptedUploadManager
	zm          *models.ZoneManager
	rm          *models.RecordManager
	nm          *models.IPFSNetworkManager
	l           *log.Logger
	signer      pbSigner.SignerClient
	orch        pbOrch.ServiceClient
	lens        pbLens.IndexerAPIClient
	dc          *dash.Client
	queues      queues
	service     string
}

// Initialize is used ot initialize our API service. debug = true is useful
// for debugging database issues.
func Initialize(cfg *config.TemporalConfig, debug bool, lens pbLens.IndexerAPIClient, orch pbOrch.ServiceClient, signer pbSigner.SignerClient) (*API, error) {
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
	im, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		nil,
		time.Minute*10,
	)
	if err != nil {
		return nil, err
	}
	imCluster, err := rtfscluster.Initialize(
		cfg.IPFSCluster.APIConnection.Host,
		cfg.IPFSCluster.APIConnection.Port,
	)
	if err != nil {
		return nil, err
	}
	// set up API struct
	api, err := new(cfg, router, lens, orch, signer, im, imCluster, debug, logfile)
	if err != nil {
		return nil, err
	}

	// init routes
	if err = api.setupRoutes(); err != nil {
		return nil, err
	}
	api.LogInfo("api initialization successful")

	// return our configured API service
	return api, nil
}

func new(cfg *config.TemporalConfig, router *gin.Engine, lens pbLens.IndexerAPIClient, orch pbOrch.ServiceClient, signer pbSigner.SignerClient, ipfs rtfs.Manager, ipfsCluster *rtfscluster.ClusterManager, debug bool, out io.Writer) (*API, error) {
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
	keys, err := kaas.NewClient(cfg.Endpoints)
	if err != nil {
		return nil, err
	}
	// setup our queues
	qmIpns, err := queue.New(queue.IpnsEntryQueue, cfg.RabbitMQ.URL, true)
	if err != nil {
		return nil, err
	}
	qmPin, err := queue.New(queue.IpfsPinQueue, cfg.RabbitMQ.URL, true)
	if err != nil {
		return nil, err
	}
	qmDatabase, err := queue.New(queue.DatabaseFileAddQueue, cfg.RabbitMQ.URL, true)
	if err != nil {
		return nil, err
	}
	qmFile, err := queue.New(queue.IpfsFileQueue, cfg.RabbitMQ.URL, true)
	if err != nil {
		return nil, err
	}
	qmCluster, err := queue.New(queue.IpfsClusterPinQueue, cfg.RabbitMQ.URL, true)
	if err != nil {
		return nil, err
	}
	qmEmail, err := queue.New(queue.EmailSendQueue, cfg.RabbitMQ.URL, true)
	if err != nil {
		return nil, err
	}
	qmKey, err := queue.New(queue.IpfsKeyCreationQueue, cfg.RabbitMQ.URL, true)
	if err != nil {
		return nil, err
	}
	qmDash, err := queue.New(queue.DashPaymentConfirmationQueue, cfg.RabbitMQ.URL, true)
	if err != nil {
		return nil, err
	}
	qmPayConfirm, err := queue.New(queue.PaymentConfirmationQueue, cfg.RabbitMQ.URL, true)
	if err != nil {
		return nil, err
	}
	return &API{
		ipfs:        ipfs,
		ipfsCluster: ipfsCluster,
		keys:        keys,
		cfg:         cfg,
		service:     "api",
		r:           router,
		l:           logger,
		dbm:         dbm,
		um:          models.NewUserManager(dbm.DB),
		im:          models.NewIPNSManager(dbm.DB),
		pm:          models.NewPaymentManager(dbm.DB),
		dm:          models.NewDropManager(dbm.DB),
		ue:          models.NewEncryptedUploadManager(dbm.DB),
		lens:        lens,
		signer:      signer,
		orch:        orch,
		dc:          dc,
		queues: queues{
			pin:        qmPin,
			file:       qmFile,
			cluster:    qmCluster,
			email:      qmEmail,
			ipns:       qmIpns,
			key:        qmKey,
			database:   qmDatabase,
			dash:       qmDash,
			payConfirm: qmPayConfirm,
		},
		zm: models.NewZoneManager(dbm.DB),
		rm: models.NewRecordManager(dbm.DB),
		nm: models.NewHostedIPFSNetworkManager(dbm.DB),
	}, nil
}

// Close releases API resources
func (api *API) Close() {
	// close queue resources
	if err := api.queues.cluster.Close(); err != nil {
		api.LogError(err, "failed to properly close cluster queue connection")
	}
	if err := api.queues.database.Close(); err != nil {
		api.LogError(err, "failed to properly close database queue connection")
	}
	if err := api.queues.email.Close(); err != nil {
		api.LogError(err, "failed to properly close email queue connection")
	}
	if err := api.queues.file.Close(); err != nil {
		api.LogError(err, "failed to properly close file queue connection")
	}
	if err := api.queues.ipns.Close(); err != nil {
		api.LogError(err, "failed to properly close ipns queue connection")
	}
	if err := api.queues.key.Close(); err != nil {
		api.LogError(err, "failed to properly close key queue connection")
	}
	if err := api.queues.pin.Close(); err != nil {
		api.LogError(err, "failed to properly close pin queue connection")
	}
}

// TLSConfig is used to enable TLS on the API service
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// ListenAndServe spins up the API server
func (api *API) ListenAndServe(ctx context.Context, addr string, tls *TLSConfig) error {
	errChan := make(chan error, 1)
	go func() {
		if tls != nil {
			errChan <- api.r.RunTLS(addr, tls.CertFile, tls.KeyFile)
			return
		}
		errChan <- api.r.Run(addr)
		return
	}()
	for {
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return nil
		}
	}
}

// setupRoutes is used to setup all of our api routes
func (api *API) setupRoutes() error {
	var (
		connLimit int
		err       error
	)
	if api.cfg.API.Connection.Limit == "" {
		connLimit = 50
	} else {
		// setup the connection limit
		connLimit, err = strconv.Atoi(api.cfg.API.Connection.Limit)
		if err != nil {
			return err
		}
	}
	// set up defaults
	api.r.Use(
		// slightly more complex xss removal
		xssMdlwr.RemoveXss(),
		// rate limiting
		limit.MaxAllowed(connLimit),
		// security restrictions
		helmet.NoSniff(),
		helmet.IENoOpen(),
		helmet.NoCache(),
		// basic xss removal
		helmet.XSSFilter(),
		// cors middleware
		middleware.CORSMiddleware(),
		// stats middleware
		stats.RequestStats())

	// set up middleware
	ginjwt := middleware.JwtConfigGenerate(api.cfg.API.JwtKey, realName, api.dbm.DB, api.l)
	authware := []gin.HandlerFunc{
		ginjwt.MiddlewareFunc(),
		middleware.APIRestrictionMiddleware(api.dbm.DB),
	}

	// V2 API
	v2 := api.r.Group("/api/v2")

	// system checks used to verify the integrity of our services
	systemChecks := v2.Group("/systems")
	{
		systemChecks.GET("/check", api.SystemsCheck)
	}

	// authless account recovery routes
	forgot := v2.Group("/forgot")
	{
		forgot.POST("/username", api.forgotUserName)
		forgot.POST("/password", api.resetPassword)
	}

	// authentication
	auth := v2.Group("/auth")
	{
		auth.POST("/register", api.registerUserAccount)
		auth.POST("/login", ginjwt.LoginHandler)
	}

	// statistics
	statistics := v2.Group("/statistics").Use(authware...)
	{
		statistics.GET("/stats", api.getStats)
	}

	// lens search engine
	lens := v2.Group("/lens")
	{
		// allow anyone to index
		lens.POST("/index", api.submitIndexRequest)
		// only allow registered users to search
		lens.POST("/search", api.submitSearchRequest)
	}

	// payments
	payments := v2.Group("/payments", authware...)
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
	account := v2.Group("/account", authware...)
	{
		token := account.Group("/token")
		{
			token.GET("/username", api.getUserFromToken)
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
	ipfs := v2.Group("/ipfs", authware...)
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
			private.GET("/networks", api.getAuthorizedPrivateNetworks)
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
				pin.POST("/:hash", api.pinToHostedIPFSNetwork)
				pin.GET("/check/:hash/:networkName", api.checkLocalNodeForPinForHostedIPFSNetwork)
			}
			// file upload routes
			file := private.Group("/file")
			{
				file.POST("/add", api.addFileToHostedIPFSNetwork)
				file.POST("/add/advanced", api.addFileToHostedIPFSNetworkAdvanced)
			}
			// pubsub routes
			pubsub := private.Group("/pubsub")
			{
				pubsub.POST("/publish/:topic", api.ipfsPubSubPublishToHostedIPFSNetwork)
			}
			// object stat route
			private.GET("/stat/:hash/:networkName", api.getObjectStatForIpfsForHostedIPFSNetwork)
			// general routes
			private.GET("/dag/:hash/:networkName", api.getDagObjectForHostedIPFSNetwork)
			private.GET("/uploads/:networkName", api.getUploadsByNetworkName)
			private.POST("/download/:hash", api.downloadContentHashForPrivateNetwork)
		}
		// utility routes
		utils := ipfs.Group("/utils")
		{
			laser := utils.Group("/laser")
			{
				laser.POST("/beam", api.beamContent)
			}
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
	ipns := v2.Group("/ipns", authware...)
	{
		// public ipns routes
		public := ipns.Group("/public")
		{
			public.POST("/publish/details", api.publishToIPNSDetails)
		}
		// private ipns routes
		private := ipns.Group("/private")
		{
			private.POST("/publish/details", api.publishDetailedIPNSToHostedIPFSNetwork)
		}
		// general routes
		ipns.GET("/records", api.getIPNSRecordsPublishedByUser)
	}

	// database
	database := v2.Group("/database", authware...)
	{
		database.GET("/uploads", api.getUploadsFromDatabase)  // admin locked
		database.GET("/uploads/:user", api.getUploadsForUser) // partial admin locked
	}

	// frontend
	frontend := v2.Group("/frontend", authware...)
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
	admin := v2.Group("/admin", authware...)
	{
		mini := admin.Group("/mini")
		{
			mini.POST("/create/bucket", api.makeBucket)
		}
	}

	api.LogInfo("Routes initialized")
	return nil
}
