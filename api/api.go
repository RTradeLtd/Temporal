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

	v1 := api.r.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.POST("/register", api.registerUserAccount)
	auth.POST("/login", ginjwt.LoginHandler)

	statsProtected := v1.Group("/statistics").Use(authware...)
	statsProtected.GET("/stats", func(c *gin.Context) {
		username := GetAuthenticatedUserFromContext(c)
		if err := api.validateAdminRequest(username); err != nil {
			FailNotAuthorized(c, UnAuthorizedAdminAccess)
			return
		}
		c.JSON(http.StatusOK, stats.Report())
	})

	// PROTECTED ROUTES -- BEGIN
	paymentsProtected := v1.Group("/payments", authware...)
	paymentsProtected.POST("/create", api.CreatePayment)
	paymentsProtected.GET("/deposit/address/:type", api.GetDepositAddress)

	accountProtected := v1.Group("/account", authware...)
	accountProtected.POST("password/change", api.changeAccountPassword)
	accountProtected.GET("/key/ipfs/get", api.getIPFSKeyNamesForAuthUser)
	accountProtected.POST("/key/ipfs/new", api.createIPFSKey)
	accountProtected.GET("/credits/available", api.getCredits)

	ipfsProtected := v1.Group("/ipfs", authware...)
	ipfsProtected.POST("/pubsub/publish/:topic", api.ipfsPubSubPublish)
	ipfsProtected.POST("/calculate-content-hash", api.calculateContentHashForFile)
	ipfsProtected.GET("/pins", api.getLocalPins) // admin locked
	ipfsProtected.GET("/object-stat/:key", api.getObjectStatForIpfs)
	ipfsProtected.GET("/check-for-pin/:hash", api.checkLocalNodeForPin) // admin locked
	ipfsProtected.POST("/download/:hash", api.downloadContentHash)
	ipfsProtected.POST("/pin/:hash", api.pinHashLocally)
	ipfsProtected.POST("/add-file", api.addFileLocally)
	ipfsProtected.POST("/add-file/advanced", api.addFileLocallyAdvanced)

	ipfsPrivateProtected := v1.Group("/ipfs-private", authware...)
	ipfsPrivateProtected.POST("/new/network", api.createHostedIPFSNetworkEntryInDatabase)                // admin locked
	ipfsPrivateProtected.GET("/network/:name", api.getIPFSPrivateNetworkByName)                          // admin locked
	ipfsPrivateProtected.POST("/ipfs/check-for-pin/:hash", api.checkLocalNodeForPinForHostedIPFSNetwork) // admin locked
	ipfsPrivateProtected.POST("/ipfs/object-stat/:key", api.getObjectStatForIpfsForHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/pubsub/publish/:topic", api.ipfsPubSubPublishToHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/pins", api.getLocalPinsForHostedIPFSNetwork) // admin locked
	ipfsPrivateProtected.GET("/networks", api.getAuthorizedPrivateNetworks)
	ipfsPrivateProtected.GET("/uploads/:network_name", api.getUploadsByNetworkName)
	ipfsPrivateProtected.POST("/ipfs/pin/:hash", api.pinToHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/ipfs/add-file", api.addFileToHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/ipfs/add-file/advanced", api.addFileToHostedIPFSNetworkAdvanced)
	ipfsPrivateProtected.POST("/ipns/publish/details", api.publishDetailedIPNSToHostedIPFSNetwork)

	ipnsProtected := v1.Group("/ipns", authware...)
	ipnsProtected.POST("/publish/details", api.publishToIPNSDetails)
	ipnsProtected.POST("/dnslink/aws/add", api.generateDNSLinkEntry) // admin locked
	ipnsProtected.GET("/records", api.getIPNSRecordsPublishedByUser)

	clusterProtected := v1.Group("/ipfs-cluster", authware...)
	clusterProtected.POST("/sync-errors-local", api.syncClusterErrorsLocally)          // admin locked
	clusterProtected.GET("/status-local-pin/:hash", api.getLocalStatusForClusterPin)   // admin locked
	clusterProtected.GET("/status-global-pin/:hash", api.getGlobalStatusForClusterPin) // admin locked
	clusterProtected.GET("/status-local", api.fetchLocalClusterStatus)                 // admin locked
	clusterProtected.POST("/pin/:hash", api.pinHashToCluster)

	databaseProtected := v1.Group("/database", authware...)
	databaseProtected.GET("/uploads", api.getUploadsFromDatabase)  // admin locked
	databaseProtected.GET("/uploads/:user", api.getUploadsForUser) // partial admin locked

	frontendProtected := v1.Group("/frontend", authware...)
	frontendProtected.GET("/cost/calculate/:hash/:holdtime", api.calculatePinCost)
	frontendProtected.POST("/cost/calculate/file", api.calculateFileCost)

	adminProtected := v1.Group("/admin", authware...)
	adminProtected.POST("/utils/file-size-check", CalculateFileSize)
	mini := adminProtected.Group("/mini")
	mini.POST("/create/bucket", api.makeBucket)
	// PROTECTED ROUTES -- END

	api.LogInfo("Routes initialized")
}
