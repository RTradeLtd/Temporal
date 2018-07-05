// Package api is the main package for Temporal's
// http api
package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/semihalev/gin-stats"

	"github.com/RTradeLtd/Temporal/api/middleware"
	"github.com/RTradeLtd/Temporal/database"
	jwt "github.com/appleboy/gin-jwt"
	helmet "github.com/danielkov/gin-helmet"
	"github.com/jinzhu/gorm"

	"github.com/aviddiviner/gin-limit"
	"github.com/dvwright/xss-mw"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/zsais/go-gin-prometheus"
)

var xssMdlwr xss.XssMw

// AdminAddress is the eth address of the admin account
var AdminAddress = "0xC6C35f43fDD71f86a2D8D4e3cA1Ce32564c38bd9"

//const experimental = true

// Setup is used to initialize our api.
// it invokes all  non exported function to setup the api.
func Setup(jwtKey, mqConnectionURL, dbPass, dbURL, ethKey, ethPass, listenAddress, dbUser, awsKey, awsSecret string) *gin.Engine {
	db, err := database.OpenDBConnection(dbPass, dbURL, dbUser)
	if err != nil {
		fmt.Println("failed to open db connection")
		log.Fatal(err)
	}
	db.LogMode(true)
	apiURL := fmt.Sprintf("%s:6768", listenAddress)
	r := gin.Default()
	r.Use(stats.RequestStats())
	r.Use(xssMdlwr.RemoveXss())
	r.Use(limit.MaxAllowed(20)) // limit to 20 con-current connections
	// create gin middleware instance for prom
	p := ginprometheus.NewPrometheus("gin")
	// set the address for prometheus to collect metrics
	p.SetListenAddress(apiURL)
	// load in prom to gin
	p.Use(r)
	// enable HSTS on all domains including subdomains
	r.Use(helmet.SetHSTS(true))
	// prevent mine content sniffing
	r.Use(helmet.NoSniff())
	r.Use(middleware.DatabaseMiddleware(db))
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = false
	r.Use(cors.New(corsConfig))
	/*corsConfig.AllowOrigins = []string{"http://null", "https://null"}
	corsConfig.AddAllowHeaders("X-Requested-With", "Access-Control-Allow-Headers", "Authorization", "Content-Type", "Set-Cookie", "X-Requested-With", "Accept", "Access-Control-Allow-Origin", "Access-Control-Request-Headers")
	r.Use(cors.New(corsConfig))*/
	authMiddleware := middleware.JwtConfigGenerate(jwtKey, db)

	setupRoutes(r, authMiddleware, db, awsKey, awsSecret, ethKey, ethPass, mqConnectionURL)

	statsProtected := r.Group("/api/v1/statistics")
	statsProtected.Use(authMiddleware.MiddlewareFunc())
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
	return r
}

// setupRoutes is used to setup all of our api routes
func setupRoutes(g *gin.Engine, authWare *jwt.GinJWTMiddleware, db *gorm.DB, awsKey, awsSecret, ethKey, ethPass, mqConnectionURL string) {
	//r.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	//r.Use(middleware.BlockchainMiddleware(true, ethKey, ethPass))
	// LOGIN
	g.POST("/api/v1/login", authWare.LoginHandler)

	// REGISTER
	g.POST("/api/v1/register", RegisterUserAccount)
	//g.POST("/api/v1/register-enterprise", RegisterEnterpriseUserAccount)

	// PROTECTED ROUTES -- BEGIN
	accountProtected := g.Group("/api/v1/account")
	accountProtected.Use(authWare.MiddlewareFunc())
	accountProtected.Use(middleware.APIRestrictionMiddleware(db))
	accountProtected.POST("password/change", ChangeAccountPassword)
	accountProtected.POST("/key/ipfs/new", CreateIPFSKey)
	accountProtected.GET("/key/ipfs/get", GetIPFSKeyNamesForAuthUser)

	ipfsProtected := g.Group("/api/v1/ipfs")
	ipfsProtected.Use(authWare.MiddlewareFunc())
	ipfsProtected.Use(middleware.APIRestrictionMiddleware(db))
	ipfsProtected.POST("/pubsub/publish/:topic", IpfsPubSubPublish) // admin locked
	ipfsProtected.GET("/pubsub/consume/:topic", IpfsPubSubConsume)  // admin locked
	ipfsProtected.GET("/pins", GetLocalPins)                        // admin locked
	ipfsProtected.GET("/object-stat/:key", GetObjectStatForIpfs)
	ipfsProtected.GET("/object/size/:key", GetFileSizeInBytesForObject)
	ipfsProtected.GET("/check-for-pin/:hash", CheckLocalNodeForPin)
	ipfsProtected.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	ipfsProtected.Use(middleware.BlockchainMiddleware(true, ethKey, ethPass))
	ipfsProtected.POST("/pin/:hash", PinHashLocally)
	ipfsProtected.POST("/add-file", AddFileLocally)

	//ipfsProtected.DELETE("/remove-pin/:hash", RemovePinFromLocalHost)

	ipfsPrivateProtected := g.Group("/api/v1/ipfs-private")
	ipfsPrivateProtected.Use(authWare.MiddlewareFunc())
	ipfsPrivateProtected.Use(middleware.APIRestrictionMiddleware(db))
	ipfsPrivateProtected.POST("/new/network", CreateHostedIPFSNetworkEntryInDatabase)
	ipfsPrivateProtected.POST("/network/name", GetIPFSPrivateNetworkByName)
	ipfsPrivateProtected.POST("/ipfs/check-for-pin/:hash", CheckLocalNodeForPinForHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/ipfs/object-stat/:key", GetObjectStatForIpfsForHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/ipfs/object/size/:key", GetFileSizeInBytesForObjectForHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/pubsub/publish/:topic", IpfsPubSubPublishToHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/pubsub/consume/:topic", IpfsPubSubConsumeForHostedIPFSNetwork)
	ipfsPrivateProtected.POST("/pins", GetLocalPinsForHostedIPFSNetwork)
	ipfsPrivateProtected.GET("/networks", GetAuthorizedPrivateNetworks)
	ipfsPrivateProtected.POST("/uploads", GetUploadsByNetworkName)
	ipfsPrivateProtected.DELETE("/pin/remove/:hash", RemovePinFromLocalHostForHostedIPFSNetwork)

	ipnsProtected := g.Group("/api/v1/ipns")
	ipnsProtected.Use(authWare.MiddlewareFunc())
	ipnsProtected.Use(middleware.APIRestrictionMiddleware(db))
	ipnsProtected.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	ipnsProtected.POST("/publish/details", PublishToIPNSDetails) // admin locked
	ipnsProtected.Use(middleware.AWSMiddleware(awsKey, awsSecret))
	ipnsProtected.POST("/dnslink/aws/add", GenerateDNSLinkEntry) // admin locked

	clusterProtected := g.Group("/api/v1/ipfs-cluster")
	clusterProtected.Use(authWare.MiddlewareFunc())
	clusterProtected.Use(middleware.APIRestrictionMiddleware(db))
	clusterProtected.POST("/sync-errors-local", SyncClusterErrorsLocally)          // admin locked
	clusterProtected.GET("/status-local-pin/:hash", GetLocalStatusForClusterPin)   // admin locked
	clusterProtected.GET("/status-global-pin/:hash", GetGlobalStatusForClusterPin) // admin locked
	clusterProtected.GET("/status-local", FetchLocalClusterStatus)                 // admin locked
	clusterProtected.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	clusterProtected.POST("/pin/:hash", PinHashToCluster)
	//clusterProtected.DELETE("/remove-pin/:hash", RemovePinFromCluster)

	databaseProtected := g.Group("/api/v1/database")
	databaseProtected.Use(authWare.MiddlewareFunc())
	databaseProtected.Use(middleware.APIRestrictionMiddleware(db))
	databaseProtected.DELETE("/garbage-collect/test", RunTestGarbageCollection)    // admin locked
	databaseProtected.DELETE("/garbage-collect/run", RunDatabaseGarbageCollection) // admin locked
	databaseProtected.GET("/uploads", GetUploadsFromDatabase)                      // admin locked
	databaseProtected.GET("/uploads/:address", GetUploadsForAddress)               // partial admin locked

	frontendProtected := g.Group("/api/v1/frontend/")
	frontendProtected.Use(authWare.MiddlewareFunc())
	frontendProtected.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	frontendProtected.Use(middleware.BlockchainMiddleware(true, ethKey, ethPass))
	frontendProtected.POST("/registration/request", SubmitPinPaymentRequest)
	frontendProtected.GET("/cost/calculate/:hash/:holdtime", CalculatePinCost)
	frontendProtected.POST("/confirm/:paymentID", ConfirmPayment)

	paymentsAPIProtected := g.Group("/api/v1/payments-api")
	paymentsAPIProtected.Use(authWare.MiddlewareFunc())
	paymentsAPIProtected.Use(middleware.APIRestrictionMiddleware(db))
	paymentsAPIProtected.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	paymentsAPIProtected.Use(middleware.BlockchainMiddleware(true, ethKey, ethPass))
	paymentsAPIProtected.POST("/register", RegisterPayment) // admin locked
	// PROTECTED ROUTES -- END

}
