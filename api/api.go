// Package api is the main package for Temporal's
// http api
package api

import (
	"github.com/RTradeLtd/Temporal/api/middleware"
	"github.com/RTradeLtd/Temporal/database"
	jwt "github.com/appleboy/gin-jwt"
	helmet "github.com/danielkov/gin-helmet"

	"github.com/aviddiviner/gin-limit"
	"github.com/dvwright/xss-mw"

	"github.com/gin-contrib/rollbar"
	"github.com/gin-gonic/gin"
	"github.com/stvp/roll"
	"github.com/zsais/go-gin-prometheus"
)

var xssMdlwr xss.XssMw
var AdminAddress string

// Setup is used to initialize our api.
// it invokes all  non exported function to setup the api.
func Setup(adminUser, adminPass, jwtKey, rollbarToken, mqConnectionURL, dbPass string) *gin.Engine {
	db := database.OpenDBConnection(dbPass)

	roll.Token = rollbarToken
	roll.Environment = "development"
	r := gin.Default()
	r.Use(xssMdlwr.RemoveXss())
	r.Use(limit.MaxAllowed(20)) // limit to 20 con-current connections
	// create gin middleware instance for prom
	p := ginprometheus.NewPrometheus("gin")
	// set the address for prometheus to collect metrics
	p.SetListenAddress("127.0.0.1:6768")
	// load in prom to gin
	p.Use(r)
	// enable HSTS on all domains including subdomains
	r.Use(helmet.SetHSTS(true))
	// prevent mine content sniffing
	r.Use(helmet.NoSniff())
	r.Use(rollbar.Recovery(false))
	r.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	r.Use(middleware.DatabaseMiddleware(dbPass))
	authMiddleware := middleware.JwtConfigGenerate(jwtKey, db)

	setupRoutes(r, adminUser, adminPass, authMiddleware)
	return r
}

// setupRoutes is used to setup all of our api routes
func setupRoutes(g *gin.Engine, adminUser string, adminPass string, authWare *jwt.GinJWTMiddleware) {

	// LOGIN
	g.POST("/api/v1/login", authWare.LoginHandler)

	// REGISTER
	g.POST("/api/v1/register", RegisterUserAccount)
	g.POST("/api/v1/register-enterprise", RegisterEnterpriseUserAccount)

	apiV1 := g.Group("/api/v1")
	apiV1.Use(authWare.MiddlewareFunc())

	// PROTECTED ROUTES -- BEGIN
	ipfsProtected := g.Group("/api/v1/ipfs")
	ipfsProtected.Use(authWare.MiddlewareFunc())
	ipfsProtected.POST("/pin/:hash", PinHashLocally)
	ipfsProtected.POST("/add-file", AddFileLocally)
	ipfsProtected.DELETE("/remove-pin/:hash", RemovePinFromLocalHost)
	clusterProtected := g.Group("/api/v1/ipfs-cluster")
	clusterProtected.Use(authWare.MiddlewareFunc())
	clusterProtected.POST("/pin/:hash", PinHashToCluster)
	clusterProtected.POST("/sync-errors-local", SyncClusterErrorsLocally)
	clusterProtected.DELETE("/remove-pin/:hash", RemovePinFromCluster)
	databaseProtected := g.Group("/api/v1/database")
	databaseProtected.Use(authWare.MiddlewareFunc())
	databaseProtected.DELETE("/api/v1/database/garbage-collect/test", RunTestGarbageCollection)
	databaseProtected.GET("/uploads", GetUploadsFromDatabase)
	databaseProtected.GET("/uploads/:address", GetUploadsForAddress)
	paymentsProtected := g.Group("/api/v1/payments")
	paymentsProtected.Use(authWare.MiddlewareFunc())
	paymentsProtected.POST("/rtc/register", RegisterRtcPayment)
	// PROTECTED ROUTES -- END

	// IPFS ROUTES [POST] -- BEGIN
	g.POST("/api/v1/ipfs/pubsub/publish/:topic", IpfsPubSubPublish)
	// IPFS ROUTES [POST] -- END

	// IPFS ROUTES [GET] -- BEGIN
	g.GET("/api/v1/ipfs/pubsub/consume/:topic", IpfsPubSubConsume)
	g.GET("/api/v1/ipfs/pins", GetLocalPins)
	g.GET("/api/v1/ipfs/object-stat/:key", GetObjectStatForIpfs)
	g.GET("/api/v1/ipfs/object/size/:key", GetFileSizeInBytesForObject)
	g.GET("/api/v1/ipfs/check-for-pin/:hash", CheckLocalNodeForPin)
	// IPFS ROUTES [GET] -- END

	// IPFS CLUSTER ROUTES [GET] -- BEGIN
	g.GET("/api/v1/ipfs-cluster/status-local-pin/:hash", GetLocalStatusForClusterPin)
	g.GET("/api/v1/ipfs-cluster/status-global-pin/:hash", GetGlobalStatusForClusterPin)
	g.GET("/api/v1/ipfs-cluster/status-local", FetchLocalClusterStatus)
	// IPFS CLUSTER ROUTES [GET] -- END

}
