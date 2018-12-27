// Package v2 is the main package for Temporal's
// http api
package v2

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/log"
	"github.com/RTradeLtd/Temporal/rtfscluster"
	"go.uber.org/zap"

	"github.com/RTradeLtd/kaas"

	"github.com/RTradeLtd/ChainRider-Go/dash"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/rtfs"

	limit "github.com/aviddiviner/gin-limit"

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
)

var (
	xssMdlwr xss.XssMw
	dev      = false
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
	upm         *models.UploadManager
	zm          *models.ZoneManager
	rm          *models.RecordManager
	nm          *models.IPFSNetworkManager
	l           *zap.SugaredLogger
	signer      pbSigner.SignerClient
	orch        pbOrch.ServiceClient
	lens        pbLens.IndexerAPIClient
	dc          *dash.Client
	queues      queues
	service     string
}

// Initialize is used ot initialize our API service. debug = true is useful
// for debugging database issues.
func Initialize(cfg *config.TemporalConfig, l *zap.SugaredLogger, debug bool, lens pbLens.IndexerAPIClient, orch pbOrch.ServiceClient, signer pbSigner.SignerClient) (*API, error) {
	l = l.Named("api")
	var (
		err    error
		router = gin.Default()
		p      = ginprometheus.NewPrometheus("gin")
	)

	// set up prometheus monitoring
	p.SetListenAddress(fmt.Sprintf("%s:6768", cfg.API.Connection.ListenAddress))
	p.Use(router)

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
	api, err := new(cfg, router, l, lens, orch, signer, im, imCluster, debug)
	if err != nil {
		return nil, err
	}

	// init routes
	if err = api.setupRoutes(); err != nil {
		return nil, err
	}
	api.l.Info("api initialization successful")

	// return our configured API service
	return api, nil
}

func new(cfg *config.TemporalConfig, router *gin.Engine, l *zap.SugaredLogger, lens pbLens.IndexerAPIClient, orch pbOrch.ServiceClient, signer pbSigner.SignerClient, ipfs rtfs.Manager, ipfsCluster *rtfscluster.ClusterManager, debug bool) (*API, error) {
	var (
		dbm *database.Manager
		err error
	)

	// set up database manager
	dbm, err = database.Initialize(cfg, database.Options{LogMode: debug})
	if err != nil {
		l.Warnw("failed to connect to database with secure connection - attempting insecure", "error", err.Error())
		dbm, err = database.Initialize(cfg, database.Options{
			LogMode:        debug,
			SSLModeDisable: true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database with insecure connection: %s", err.Error())
		}
		l.Warn("insecure database connection established")
	} else {
		l.Info("secure database connection established")
	}
	var networkVersion string
	if dev {
		networkVersion = "testnet"
	} else {
		networkVersion = "main"
	}
	dc := dash.NewClient(&dash.ConfigOpts{
		APIVersion:      "v1",
		DigitalCurrency: "dash",
		//TODO: change to main before production release
		Blockchain: networkVersion,
		Token:      cfg.APIKeys.ChainRider,
	})
	keys, err := kaas.NewClient(cfg.Endpoints)
	if err != nil {
		return nil, err
	}
	logger, err := log.NewLogger("", dev)
	if err != nil {
		return nil, err
	}
	// setup our queues
	qmIpns, err := queue.New(queue.IpnsEntryQueue, cfg.RabbitMQ.URL, true, logger)
	if err != nil {
		return nil, err
	}
	qmPin, err := queue.New(queue.IpfsPinQueue, cfg.RabbitMQ.URL, true, logger)
	if err != nil {
		return nil, err
	}
	qmDatabase, err := queue.New(queue.DatabaseFileAddQueue, cfg.RabbitMQ.URL, true, logger)
	if err != nil {
		return nil, err
	}
	qmFile, err := queue.New(queue.IpfsFileQueue, cfg.RabbitMQ.URL, true, logger)
	if err != nil {
		return nil, err
	}
	qmCluster, err := queue.New(queue.IpfsClusterPinQueue, cfg.RabbitMQ.URL, true, logger)
	if err != nil {
		return nil, err
	}
	qmEmail, err := queue.New(queue.EmailSendQueue, cfg.RabbitMQ.URL, true, logger)
	if err != nil {
		return nil, err
	}
	qmKey, err := queue.New(queue.IpfsKeyCreationQueue, cfg.RabbitMQ.URL, true, logger)
	if err != nil {
		return nil, err
	}
	qmDash, err := queue.New(queue.DashPaymentConfirmationQueue, cfg.RabbitMQ.URL, true, logger)
	if err != nil {
		return nil, err
	}
	qmPayConfirm, err := queue.New(queue.PaymentConfirmationQueue, cfg.RabbitMQ.URL, true, logger)
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
		l:           l,
		dbm:         dbm,
		um:          models.NewUserManager(dbm.DB),
		im:          models.NewIPNSManager(dbm.DB),
		pm:          models.NewPaymentManager(dbm.DB),
		dm:          models.NewDropManager(dbm.DB),
		ue:          models.NewEncryptedUploadManager(dbm.DB),
		upm:         models.NewUploadManager(dbm.DB),
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

// TLSFiles is used to enable TLS on the API service
type TLSFiles struct {
	CertFile string
	KeyFile  string
}

// ListenAndServe spins up the API server
func (api *API) ListenAndServe(ctx context.Context, addr string, tlsFiles *TLSFiles) error {
	server := &http.Server{
		Addr:    addr,
		Handler: api.r,
	}
	errChan := make(chan error, 1)
	go func() {
		if tlsFiles != nil {
			// configure TLS to override defaults
			tlsConfig := &tls.Config{
				CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
				//PreferServerCipherSuites: true,
				CipherSuites: []uint16{
					// http/2 mandated supported cipher
					// unforunately this is a less secure cipher
					// but specifying it first is the only way to accept
					// http/2 connections without go throwing an error
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					// super duper secure ciphers
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				},
				// consider whether or not to fix to tls1.2
				MinVersion: tls.VersionTLS11,
			}
			// set tls configuration
			server.TLSConfig = tlsConfig
			errChan <- server.ListenAndServeTLS(tlsFiles.CertFile, tlsFiles.KeyFile)
			return
		}
		errChan <- server.ListenAndServe()
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
		// allows for automatic xss removal
		// greater than what can be configured with HTTP Headers
		xssMdlwr.RemoveXss(),
		// rate limiting
		limit.MaxAllowed(connLimit),
		// security middleware
		middleware.NewSecWare(dev),
		// cors middleware
		middleware.CORSMiddleware(),
		// stats middleware
		stats.RequestStats())

	// set up middleware
	ginjwt := middleware.JwtConfigGenerate(api.cfg.JWT.Key, api.cfg.JWT.Realm, api.dbm.DB, api.l)
	authware := []gin.HandlerFunc{ginjwt.MiddlewareFunc()}

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
		}

		// private ipfs routes
		private := ipfs.Group("/private")
		{
			// network management routes
			private.GET("/networks", api.getAuthorizedPrivateNetworks)
			network := private.Group("/network")
			{
				network.GET("/:name", api.getIPFSPrivateNetworkByName)
				network.POST("/new", api.createIPFSNetwork)
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
		}
		// utility routes
		utils := ipfs.Group("/utils")
		{
			// generic download
			utils.POST("/download/:hash", api.downloadContentHash)
			laser := utils.Group("/laser")
			{
				laser.POST("/beam", api.beamContent)
			}
		}
		// ipfs cluster routes
		cluster := ipfs.Group("/cluster")
		{
			// general routes
			cluster.POST("/pin/:hash", api.pinHashToCluster)
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
		database.GET("/uploads", api.getUploadsForUser)
		database.GET("/uploads/encrypted", api.getEncryptedUploadsForUser)
	}

	// frontend
	frontend := v2.Group("/frontend", authware...)
	{
		cost := frontend.Group("/cost")
		{
			calculate := cost.Group("/calculate")
			{
				calculate.GET("/:hash/:holdtime", api.calculatePinCost)
				calculate.POST("/file", api.calculateFileCost)
			}
		}
	}

	api.l.Info("Routes initialized")
	return nil
}
