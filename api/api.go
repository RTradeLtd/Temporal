// Package api is the main package for Temporal's
// http api
package api

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/RTradeLtd/Temporal/api/middleware"
	"github.com/RTradeLtd/Temporal/api/rtfs_cluster"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	jwt "github.com/appleboy/gin-jwt"
	helmet "github.com/danielkov/gin-helmet"

	"github.com/aviddiviner/gin-limit"
	"github.com/dvwright/xss-mw"
	gocid "github.com/ipfs/go-cid"

	"time"

	"github.com/RTradeLtd/Temporal/api/rtfs"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/gin-contrib/rollbar"
	"github.com/gin-gonic/gin"
	"github.com/stvp/roll"
	"github.com/zsais/go-gin-prometheus"
)

var xssMdlwr xss.XssMw
var realmName = "temporal-realm"

// Setup is used to initialize our api.
// it invokes all  non exported function to setup the api.
func Setup(adminUser, adminPass, jwtKey, rollbarToken, mqConnectionURL string) *gin.Engine {

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
	// read 1000 random numbers, used to help randomnize the JWT
	c := 1000
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error generating random number")
		os.Exit(1)
	}
	// see appleboy package example, slightly modified
	// will implement metamaks/msg signing with ethereum accounts
	// as the authentication metho
	authMiddleware := &jwt.GinJWTMiddleware{
		Realm:      realmName,
		Key:        []byte(fmt.Sprintf("%v+%s", b, jwtKey)),
		Timeout:    time.Hour * 24,
		MaxRefresh: time.Hour * 24,
		Authenticator: func(userId string, password string, c *gin.Context) (string, bool) {
			if userId == adminUser && password == adminPass {
				return userId, true
			}

			return userId, false
		},
		Authorizator: func(userId string, c *gin.Context) bool {
			if userId == adminUser {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},

		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	}

	setupRoutes(r, adminUser, adminPass, authMiddleware)
	return r
}

// setupRoutes is used to setup all of our api routes
func setupRoutes(g *gin.Engine, adminUser string, adminPass string, authWare *jwt.GinJWTMiddleware) {

	// LOGIN
	g.POST("/api/v1/login", authWare.LoginHandler)

	apiV1 := g.Group("/api/v1")
	apiV1.Use(authWare.MiddlewareFunc())
	apiV1.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "hello"})
	})
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
	// PROTECTED ROUTES -- END

	// IPFS ROUTES [POST] -- BEGIN
	g.POST("/api/v1/ipfs/pubsub/publish/:topic", IpfsPubSubPublish)
	g.POST("/api/v1/ipfs/pubsub/publish-test/:topic", IpfsPubSubTest)
	// IPFS ROUTES [POST] -- END

	// IPFS ROUTES [GET] -- BEGIN
	g.GET("/api/v1/ipfs/pubsub/consume/:topic", IpfsPubSubConsume)
	g.GET("/api/v1/ipfs/pins", GetLocalPins)
	g.GET("/api/v1/ipfs/object-stat/:key", GetObjectStatForIpfs)
	g.GET("/api/v1/ipfs/check-for-pin/:hash", CheckLocalNodeForPin)
	// IPFS ROUTES [GET] -- END

	// IPFS CLUSTER ROUTES [GET] -- BEGIN
	g.GET("/api/v1/ipfs-cluster/status-local-pin/:hash", GetLocalStatusForClusterPin)
	g.GET("/api/v1/ipfs-cluster/status-global-pin/:hash", GetGlobalStatusForClusterPin)
	g.GET("/api/v1/ipfs-cluster/status-local", FetchLocalClusterStatus)
	// IPFS CLUSTER ROUTES [GET] -- END

	// DATABASE ROUTES [GET] -- BEGIN
	g.GET("/api/v1/database/uploads", GetUploadsFromDatabase)
	g.GET("/api/v1/database/uploads/:address", GetUploadsForAddress)
	// DATABASE ROUTES [GET] -- END
}

// PinHashLocally is used to pin a hash to the local ipfs node
func PinHashLocally(c *gin.Context) {
	contextCopy := c.Copy()
	hash := contextCopy.Param("hash")
	uploadAddress := contextCopy.PostForm("uploadAddress")
	holdTimeInMonths := contextCopy.PostForm("holdTime")
	holdTimeInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		c.Error(err)
		fmt.Println(err)
		return
	}
	// construct the rabbitmq message to add this entry to the database
	dpa := queue.DatabasePinAdd{
		Hash:             hash,
		UploaderAddress:  uploadAddress,
		HoldTimeInMonths: holdTimeInt,
	}
	// assert type assertion retrieving info from middleware
	mqConnectionURL := c.MustGet("mq_conn_url").(string)
	// initialize the queue
	qm, err := queue.Initialize(queue.DatabasePinAddQueue, mqConnectionURL)
	if err != nil {
		c.Error(err)
		fmt.Println(err)

		return
	}
	// publish the message, if there was an error finish processing
	err = qm.PublishMessage(dpa)
	if err != nil {
		c.Error(err)
		fmt.Println(err)
		return
	}
	qm.Close()
	go func() {
		// currently after it is pinned, it is sent to the cluster to be pinned
		manager := rtfs.Initialize("")
		// before exiting, it is pinned to the cluster
		err = manager.Pin(hash)
		if err != nil {
			fmt.Println(err)
		}
	}()
	c.JSON(http.StatusOK, gin.H{"upload": dpa})
}

// AddFileLocally is used to add a file to our local ipfs node
// this will have to be done first before pushing any file's to the cluster
// this needs to be optimized so that the process doesn't "hang" while uploading
func AddFileLocally(c *gin.Context) {
	fmt.Println("fetching file")
	// fetch the file, and create a handler to interact with it
	fileHandler, err := c.FormFile("file")
	if err != nil {
		c.Error(err)
		return
	}
	uploaderAddress := c.PostForm("uploaderAddress")
	holdTimeinMonths := c.PostForm("holdTime")
	holdTimeinMonthsInt, err := strconv.ParseInt(holdTimeinMonths, 10, 64)
	if err != nil {
		c.Error(err)
		return
	}
	fmt.Println("opening file")
	// open the file
	openFile, err := fileHandler.Open()
	if err != nil {
		c.Error(err)
		return
	}
	fmt.Println("initializing manager")
	// initialize a connection to the local ipfs node
	manager := rtfs.Initialize("")
	// pin the file
	fmt.Println("adding file")
	resp, err := manager.Shell.Add(openFile)
	if err != nil {
		c.Error(err)
		return
	}
	fmt.Println("file added")
	// construct a message to rabbitmq to upad the database
	dfa := queue.DatabaseFileAdd{
		Hash:             resp,
		HoldTimeInMonths: holdTimeinMonthsInt,
		UploaderAddress:  uploaderAddress,
	}
	mqConnectionURL := c.MustGet("mq_conn_url").(string)
	// initialize a connectino to rabbitmq
	qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqConnectionURL)
	if err != nil {
		c.Error(err)
		return
	}
	// publish the message
	err = qm.PublishMessage(dfa)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": resp})
}

// PinHashToCluster is used to pin a hash to the local ipfs node
func PinHashToCluster(c *gin.Context) {
	contextCopy := c.Copy()
	hash := contextCopy.Param("hash")
	uploadAddress := contextCopy.PostForm("uploadAddress")
	holdTimeInMonths := contextCopy.PostForm("holdTime")
	holdTimeInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		c.Error(err)
		fmt.Println(err)
		return
	}
	// construct the rabbitmq message to add this entry to the database
	dpa := queue.DatabasePinAdd{
		Hash:             hash,
		UploaderAddress:  uploadAddress,
		HoldTimeInMonths: holdTimeInt,
	}
	// assert type assertion retrieving info from middleware
	mqConnectionURL := c.MustGet("mq_conn_url").(string)
	// initialize the queue
	qm, err := queue.Initialize(queue.DatabasePinAddQueue, mqConnectionURL)
	if err != nil {
		c.Error(err)
		fmt.Println(err)

		return
	}
	// publish the message, if there was an error finish processing
	err = qm.PublishMessage(dpa)
	if err != nil {
		c.Error(err)
		fmt.Println(err)
		return
	}
	qm.Close()
	go func() {
		// currently after it is pinned, it is sent to the cluster to be pinned
		manager := rtfs_cluster.Initialize()
		decodedHash := manager.DecodeHashString(hash)
		// before exiting, it is pinned to the cluster
		err = manager.Pin(decodedHash)
		if err != nil {
			fmt.Println(err)
		}
	}()
	c.JSON(http.StatusOK, gin.H{"upload": dpa})
}

// SyncClusterErrorsLocally is used to parse through the local cluster state
// and sync any errors that are detected.
func SyncClusterErrorsLocally(c *gin.Context) {
	// initialize a conection to the cluster
	manager := rtfs_cluster.Initialize()
	// parse the local cluster status, and sync any errors, retunring the cids that were in an error state
	syncedCids, err := manager.ParseLocalStatusAllAndSync()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"synced-cids": syncedCids})
}

// IpfsPubSubPublish is used to publish a pubsub msg
func IpfsPubSubPublish(c *gin.Context) {
	topic := c.Param("topic")
	message := c.PostForm("message")
	manager := rtfs.Initialize("")
	err := manager.PublishPubSubMessage(topic, message)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"topic": topic, "message": message})
}

// IpfsPubSubTest runs a pubsub test
func IpfsPubSubTest(c *gin.Context) {
	manager := rtfs.Initialize("")
	err := manager.PublishPubSubTest(manager.PubTopic)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, nil)
}

// IpfsPubSubConsume is used to consume pubsub messages
func IpfsPubSubConsume(c *gin.Context) {
	contextCopy := c.Copy()
	topic := contextCopy.Param("topic")

	go func() {
		manager := rtfs.Initialize("")
		manager.SubscribeToPubSubTopic(topic)
		manager.ConsumeSubscription(manager.PubSub)
	}()

	c.JSON(http.StatusOK, gin.H{"status": "consuming messages in background"})
}

// RemovePinFromLocalHost is used to remove a pin from the ipfs instance
// TODO: fully implement
func RemovePinFromLocalHost(c *gin.Context) {
	contextCopy := c.Copy()
	// fetch hash param
	hash := contextCopy.Param("hash")

	go func() {
		// initialise a connetion to the local ipfs node
		manager := rtfs.Initialize("")
		// remove the file from the local ipfs state
		// TODO: implement some kind of error handling and notification
		manager.Shell.Unpin(hash)
	}()
	// TODO:
	// change to send a message to the cluster to depin
	mqConnectionURL := c.MustGet("mq_conn_url").(string)
	qm, err := queue.Initialize(queue.IpfsQueue, mqConnectionURL)
	if err != nil {
		c.Error(err)
		return
	}
	// TODO:
	// add in appropriate rabbitmq processing to delete from database
	qm.PublishMessage(hash)
	c.JSON(http.StatusOK, gin.H{"deleted": hash})
}

// RemovePinFromCluster is used to remove a pin from the cluster global state
// this will mean that all nodes in the cluster will no longer track the pin
// TODO: fully implement, add in goroutines
func RemovePinFromCluster(c *gin.Context) {
	hash := c.Param("hash")
	manager := rtfs_cluster.Initialize()
	err := manager.RemovePinFromCluster(hash)
	if err != nil {
		c.Error(err)
		return
	}
	mqConnectionURL := c.MustGet("mq_conn_url").(string)
	qm, err := queue.Initialize(queue.IpfsQueue, mqConnectionURL)
	if err != nil {
		c.Error(err)
		return
	}
	qm.PublishMessage(hash)
	c.JSON(http.StatusOK, gin.H{"deleted": hash})
}

// RunTestGarbageCollection is used to run a test
// of our garbage collector
func RunTestGarbageCollection(c *gin.Context) {
	db := database.OpenDBConnection()
	um := models.NewUploadManager(db)
	deletedUploads := um.RunTestDatabaseGarbageCollection()
	c.JSON(http.StatusOK, gin.H{"deleted": deletedUploads})
}

// GetLocalStatusForClusterPin is used to get teh localnode's cluster status for a particular pin
func GetLocalStatusForClusterPin(c *gin.Context) {
	hash := c.Param("hash")
	// initialize a connection to the cluster
	manager := rtfs_cluster.Initialize()
	// get the cluster status for the cid only asking the local cluster node
	status, err := manager.GetStatusForCidLocally(hash)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusFound, gin.H{"status": status})
}

// GetGlobalStatusForClusterPin is used to get the global cluster status for a particular pin
func GetGlobalStatusForClusterPin(c *gin.Context) {
	hash := c.Param("hash")
	// initialize a connection to the cluster
	manager := rtfs_cluster.Initialize()
	// get teh cluster wide status for this particular pin
	status, err := manager.GetStatusForCidGlobally(hash)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusFound, gin.H{"status": status})
}

// FetchLocalClusterStatus is used to fetch the status of the localhost's
// cluster state, and not the rest of the cluster
// TODO: cleanup
func FetchLocalClusterStatus(c *gin.Context) {
	// this will hold all the retrieved content hashes
	var cids []*gocid.Cid
	// this will hold all the statuses of the content hashes
	var statuses []string
	// initialize a connection to the cluster
	manager := rtfs_cluster.Initialize()
	// fetch a map of all the statuses
	maps, err := manager.FetchLocalStatus()
	if err != nil {
		c.Error(err)
		return
	}
	// parse the maps
	for k, v := range maps {
		cids = append(cids, k)
		statuses = append(statuses, v)
	}
	c.JSON(http.StatusOK, gin.H{"cids": cids, "statuses": statuses})
}

// GetLocalPins is used to get the pins tracked by the local ipfs node
func GetLocalPins(c *gin.Context) {
	// initialize a connection toe the local ipfs node
	manager := rtfs.Initialize("")
	// get all the known local pins
	// WARNING: THIS COULD BE A VERY LARGE LIST
	pinInfo, err := manager.Shell.Pins()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"pins": pinInfo})
}

// GetUploadsFromDatabase is used to read a list of uploads from our database
// TODO: cleanup
func GetUploadsFromDatabase(c *gin.Context) {
	// open a connection to the database
	db := database.OpenDBConnection()
	// create an upload manager interface
	um := models.NewUploadManager(db)
	// fetch the uplaods
	uploads := um.GetUploads()
	if uploads == nil {
		um.DB.Close()
		c.JSON(http.StatusNotFound, nil)
		return
	}
	um.DB.Close()
	c.JSON(http.StatusFound, gin.H{"uploads": uploads})
}

// GetUploadsForAddress is used to read a list of uploads from a particular eth address
// TODO: cleanup
func GetUploadsForAddress(c *gin.Context) {
	// open connection to the database
	db := database.OpenDBConnection()
	// establish a new upload manager
	um := models.NewUploadManager(db)
	// fetch all uploads for that address
	uploads := um.GetUploadsForAddress(c.Param("address"))
	if uploads == nil {
		um.DB.Close()
		c.JSON(http.StatusNotFound, nil)
		return
	}
	um.DB.Close()
	c.JSON(http.StatusFound, gin.H{"uploads": uploads})
}

// GetObjectStatForIpfs is used to get the
// particular object state from the local
// ipfs node
func GetObjectStatForIpfs(c *gin.Context) {
	key := c.Param("key")
	manager := rtfs.Initialize("")
	stats, err := manager.ObjectStat(key)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// CheckLocalNodeForPin is used to check whether or not
// the local node has pinned the content
func CheckLocalNodeForPin(c *gin.Context) {
	hash := c.Param("hash")
	manager := rtfs.Initialize("")
	present, err := manager.ParseLocalPinsForHash(hash)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"present": present})
}
