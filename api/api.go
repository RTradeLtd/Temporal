// Package api is the main package for Temporal's
// http api
package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/RTradeLtd/Temporal/config"
	limit "github.com/aviddiviner/gin-limit"
	xss "github.com/dvwright/xss-mw"
	stats "github.com/semihalev/gin-stats"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	"github.com/RTradeLtd/Temporal/api/middleware"
	"github.com/RTradeLtd/Temporal/database"
	jwt "github.com/appleboy/gin-jwt"
	helmet "github.com/danielkov/gin-helmet"
	"github.com/jinzhu/gorm"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var xssMdlwr xss.XssMw

// AdminAddress is the eth address of the admin account
var AdminAddress = "0x7E4A2359c745A982a54653128085eAC69E446DE1"

// API is our API service
type API struct {
	Router  *gin.Engine
	TConfig *config.TemporalConfig
	DBM     *database.DatabaseManager
	Logger  *log.Logger
}

// Initialize is used ot initialize our API service
func Initialize(cfg *config.TemporalConfig, logMode bool) (*API, error) {
	// load config variables
	api := API{}
	// setup logging
	err := api.setupLogging()
	if err != nil {
		return nil, err
	}
	api.TConfig = cfg
	AdminAddress = cfg.API.AdminUser
	listenAddress := cfg.API.Connection.ListenAddress
	prometheusListenAddress := fmt.Sprintf("%s:6768", listenAddress)
	jwtKey := cfg.API.JwtKey
	// setup our database connection
	db, err := database.Initialize(cfg, false)
	if err != nil {
		return nil, err
	}
	api.DBM = db
	api.DBM.DB.LogMode(logMode)
	// generate our default router
	router := gin.Default()

	// load our global middlewares
	p := ginprometheus.NewPrometheus("gin")
	// set the address for prometheus to collect metrics
	p.SetListenAddress(prometheusListenAddress)
	// load in prom to gin
	p.Use(router)
	router.Use(xssMdlwr.RemoveXss())
	router.Use(limit.MaxAllowed(20)) // limit to 20 con-current connections
	// enable HSTS on all domains including subdomains
	router.Use(helmet.SetHSTS(true))
	// prevent mine content sniffing
	router.Use(helmet.NoSniff())
	//r.Use(middleware.DatabaseMiddleware(db))
	router.Use(middleware.CORSMiddleware())
	// generate our auth middleware to pass to setup routes
	authMiddleware := middleware.JwtConfigGenerate(jwtKey, db.DB, api.Logger)
	// setup our routes
	api.setupRoutes(router, authMiddleware, db.DB, cfg)
	api.Router = router

	return &api, nil
}

func (api *API) setupLogging() error {
	logFile, err := os.OpenFile("/var/log/temporal/api_service.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	logger := log.New()
	logger.Out = logFile
	api.Logger = logger
	api.Logger.Info("Logging initialized")
	return nil
}

// setupRoutes is used to setup all of our api routes
func (api *API) setupRoutes(g *gin.Engine, authWare *jwt.GinJWTMiddleware, db *gorm.DB, cfg *config.TemporalConfig) {

	statsProtected := g.Group("/api/v1/statistics")
	statsProtected.Use(authWare.MiddlewareFunc())
	statsProtected.Use(middleware.APIRestrictionMiddleware(db))
	statsProtected.Use(stats.RequestStats())
	statsProtected.GET("/stats", func(c *gin.Context) { // admin locked
		ethAddress := GetAuthenticatedUserFromContext(c)
		if ethAddress != AdminAddress {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "unauthorized access",
			})
			return
		}
		c.JSON(http.StatusOK, stats.Report())
	})

	auth := g.Group("/api/v1/auth")
	auth.POST("/register", api.registerUserAccount)
	auth.POST("/login", authWare.LoginHandler)

	// PROTECTED ROUTES -- BEGIN
	accountProtected := g.Group("/api/v1/account")
	accountProtected.Use(authWare.MiddlewareFunc())
	accountProtected.Use(middleware.APIRestrictionMiddleware(db))
	accountProtected.POST("password/change", api.changeAccountPassword)
	accountProtected.GET("/key/ipfs/get", api.getIPFSKeyNamesForAuthUser)
	accountProtected.POST("/key/ipfs/new", api.createIPFSKey)

	ipfsProtected := g.Group("/api/v1/ipfs")
	ipfsProtected.Use(authWare.MiddlewareFunc())
	ipfsProtected.Use(middleware.APIRestrictionMiddleware(db))
	ipfsProtected.POST("/pubsub/publish/:topic", api.ipfsPubSubPublish)
	ipfsProtected.POST("/calculate-content-hash", api.calculateContentHashForFile)
	ipfsProtected.GET("/pins", api.getLocalPins) // admin locked
	ipfsProtected.GET("/object-stat/:key", api.getObjectStatForIpfs)
	ipfsProtected.GET("/object/size/:key", api.getFileSizeInBytesForObject)
	ipfsProtected.GET("/check-for-pin/:hash", api.checkLocalNodeForPin) // admin locked
	ipfsProtected.POST("/download/:hash", api.downloadContentHash)
	ipfsProtected.POST("/pin/:hash", api.pinHashLocally)
	ipfsProtected.POST("/add-file", api.addFileLocally)
	ipfsProtected.POST("/add-file/advanced", api.addFileLocallyAdvanced)
	ipfsProtected.DELETE("/remove-pin/:hash", api.removePinFromLocalHost) // admin locked

	ipfsPrivateProtected := g.Group("/api/v1/ipfs-private")
	ipfsPrivateProtected.Use(authWare.MiddlewareFunc())
	ipfsPrivateProtected.Use(middleware.APIRestrictionMiddleware(db))
	ipfsPrivateProtected.POST("/new/network", api.createHostedIPFSNetworkEntryInDatabase)                // admin locked
	ipfsPrivateProtected.GET("/network/:name", api.getIPFSPrivateNetworkByName)                          // admin locked
	ipfsPrivateProtected.POST("/ipfs/check-for-pin/:hash", api.checkLocalNodeForPinForHostedIPFSNetwork) // admin locked
	ipfsPrivateProtected.POST("/ipfs/object-stat/:key", api.getObjectStatForIpfsForHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/ipfs/object/size/:key", api.getFileSizeInBytesForObjectForHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/pubsub/publish/:topic", api.ipfsPubSubPublishToHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/pins", api.getLocalPinsForHostedIPFSNetwork) // admin locked
	ipfsPrivateProtected.GET("/networks", api.getAuthorizedPrivateNetworks)
	ipfsPrivateProtected.POST("/uploads", api.getUploadsByNetworkName)
	ipfsPrivateProtected.POST("/ipfs/pin/:hash", api.pinToHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/ipfs/add-file", api.addFileToHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/ipfs/add-file/advanced", api.addFileToHostedIPFSNetworkAdvanced)
	ipfsPrivateProtected.POST("/ipns/publish/details", api.publishDetailedIPNSToHostedIPFSNetwork)
	ipfsPrivateProtected.DELETE("/ipfs/pin/remove/:hash", api.removePinFromLocalHostForHostedIPFSNetwork)

	ipnsProtected := g.Group("/api/v1/ipns")
	ipnsProtected.Use(authWare.MiddlewareFunc())
	ipnsProtected.Use(middleware.APIRestrictionMiddleware(db))
	ipnsProtected.POST("/publish/details", api.publishToIPNSDetails)
	ipnsProtected.POST("/dnslink/aws/add", api.generateDNSLinkEntry) // admin locked

	clusterProtected := g.Group("/api/v1/ipfs-cluster")
	clusterProtected.Use(authWare.MiddlewareFunc())
	clusterProtected.Use(middleware.APIRestrictionMiddleware(db))
	clusterProtected.POST("/sync-errors-local", api.syncClusterErrorsLocally)          // admin locked
	clusterProtected.GET("/status-local-pin/:hash", api.getLocalStatusForClusterPin)   // admin locked
	clusterProtected.GET("/status-global-pin/:hash", api.getGlobalStatusForClusterPin) // admin locked
	clusterProtected.GET("/status-local", api.fetchLocalClusterStatus)                 // admin locked
	clusterProtected.POST("/pin/:hash", api.pinHashToCluster)
	clusterProtected.DELETE("/remove-pin/:hash", api.removePinFromCluster) // admin locked

	databaseProtected := g.Group("/api/v1/database")
	databaseProtected.Use(authWare.MiddlewareFunc())
	databaseProtected.Use(middleware.APIRestrictionMiddleware(db))
	databaseProtected.GET("/uploads", api.getUploadsFromDatabase)     // admin locked
	databaseProtected.GET("/uploads/:user", api.getUploadsForAddress) // partial admin locked

	frontendProtected := g.Group("/api/v1/frontend/")
	frontendProtected.Use(authWare.MiddlewareFunc())
	frontendProtected.POST("/utils/ipfs/hash/calculate", api.calculateIPFSFileHash)
	frontendProtected.GET("/cost/calculate/:hash/:holdtime", api.calculatePinCost)
	frontendProtected.POST("/cost/calculate/file", api.calculateFileCost)
	frontendProtected.POST("/payment/pin/confirm/:hash", api.submitPinPaymentConfirmation)
	frontendProtected.POST("/payment/pin/create/:hash", api.createPinPayment)
	frontendProtected.POST("/payment/pin/confirm", api.submitPinPaymentConfirmation)
	frontendProtected.POST("/payment/pin/submit/:hash", api.submitPaymentToContract)
	frontendProtected.POST("/payment/file/create", api.createFilePayment)

	adminProtected := g.Group("/api/v1/admin")
	adminProtected.Use(authWare.MiddlewareFunc())
	adminProtected.Use(middleware.APIRestrictionMiddleware(db))
	adminProtected.POST("/utils/file-size-check", CalculateFileSize)
	mini := adminProtected.Group("/mini")
	mini.POST("/create/bucket", api.makeBucket)
	// PROTECTED ROUTES -- END

	api.Logger.Info("Routes initialized")
}
