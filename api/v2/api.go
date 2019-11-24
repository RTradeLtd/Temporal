package v2

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/streadway/amqp"

	"github.com/RTradeLtd/Temporal/rtfscluster"
	"github.com/RTradeLtd/Temporal/utils"
	pbLens "github.com/RTradeLtd/grpc/lensv2"
	pbOrch "github.com/RTradeLtd/grpc/nexus"
	pbSigner "github.com/RTradeLtd/grpc/pay"
	pbBchWallet "github.com/gcash/bchwallet/rpc/walletrpc"

	"github.com/RTradeLtd/kaas/v2"
	"go.uber.org/zap"

	"github.com/RTradeLtd/ChainRider-Go/dash"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/rtfs/v2"

	limit "github.com/aviddiviner/gin-limit"

	"github.com/RTradeLtd/config/v2"
	stats "github.com/semihalev/gin-stats"

	"github.com/RTradeLtd/Temporal/api/middleware"
	"github.com/RTradeLtd/database/v2"
	"github.com/RTradeLtd/database/v2/models"

	"github.com/gin-gonic/gin"
)

// API is our API service
type API struct {
	ipfs        rtfs.Manager
	ipfsCluster *rtfscluster.ClusterManager
	keys        keys
	r           *gin.Engine
	cfg         *config.TemporalConfig
	dbm         *database.Manager
	um          *models.UserManager
	im          *models.IpnsManager
	pm          *models.PaymentManager
	ue          *models.EncryptedUploadManager
	upm         *models.UploadManager
	zm          *models.ZoneManager
	rm          *models.RecordManager
	nm          *models.HostedNetworkManager
	usage       *models.UsageManager
	orgs        *models.OrgManager
	l           *zap.SugaredLogger
	signer      pbSigner.SignerClient
	orch        pbOrch.ServiceClient
	lens        pbLens.LensV2Client
	bchWallet   pbBchWallet.WalletServiceClient
	dc          *dash.Client
	queues      queues
	clam        *utils.Shell
	service     string

	version string
}

// Initialize is used ot initialize our API service. debug = true is useful
// for debugging database issues.
func Initialize(
	ctx context.Context,
	// configuration
	cfg *config.TemporalConfig,
	version string,
	opts Options,
	clients Clients,
	// API dependencies
	l *zap.SugaredLogger,

) (*API, error) {
	var (
		err    error
		router = gin.Default()
	)
	// update dev mode
	dev = opts.DevMode
	l = l.Named("api")
	im, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		"", time.Minute*60,
	)
	if err != nil {
		return nil, err
	}
	imCluster, err := rtfscluster.Initialize(
		ctx,
		cfg.IPFSCluster.APIConnection.Host,
		cfg.IPFSCluster.APIConnection.Port,
	)
	if err != nil {
		return nil, err
	}

	// set up API struct
	api, err := new(cfg, router, l, clients, im, imCluster, opts.DebugLogging)
	if err != nil {
		return nil, err
	}
	api.version = version

	// init routes
	if err = api.setupRoutes(opts.DebugLogging); err != nil {
		return nil, err
	}
	api.l.Info("api initialization successful")

	// return our configured API service
	return api, nil
}

func new(cfg *config.TemporalConfig, router *gin.Engine, l *zap.SugaredLogger, clients Clients, ipfs rtfs.Manager, ipfsCluster *rtfscluster.ClusterManager, debug bool) (*API, error) {
	var (
		dbm *database.Manager
		err error
	)

	// set up database manager
	dbm, err = database.New(cfg, database.Options{LogMode: debug})
	if err != nil {
		l.Warnw("failed to connect to database with secure connection - attempting insecure", "error", err.Error())
		dbm, err = database.New(cfg, database.Options{
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
	kb1, err := kaas.NewClient(cfg.Services, false)
	if err != nil {
		return nil, err
	}
	kb2, err := kaas.NewClient(cfg.Services, true)
	if err != nil {
		return nil, err
	}
	// setup our queues
	qmIpns, err := queue.New(queue.IpnsEntryQueue, cfg.RabbitMQ.URL, true, dev, cfg, l.Named("ipns"))
	if err != nil {
		return nil, err
	}
	qmPin, err := queue.New(queue.IpfsPinQueue, cfg.RabbitMQ.URL, true, dev, cfg, l.Named("pin"))
	if err != nil {
		return nil, err
	}
	qmCluster, err := queue.New(queue.IpfsClusterPinQueue, cfg.RabbitMQ.URL, true, dev, cfg, l.Named("cluster"))
	if err != nil {
		return nil, err
	}
	qmEmail, err := queue.New(queue.EmailSendQueue, cfg.RabbitMQ.URL, true, dev, cfg, l.Named("email"))
	if err != nil {
		return nil, err
	}
	qmKey, err := queue.New(queue.IpfsKeyCreationQueue, cfg.RabbitMQ.URL, true, dev, cfg, l.Named("key"))
	if err != nil {
		return nil, err
	}
	qmDash, err := queue.New(queue.DashPaymentConfirmationQueue, cfg.RabbitMQ.URL, true, dev, cfg, l.Named("dash"))
	if err != nil {
		return nil, err
	}
	qmEth, err := queue.New(queue.EthPaymentConfirmationQueue, cfg.RabbitMQ.URL, true, dev, cfg, l.Named("eth"))
	if err != nil {
		return nil, err
	}
	qmBch, err := queue.New(queue.BitcoinCashPaymentConfirmationQueue, cfg.RabbitMQ.URL, true, dev, cfg, l.Named("bch"))
	if err != nil {
		return nil, err
	}
	qmENS, err := queue.New(queue.ENSRequestQueue, cfg.RabbitMQ.URL, true, dev, cfg, l.Named("ens"))
	if err != nil {
		return nil, err
	}
	clam, err := utils.NewShell("")
	if err != nil {
		return nil, err
	}
	if cfg.Stripe.SecretKey == "" {
		stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
		cfg.Stripe.SecretKey = stripeSecretKey
	}
	if cfg.Stripe.PublishableKey == "" {
		stripePublishableKey := os.Getenv("STRIPE_PUBLISHABLE_KEY")
		cfg.Stripe.PublishableKey = stripePublishableKey
	}
	// return
	return &API{
		ipfs:        ipfs,
		ipfsCluster: ipfsCluster,
		keys:        keys{kb1: kb1, kb2: kb2},
		cfg:         cfg,
		service:     "api",
		r:           router,
		l:           l,
		dbm:         dbm,
		um:          models.NewUserManager(dbm.DB),
		im:          models.NewIPNSManager(dbm.DB),
		pm:          models.NewPaymentManager(dbm.DB),
		ue:          models.NewEncryptedUploadManager(dbm.DB),
		upm:         models.NewUploadManager(dbm.DB),
		usage:       models.NewUsageManager(dbm.DB),
		orgs:        models.NewOrgManager(dbm.DB),
		lens:        clients.Lens,
		signer:      clients.Signer,
		orch:        clients.Orch,
		bchWallet:   clients.BchWallet,
		dc:          dc,
		queues: queues{
			pin:     qmPin,
			cluster: qmCluster,
			email:   qmEmail,
			ipns:    qmIpns,
			key:     qmKey,
			dash:    qmDash,
			eth:     qmEth,
			bch:     qmBch,
			ens:     qmENS,
		},
		zm:   models.NewZoneManager(dbm.DB),
		rm:   models.NewRecordManager(dbm.DB),
		nm:   models.NewHostedNetworkManager(dbm.DB),
		clam: clam,
	}, nil
}

// Close releases API resources
func (api *API) Close() {
	// close queue resources
	if err := api.queues.cluster.Close(); err != nil {
		api.l.Error(err, "failed to properly close cluster queue connection")
	}
	if err := api.queues.email.Close(); err != nil {
		api.l.Error(err, "failed to properly close email queue connection")
	}
	if err := api.queues.ipns.Close(); err != nil {
		api.l.Error(err, "failed to properly close ipns queue connection")
	}
	if err := api.queues.key.Close(); err != nil {
		api.l.Error(err, "failed to properly close key queue connection")
	}
	if err := api.queues.pin.Close(); err != nil {
		api.l.Error(err, "failed to properly close pin queue connection")
	}
}

// TLSConfig is used to enable TLS on the API service
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// ListenAndServe spins up the API server
func (api *API) ListenAndServe(ctx context.Context, addr string, tlsConfig *TLSConfig) error {
	server := &http.Server{
		Addr:    addr,
		Handler: api.r,
	}
	errChan := make(chan error, 1)
	go func() {
		if tlsConfig != nil {
			// configure TLS to override defaults
			tlsCfg := &tls.Config{
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
			server.TLSConfig = tlsCfg
			errChan <- server.ListenAndServeTLS(tlsConfig.CertFile, tlsConfig.KeyFile)
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
			return server.Close()
		case msg := <-api.queues.cluster.ErrCh:
			qmCluster, err := api.handleQueueError(msg, api.cfg.RabbitMQ.URL, queue.IpfsClusterPinQueue, true)
			if err != nil {
				return server.Close()
			}
			api.queues.cluster = qmCluster
		case msg := <-api.queues.dash.ErrCh:
			qmDash, err := api.handleQueueError(msg, api.cfg.RabbitMQ.URL, queue.DashPaymentConfirmationQueue, true)
			if err != nil {
				return server.Close()
			}
			api.queues.dash = qmDash
		case msg := <-api.queues.email.ErrCh:
			qmEmail, err := api.handleQueueError(msg, api.cfg.RabbitMQ.URL, queue.EmailSendQueue, true)
			if err != nil {
				return server.Close()
			}
			api.queues.email = qmEmail
		case msg := <-api.queues.ipns.ErrCh:
			qmIpns, err := api.handleQueueError(msg, api.cfg.RabbitMQ.URL, queue.IpnsEntryQueue, true)
			if err != nil {
				return server.Close()
			}
			api.queues.ipns = qmIpns
		case msg := <-api.queues.key.ErrCh:
			qmKey, err := api.handleQueueError(msg, api.cfg.RabbitMQ.URL, queue.IpfsKeyCreationQueue, true)
			if err != nil {
				return server.Close()
			}
			api.queues.key = qmKey
		case msg := <-api.queues.eth.ErrCh:
			qmEth, err := api.handleQueueError(msg, api.cfg.RabbitMQ.URL, queue.EthPaymentConfirmationQueue, true)
			if err != nil {
				return server.Close()
			}
			api.queues.eth = qmEth
		case msg := <-api.queues.pin.ErrCh:
			qmPin, err := api.handleQueueError(msg, api.cfg.RabbitMQ.URL, queue.IpfsPinQueue, true)
			if err != nil {
				return server.Close()
			}
			api.queues.pin = qmPin
		case msg := <-api.queues.bch.ErrCh:
			qmBch, err := api.handleQueueError(msg, api.cfg.RabbitMQ.URL, queue.BitcoinCashPaymentConfirmationQueue, true)
			if err != nil {
				return server.Close()
			}
			api.queues.bch = qmBch
		}
	}
}

// setupRoutes is used to setup all of our api routes
func (api *API) setupRoutes(debug bool) error {
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
	// ensure we have valid cors configuration, otherwise default to allow all
	var allowedOrigins []string
	if len(api.cfg.API.Connection.CORS.AllowedOrigins) > 0 {
		allowedOrigins = api.cfg.API.Connection.CORS.AllowedOrigins
	}
	// set up defaults
	api.r.Use(
		// cors middleware
		middleware.CORSMiddleware(dev, debug, allowedOrigins),
		// allows for automatic xss removal
		// greater than what can be configured with HTTP Headers
		xssMdlwr.RemoveXss(),
		// rate limiting
		limit.MaxAllowed(connLimit),
		// security middleware
		middleware.NewSecWare(dev),
		// request id middleware
		middleware.RequestID(),
		// stats middleware
		stats.RequestStats())

	// set up middleware
	ginjwt := middleware.JwtConfigGenerate(api.cfg.JWT.Key, api.cfg.JWT.Realm, api.dbm.DB, api.l)
	authware := []gin.HandlerFunc{ginjwt.MiddlewareFunc()}

	// V2 API
	v2 := api.r.Group("/v2")

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
		auth.GET("/refresh", ginjwt.RefreshHandler)
	}

	// statistics
	statistics := v2.Group("/statistics").Use(authware...)
	{
		statistics.GET("/stats", api.getStats)
	}

	// lens search engine
	lens := v2.Group("/lens")
	{
		lens.POST("/index", api.submitIndexRequest)
		// only allow registered users to search
		lens.POST("/search", api.submitSearchRequest)
	}

	// payments
	payments := v2.Group("/payments", authware...)
	{
		dash := payments.Group("/dash")
		{
			dash.POST("/create", api.CreateDashPayment)
		}
		eth := payments.Group("/eth")
		{
			eth.POST("/request", api.RequestSignedPaymentMessage)
			eth.POST("/confirm", api.ConfirmETHPayment)
		}
		bch := payments.Group("/bch")
		{
			bch.POST("/create", api.createBchPayment)
			bch.POST("/confirm", api.confirmBchPayment)
		}
		stripe := payments.Group("/stripe")
		{
			stripe.POST("/charge", api.stripeCharge)
		}
		payments.GET("/status/:number", api.getPaymentStatus)
	}

	// accounts
	account := v2.Group("/account")
	{
		token := account.Group("/token", authware...)
		{
			token.GET("/username", api.getUserFromToken)
		}
		password := account.Group("/password", authware...)
		{
			password.POST("/change", api.changeAccountPassword)
		}
		key := account.Group("/key", authware...)
		{
			key.GET("/export/:name", api.exportKey)
			ipfs := key.Group("/ipfs")
			{
				ipfs.GET("/get", api.getIPFSKeyNamesForAuthUser)
				ipfs.POST("/new", api.createIPFSKey)
			}
		}
		credits := account.Group("/credits", authware...)
		{
			credits.GET("/available", api.getCredits)
		}
		email := account.Group("/email")
		{
			// auth-less account email routes
			token := email.Group("/verify")
			{
				token.GET("/:user/:token", api.verifyEmailAddress)
			}
			// authenticatoin email routes
			auth := email.Use(authware...)
			{
				auth.POST("/forgot", api.forgotEmail)
			}
		}
		auth := account.Use(authware...)
		{
			// used to upgrade account to light tier
			auth.POST("/upgrade", api.upgradeAccount)
			auth.GET("/usage", api.usageData)
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
				pin.POST("/:hash/extend", api.extendPin)
			}
			// file upload routes
			file := public.Group("/file")
			{
				file.POST("/add", api.addFile)
				file.POST("/add/directory", api.uploadDirectory)
			}
			// pubsub routes
			pubsub := public.Group("/pubsub")
			{
				pubsub.POST("/publish/:topic", api.ipfsPubSubPublish)
			}
			// general routes
			public.GET("/stat/:hash", api.getObjectStatForIpfs)
			public.GET("/dag/:hash", api.getDagObject)
		}

		// private ipfs routes
		private := ipfs.Group("/private")
		{
			// network management routes
			private.GET("/networks", api.getAuthorizedPrivateNetworks)
			network := private.Group("/network")
			{
				users := network.Group("/users")
				{
					users.DELETE("/remove", api.removeUsersFromNetwork)
					users.POST("/add", api.addUsersToNetwork)
				}
				owners := network.Group("/owners")
				{
					owners.POST("/add", api.addOwnersToNetwork)
				}
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
	}

	// ipns
	ipns := v2.Group("/ipns", authware...)
	{
		// public ipns routes
		public := ipns.Group("/public")
		{
			public.POST("/publish/details", api.publishToIPNSDetails)
			// used to handle pinning of IPNS records on public ipfs
			// this involves first resolving the record, parsing it
			// and extracting the hash to pin
			public.POST("/pin", api.pinIPNSHash)
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
				calculate.GET("/:hash/:hold_time", api.calculatePinCost)
				calculate.POST("/file", api.calculateFileCost)
			}
		}
	}

	// organization routes
	org := v2.Group("/org", authware...)
	{
		get := org.Group("/get")
		{
			get.GET("/model", api.getOrganization)
			get.GET("/billing/report", api.getOrgBillingReport)
		}
		org.POST("/new", api.newOrganization)
		org.POST("/register/user", api.registerOrgUser)
	}
	api.l.Info("Routes initialized")
	return nil
}

// HandleQueueError is used to handle queue connectivity issues requiring us to re-connect
func (api *API) handleQueueError(amqpErr *amqp.Error, rabbitMQURL string, queueType queue.Queue, publish bool) (*queue.Manager, error) {
	api.l.Errorw(
		"a protocol connection error stopping rabbitmq was received",
		"queue", queueType.String(),
		"error", amqpErr.Error())
	qManager, err := queue.New(queueType, rabbitMQURL, publish, dev, api.cfg, api.l)
	if err != nil {
		api.l.Errorw(
			"failed to re-establish queue process, exiting",
			"queue", queueType.String(),
			"error", err.Error())
		return nil, err
	}
	api.l.Warnw(
		"successfully re-established queue connection", "queue", queueType.String())
	return qManager, nil
}
