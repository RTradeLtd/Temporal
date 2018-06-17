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

	"github.com/gin-contrib/rollbar"
	"github.com/gin-gonic/gin"
	"github.com/stvp/roll"
	"github.com/zsais/go-gin-prometheus"
)

var xssMdlwr xss.XssMw

// AdminAddress is the eth address of the admin account
var AdminAddress = "0xC6C35f43fDD71f86a2D8D4e3cA1Ce32564c38bd9"

const experimental = true

// Setup is used to initialize our api.
// it invokes all  non exported function to setup the api.
func Setup(jwtKey, rollbarToken, mqConnectionURL, dbPass, dbURL, ethKey, ethPass, listenAddress, dbUser string) *gin.Engine {
	db, err := database.OpenDBConnection(dbPass, dbURL, dbUser)
	if err != nil {
		fmt.Println("failed to open db connection")
		log.Fatal(err)
	}

	apiURL := fmt.Sprintf("%s:6768", listenAddress)
	roll.Token = rollbarToken
	roll.Environment = "development"
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
	r.Use(rollbar.Recovery(false))
	r.Use(middleware.RabbitMQMiddleware(mqConnectionURL))
	r.Use(middleware.DatabaseMiddleware(db))
	r.Use(middleware.BlockchainMiddleware(true, ethKey, ethPass))
	authMiddleware := middleware.JwtConfigGenerate(jwtKey, db)

	setupRoutes(r, authMiddleware, db)

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
func setupRoutes(g *gin.Engine, authWare *jwt.GinJWTMiddleware, db *gorm.DB) {

	// LOGIN
	g.POST("/api/v1/login", authWare.LoginHandler)

	// REGISTER
	g.POST("/api/v1/register", RegisterUserAccount)
	//g.POST("/api/v1/register-enterprise", RegisterEnterpriseUserAccount)

	// PROTECTED ROUTES -- BEGIN

	accountProtected := g.Group("/api/v1/account")
	accountProtected.Use(authWare.MiddlewareFunc())
	accountProtected.POST("password/change", ChangeAccountPassword)

	ipfsProtected := g.Group("/api/v1/ipfs")
	ipfsProtected.Use(authWare.MiddlewareFunc())
	ipfsProtected.Use(middleware.APIRestrictionMiddleware(db))
	ipfsProtected.POST("/pubsub/publish/:topic", IpfsPubSubPublish) // admin locked
	ipfsProtected.POST("/pin/:hash", PinHashLocally)
	ipfsProtected.POST("/add-file", AddFileLocally)
	ipfsProtected.GET("/pubsub/consume/:topic", IpfsPubSubConsume) // admin locked
	ipfsProtected.GET("/pins", GetLocalPins)                       // admin locked
	ipfsProtected.GET("/object-stat/:key", GetObjectStatForIpfs)
	ipfsProtected.GET("/object/size/:key", GetFileSizeInBytesForObject)
	ipfsProtected.GET("/check-for-pin/:hash", CheckLocalNodeForPin)
	//ipfsProtected.DELETE("/remove-pin/:hash", RemovePinFromLocalHost)

	ipnsProtected := g.Group("/api/v1/ipns")
	ipnsProtected.Use(authWare.MiddlewareFunc())
	ipnsProtected.Use(middleware.APIRestrictionMiddleware(db))
	ipnsProtected.GET("/publish/:hash", PublishToIPNS)           // admin locked
	ipnsProtected.POST("/publish/details", PublishToIPNSDetails) // admin locked

	clusterProtected := g.Group("/api/v1/ipfs-cluster")
	clusterProtected.Use(authWare.MiddlewareFunc())
	clusterProtected.Use(middleware.APIRestrictionMiddleware(db))
	clusterProtected.POST("/pin/:hash", PinHashToCluster)
	clusterProtected.POST("/sync-errors-local", SyncClusterErrorsLocally)          // admin locked
	clusterProtected.GET("/status-local-pin/:hash", GetLocalStatusForClusterPin)   // admin locked
	clusterProtected.GET("/status-global-pin/:hash", GetGlobalStatusForClusterPin) // admin locked
	clusterProtected.GET("/status-local", FetchLocalClusterStatus)                 // admin locked
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
	frontendProtected.POST("/registration/request", SubmitPinPaymentRequest)
	frontendProtected.GET("/cost/calculate/:hash/:holdtime", CalculatePinCost)
	frontendProtected.POST("/confirm/:paymentID", ConfirmPayment)

	paymentsAPIProtected := g.Group("/api/v1/payments-api")
	paymentsAPIProtected.Use(authWare.MiddlewareFunc())
	paymentsAPIProtected.Use(middleware.APIRestrictionMiddleware(db))
	paymentsAPIProtected.POST("/register", RegisterPayment) // admin locked
	// PROTECTED ROUTES -- END

	/*	if experimental {
			tusHandler, err := generateTUSHandler()
			if err != nil {
				log.Fatal(err)
			}

			tusProtected := g.Group("/api/v1/tus")
			tusProtected.Use(authWare.MiddlewareFunc())
			tusProtected.Use(middleware.APIRestrictionMiddleware(db))
			tusProtected.Any("/files", gin.WrapH(tusHandler))
			tusProtected.GET("/metrics", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"metrics": tusHandler.Metrics,
				})
			})
		}
	*/
}
