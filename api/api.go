// Package api is the main package for Temporal's
// http api
package api

import (
	"fmt"
	"net/http"

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
)

var xssMdlwr xss.XssMw

// AdminAddress is the eth address of the admin account
var AdminAddress = "0x7E4A2359c745A982a54653128085eAC69E446DE1"

// API is our API service
type API struct {
	Router  *gin.Engine
	TConfig *config.TemporalConfig
	DBM     *database.DatabaseManager
}

// Initialize is used ot initialize our API service
func Initialize(cfg *config.TemporalConfig, logMode bool) (*API, error) {
	// load config variables
	api := API{}
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

	// load our middleware
	router.Use(stats.RequestStats())
	router.Use(xssMdlwr.RemoveXss())
	router.Use(limit.MaxAllowed(20)) // limit to 20 con-current connections
	// create gin middleware instance for prom
	p := ginprometheus.NewPrometheus("gin")
	// set the address for prometheus to collect metrics
	p.SetListenAddress(prometheusListenAddress)
	// load in prom to gin
	p.Use(router)
	// enable HSTS on all domains including subdomains
	router.Use(helmet.SetHSTS(true))
	// prevent mine content sniffing
	router.Use(helmet.NoSniff())
	//r.Use(middleware.DatabaseMiddleware(db))
	router.Use(middleware.CORSMiddleware())
	// generate our auth middleware to pass to setup routes
	authMiddleware := middleware.JwtConfigGenerate(jwtKey, db.DB)
	// setup our routes
	api.setupRoutes(router, authMiddleware, db.DB, cfg)
	api.Router = router
	return &api, nil
}

// setupRoutes is used to setup all of our api routes
func (api *API) setupRoutes(g *gin.Engine, authWare *jwt.GinJWTMiddleware, db *gorm.DB, cfg *config.TemporalConfig) {

	mqConnectionURL := cfg.RabbitMQ.URL
	ethKey := cfg.Ethereum.Account.KeyFile
	ethPass := cfg.Ethereum.Account.KeyPass
	awsKey := cfg.AWS.KeyID
	awsSecret := cfg.AWS.Secret
	endpoint := fmt.Sprintf("%s:%s", cfg.MINIO.Connection.IP, cfg.MINIO.Connection.Port)
	minioKey := cfg.MINIO.AccessKey
	minioSecret := cfg.MINIO.SecretKey

	statsProtected := g.Group("/api/v1/statistics")
	statsProtected.Use(authWare.MiddlewareFunc())
	statsProtected.Use(middleware.APIRestrictionMiddleware(db))
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

	// LOGIN
	g.Use(middleware.DatabaseMiddleware(db))
	g.POST("/api/v1/login", authWare.LoginHandler)
	// REGISTER
	g.POST("/api/v1/register", registerUserAccount)
	//g.POST("/api/v1/register-enterprise", RegisterEnterpriseUserAccount)

	// PROTECTED ROUTES -- BEGIN
	accountProtected := g.Group("/api/v1/account")
	accountProtected.Use(authWare.MiddlewareFunc())
	accountProtected.Use(middleware.APIRestrictionMiddleware(db))
	accountProtected.Use(middleware.DatabaseMiddleware(db))
	accountProtected.POST("password/change", changeAccountPassword)
	accountProtected.GET("/key/ipfs/get", getIPFSKeyNamesForAuthUser)
	accountProtected.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	accountProtected.POST("/key/ipfs/new", createIPFSKey)

	ipfsProtected := g.Group("/api/v1/ipfs")
	ipfsProtected.Use(authWare.MiddlewareFunc())
	ipfsProtected.Use(middleware.APIRestrictionMiddleware(db))
	// DATABASE-LESS routes
	ipfsProtected.POST("/pubsub/publish/:topic", ipfsPubSubPublish)
	ipfsProtected.POST("/calculate-content-hash", calculateContentHashForFile)
	ipfsProtected.GET("/pins", getLocalPins) // admin locked
	ipfsProtected.GET("/object-stat/:key", getObjectStatForIpfs)
	ipfsProtected.GET("/object/size/:key", getFileSizeInBytesForObject)
	ipfsProtected.GET("/check-for-pin/:hash", checkLocalNodeForPin) // admin locked
	ipfsProtected.Use(middleware.DatabaseMiddleware(db))
	ipfsProtected.POST("/download/:hash", downloadContentHash)

	// DATABASE-USING ROUTES
	ipfsProtected.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	ipfsProtected.Use(middleware.DatabaseMiddleware(db))
	ipfsProtected.POST("/pin/:hash", pinHashLocally)
	ipfsProtected.POST("/add-file", addFileLocally)
	ipfsProtected.Use(middleware.MINIMiddleware(minioKey, minioSecret, endpoint, true))
	ipfsProtected.POST("/add-file/advanced", addFileLocallyAdvanced)

	//ipfsProtected.DELETE("/remove-pin/:hash", RemovePinFromLocalHost)

	ipfsPrivateProtected := g.Group("/api/v1/ipfs-private")
	ipfsPrivateProtected.Use(authWare.MiddlewareFunc())
	ipfsPrivateProtected.Use(middleware.APIRestrictionMiddleware(db))
	ipfsPrivateProtected.Use(middleware.DatabaseMiddleware(db))
	ipfsPrivateProtected.POST("/new/network", CreateHostedIPFSNetworkEntryInDatabase)                // admin locked
	ipfsPrivateProtected.GET("/network/:name", GetIPFSPrivateNetworkByName)                          // admin locked
	ipfsPrivateProtected.POST("/ipfs/check-for-pin/:hash", CheckLocalNodeForPinForHostedIPFSNetwork) // admin locked
	ipfsPrivateProtected.POST("/ipfs/object-stat/:key", GetObjectStatForIpfsForHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/ipfs/object/size/:key", GetFileSizeInBytesForObjectForHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/pubsub/publish/:topic", IpfsPubSubPublishToHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/pins", GetLocalPinsForHostedIPFSNetwork) // admin locked
	ipfsPrivateProtected.GET("/networks", GetAuthorizedPrivateNetworks)
	ipfsPrivateProtected.POST("/uploads", GetUploadsByNetworkName)
	ipfsPrivateProtected.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	ipfsPrivateProtected.POST("/ipfs/pin/:hash", PinToHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/ipfs/add-file", AddFileToHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/ipfs/add-file/advanced", AddFileToHostedIPFSNetworkAdvanced)
	ipfsPrivateProtected.DELETE("/ipfs/pin/remove/:hash", RemovePinFromLocalHostForHostedIPFSNetwork) // admin locked

	ipnsProtected := g.Group("/api/v1/ipns")
	ipnsProtected.Use(authWare.MiddlewareFunc())
	ipnsProtected.Use(middleware.APIRestrictionMiddleware(db))
	ipnsProtected.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	ipnsProtected.Use(middleware.DatabaseMiddleware(db))
	ipnsProtected.POST("/publish/details", publishToIPNSDetails)
	ipnsProtected.Use(middleware.AWSMiddleware(awsKey, awsSecret))
	ipnsProtected.POST("/dnslink/aws/add", generateDNSLinkEntry) // admin locked

	clusterProtected := g.Group("/api/v1/ipfs-cluster")
	clusterProtected.Use(authWare.MiddlewareFunc())
	clusterProtected.Use(middleware.APIRestrictionMiddleware(db))
	clusterProtected.POST("/sync-errors-local", syncClusterErrorsLocally)          // admin locked
	clusterProtected.GET("/status-local-pin/:hash", getLocalStatusForClusterPin)   // admin locked
	clusterProtected.GET("/status-global-pin/:hash", getGlobalStatusForClusterPin) // admin locked
	clusterProtected.GET("/status-local", fetchLocalClusterStatus)                 // admin locked
	clusterProtected.Use(middleware.RabbitMQMiddleware(mqConnectionURL))           // admin locked
	clusterProtected.POST("/pin/:hash", pinHashToCluster)
	clusterProtected.DELETE("/remove-pin/:hash", removePinFromCluster) // admin locked

	databaseProtected := g.Group("/api/v1/database")
	databaseProtected.Use(authWare.MiddlewareFunc())
	databaseProtected.Use(middleware.APIRestrictionMiddleware(db))
	databaseProtected.Use(middleware.DatabaseMiddleware(db))
	databaseProtected.GET("/uploads", getUploadsFromDatabase)     // admin locked
	databaseProtected.GET("/uploads/:user", getUploadsForAddress) // partial admin locked

	frontendProtected := g.Group("/api/v1/frontend/")
	frontendProtected.Use(authWare.MiddlewareFunc())
	frontendProtected.GET("/cost/calculate/:hash/:holdtime", calculatePinCost)
	frontendProtected.POST("/cost/calculate/file", calculateFileCost)
	frontendProtected.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	frontendProtected.Use(middleware.BlockchainMiddleware(true, ethKey, ethPass))
	frontendProtected.Use(middleware.DatabaseMiddleware(db))
	frontendProtected.POST("/payment/pin/confirm/:hash", submitPinPaymentConfirmation)
	frontendProtected.POST("/payment/pin/create/:hash", createPinPayment)
	frontendProtected.POST("/payment/pin/confirm", submitPinPaymentConfirmation)
	frontendProtected.POST("/payment/pin/submit/:hash", submitPaymentToContract)
	frontendProtected.Use(middleware.MINIMiddleware(minioKey, minioSecret, endpoint, true))
	frontendProtected.POST("/payment/file/create", createFilePayment)

	adminProtected := g.Group("/api/v1/admin")
	adminProtected.Use(authWare.MiddlewareFunc())
	adminProtected.Use(middleware.APIRestrictionMiddleware(db))
	adminProtected.POST("/utils/file-size-check", CalculateFileSize)
	mini := adminProtected.Group("/mini")
	mini.Use(middleware.MINIMiddleware(minioKey, minioSecret, endpoint, true))
	mini.POST("/create/bucket", makeBucket)
	// PROTECTED ROUTES -- END

}
