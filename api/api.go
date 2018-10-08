// Package api is the main package for Temporal's
// http api
package api

import (
	"fmt"
	"io"
	"net/http"
	"os"

	limit "github.com/aviddiviner/gin-limit"
	helmet "github.com/danielkov/gin-helmet"
	"github.com/sirupsen/logrus"

	"github.com/RTradeLtd/config"
	xss "github.com/dvwright/xss-mw"
	stats "github.com/semihalev/gin-stats"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	"github.com/RTradeLtd/Temporal/api/middleware"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	xssMdlwr xss.XssMw
)

// API is our API service
type API struct {
	r       *gin.Engine
	cfg     *config.TemporalConfig
	dbm     *database.DatabaseManager
	um      *models.UserManager
	im      *models.IpnsManager
	l       *log.Logger
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
	logfile, err := os.OpenFile("/var/log/temporal/api_service.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %s", err)
	}

	// set up API struct
	api, err := new(cfg, router, debug, logfile)
	if err != nil {
		return nil, err
	}

	// init routes
	api.setupRoutes()
	api.LogInfo("api initialization successful")

	// return our configured API service
	return api, nil
}

func new(cfg *config.TemporalConfig, router *gin.Engine, debug bool, out io.Writer) (*API, error) {
	var (
		logger = log.New()
		dbm    *database.DatabaseManager
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
	dbm, err = database.Initialize(cfg, database.DatabaseOptions{LogMode: debug})
	if err != nil {
		logger.Warnf("failed to connect to database: %s", err.Error())
		logger.Warnf("failed to connect to database with secure connection - attempting insecure connection...")
		dbm, err = database.Initialize(cfg, database.DatabaseOptions{
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

	return &API{
		cfg:     cfg,
		service: "api",
		r:       router,
		l:       logger,
		dbm:     dbm,
		um:      models.NewUserManager(dbm.DB),
		im:      models.NewIPNSManager(dbm.DB),
	}, nil
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
func (api *API) setupRoutes() {
	// set up defaults
	api.r.Use(
		xssMdlwr.RemoveXss(),
		limit.MaxAllowed(20),
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

	// authentication
	auth := v1.Group("/auth")
	auth.POST("/register", api.registerUserAccount)
	auth.POST("/login", ginjwt.LoginHandler)

	// statistics
	statistics := v1.Group("/statistics").Use(authware...)
	statistics.GET("/stats", func(c *gin.Context) {
		username := GetAuthenticatedUserFromContext(c)
		if err := api.validateAdminRequest(username); err != nil {
			FailNotAuthorized(c, UnAuthorizedAdminAccess)
			return
		}
		c.JSON(http.StatusOK, stats.Report())
	})

	// payments
	payments := v1.Group("/payments", authware...)
	payments.POST("/create", api.CreatePayment)
	payments.GET("/deposit/address/:type", api.GetDepositAddress)

	// accounts
	account := v1.Group("/account", authware...)
	account.POST("password/change", api.changeAccountPassword)
	account.GET("/key/ipfs/get", api.getIPFSKeyNamesForAuthUser)
	account.POST("/key/ipfs/new", api.createIPFSKey)
	account.GET("/credits/available", api.getCredits)

	// ipfs
	ipfs := v1.Group("/ipfs", authware...)
	ipfs.POST("/pubsub/publish/:topic", api.ipfsPubSubPublish)
	ipfs.POST("/calculate-content-hash", api.calculateContentHashForFile)
	ipfs.GET("/pins", api.getLocalPins) // admin locked
	ipfs.GET("/object-stat/:key", api.getObjectStatForIpfs)
	ipfs.GET("/check-for-pin/:hash", api.checkLocalNodeForPin) // admin locked
	ipfs.POST("/download/:hash", api.downloadContentHash)
	ipfs.POST("/pin/:hash", api.pinHashLocally)
	ipfs.POST("/add-file", api.addFileLocally)
	ipfs.POST("/add-file/advanced", api.addFileLocallyAdvanced)

	// ipfs-private
	ipfsPrivate := v1.Group("/ipfs-private", authware...)
	ipfsPrivate.POST("/new/network", api.createHostedIPFSNetworkEntryInDatabase)                // admin locked
	ipfsPrivate.GET("/network/:name", api.getIPFSPrivateNetworkByName)                          // admin locked
	ipfsPrivate.POST("/ipfs/check-for-pin/:hash", api.checkLocalNodeForPinForHostedIPFSNetwork) // admin locked
	ipfsPrivate.POST("/ipfs/object-stat/:key", api.getObjectStatForIpfsForHostedIPFSNetwork)
	ipfsPrivate.POST("/pubsub/publish/:topic", api.ipfsPubSubPublishToHostedIPFSNetwork)
	ipfsPrivate.POST("/pins", api.getLocalPinsForHostedIPFSNetwork) // admin locked
	ipfsPrivate.GET("/networks", api.getAuthorizedPrivateNetworks)
	ipfsPrivate.GET("/uploads/:network_name", api.getUploadsByNetworkName)
	ipfsPrivate.POST("/ipfs/pin/:hash", api.pinToHostedIPFSNetwork)
	ipfsPrivate.POST("/ipfs/add-file", api.addFileToHostedIPFSNetwork)
	ipfsPrivate.POST("/ipfs/add-file/advanced", api.addFileToHostedIPFSNetworkAdvanced)
	ipfsPrivate.POST("/ipns/publish/details", api.publishDetailedIPNSToHostedIPFSNetwork)

	// ipns
	ipns := v1.Group("/ipns", authware...)
	ipns.POST("/publish/details", api.publishToIPNSDetails)
	ipns.POST("/dnslink/aws/add", api.generateDNSLinkEntry) // admin locked
	ipns.GET("/records", api.getIPNSRecordsPublishedByUser)

	// ipfs-cluster
	cluster := v1.Group("/ipfs-cluster", authware...)
	cluster.POST("/sync-errors-local", api.syncClusterErrorsLocally)          // admin locked
	cluster.GET("/status-local-pin/:hash", api.getLocalStatusForClusterPin)   // admin locked
	cluster.GET("/status-global-pin/:hash", api.getGlobalStatusForClusterPin) // admin locked
	cluster.GET("/status-local", api.fetchLocalClusterStatus)                 // admin locked
	cluster.POST("/pin/:hash", api.pinHashToCluster)

	// database
	database := v1.Group("/database", authware...)
	database.GET("/uploads", api.getUploadsFromDatabase)  // admin locked
	database.GET("/uploads/:user", api.getUploadsForUser) // partial admin locked

	// frontend
	frontend := v1.Group("/frontend", authware...)
	frontend.GET("/cost/calculate/:hash/:holdtime", api.calculatePinCost)
	frontend.POST("/cost/calculate/file", api.calculateFileCost)

	// admin
	admin := v1.Group("/admin", authware...)
	admin.POST("/utils/file-size-check", CalculateFileSize)
	mini := admin.Group("/mini")
	mini.POST("/create/bucket", api.makeBucket)

	api.LogInfo("Routes initialized")
}
