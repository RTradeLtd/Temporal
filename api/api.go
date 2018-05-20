package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/RTradeLtd/Temporal/api/rtfs_cluster"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	gocid "github.com/ipfs/go-cid"

	"github.com/RTradeLtd/Temporal/api/rtfs"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/gin-contrib/rollbar"
	"github.com/gin-gonic/gin"
	"github.com/stvp/roll"
	"github.com/zsais/go-gin-prometheus"
)

// Setup is used to initialize our api.
// it invokes all  non exported function to setup the api.
func Setup() *gin.Engine {
	// we use rollbar for logging errors
	token := os.Getenv("ROLLBAR_TOKEN")
	if token == "" {
		log.Fatal("invalid token")
	}
	roll.Token = token
	roll.Environment = "development"
	r := gin.Default()
	// create gin middleware instance for prom
	p := ginprometheus.NewPrometheus("gin")
	// set the address for prometheus to collect metrics
	p.SetListenAddress("127.0.0.1:6768")
	// load in prom to gin
	p.Use(r)
	r.Use(rollbar.Recovery(false))
	setupRoutes(r)
	return r
}

// setupRoutes is used to setup all of our api routes
func setupRoutes(g *gin.Engine) {

	g.POST("/api/v1/ipfs/pin/:hash", pinHashLocally)
	g.POST("/api/v1/ipfs/add-file", addFileLocally)
	g.POST("/api/v1/ipfs-cluster/pin/:hash", pinHashToCluster)
	g.POST("/api/v1/ipfs-cluster/sync-errors-local", syncClusterErrorsLocally)
	g.POST("/api/v1/ipfs/pubsub/publish/:topic", ipfsPubSubPublish)
	g.POST("/api/v1/ipfs/pubsub/publish-test/:topic", ipfsPubSubTest)
	g.GET("/api/v1/ipfs/pubsub/consume/:topic", ipfsPubSubConsume)
	g.DELETE("/api/v1/ipfs/remove-pin/:hash", removePinFromLocalHost)
	g.DELETE("/api/v1/ipfs-cluster/remove-pin/:hash", removePinFromCluster)
	g.DELETE("/api/v1/database/garbage-collect/test", runTestGarbageCollection)
	g.GET("/api/v1/ipfs-cluster/status-local-pin/:hash", getLocalStatusForClusterPin)
	g.GET("/api/v1/ipfs-cluster/status-global-pin/:hash", getGlobalStatusForClusterPin)
	g.GET("/api/v1/ipfs-cluster/status-local", fetchLocalClusterStatus)
	g.GET("/api/v1/ipfs/pins", getLocalPins)
	g.GET("/api/v1/database/uploads", getUploadsFromDatabase)
	g.GET("/api/v1/database/uploads/:address", getUploadsForAddress)
	g.GET("/api/v1/ipfs/object-stat/:key", getObjectStatForIpfs)
	g.GET("/api/v1/ipfs/check-for-pin/:hash", checkLocalNodeForPin)
}

// pinHashLocally is used to pin a hash to the local ipfs node
// TODO: add in the cluster pin event, optimize
func pinHashLocally(c *gin.Context) {
	hash := c.Param("hash")
	uploadAddress := c.PostForm("uploadAddress")
	holdTimeInMonths := c.PostForm("holdTime")
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
	// initialize the queue
	qm, err := queue.Initialize(queue.DatabasePinAddQueue)
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
	/*go func() {*/
	// currently after it is pinned, it is sent to the cluster to be pinned
	manager := rtfs.Initialize("")
	// before exiting, it is pinned to the cluster
	err = manager.Pin(hash)
	if err != nil {
		fmt.Println(err)
	}
	/*}()*/
	c.JSON(http.StatusOK, gin.H{"upload": dpa})
}

// addFileLocally is used to add a file to our local ipfs node
// this will have to be done first before pushing any file's to the cluster
// this needs to be optimized so that the process doesn't "hang" while uploading
func addFileLocally(c *gin.Context) {
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
	// open the file
	openFile, err := fileHandler.Open()
	if err != nil {
		c.Error(err)
		return
	}
	//respCalc, err := gocid.Parse(fileByteData)
	if err != nil {
		c.Error(err)
		return
	}
	// initialize a connection to the local ipfs node
	manager := rtfs.Initialize("")
	// pin the file
	resp, err := manager.Shell.Add(openFile)
	if err != nil {
		c.Error(err)
		return
	}

	// construct a message to rabbitmq to upad the database
	dfa := queue.DatabaseFileAdd{
		Hash:             resp,
		HoldTimeInMonths: holdTimeinMonthsInt,
		UploaderAddress:  uploaderAddress,
	}
	// initialize a connectino to rabbitmq
	qm, err := queue.Initialize(queue.DatabaseFileAddQueue)
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

// pinHashToCluster is used to pin a hash to the global cluster state
func pinHashToCluster(c *gin.Context) {
	hash := c.Param("hash")
	manager := rtfs_cluster.Initialize()
	// decode the hash
	contentIdentifier := manager.DecodeHashString(hash)
	// pin the hash to the cluster, using -1 for replication factor, indicating to pin EVERYWHERE in the cluster
	manager.Client.Pin(contentIdentifier, -1, -1, hash)
	c.JSON(http.StatusOK, gin.H{"hash": hash})
}

// syncCluserErrorsLocally is used to parse through the local cluster state
// and sync any errors that are detected.
func syncClusterErrorsLocally(c *gin.Context) {
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

func ipfsPubSubPublish(c *gin.Context) {
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

func ipfsPubSubTest(c *gin.Context) {
	manager := rtfs.Initialize("")
	err := manager.PublishPubSubTest(manager.PubTopic)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, nil)
}

func ipfsPubSubConsume(c *gin.Context) {
	topic := c.Param("topic")
	manager := rtfs.Initialize("")
	manager.SubscribeToPubSubTopic(topic)
	manager.ConsumeSubscription(manager.PubSub)
}

// removePinFromLocalHost is used to remove a pin from the ipfs instance
// TODO: fully implement
func removePinFromLocalHost(c *gin.Context) {
	// fetch hash param
	hash := c.Param("hash")
	// initialise a connetion to the local ipfs node
	manager := rtfs.Initialize("")
	// remove the file from the local ipfs state
	err := manager.Shell.Unpin(hash)
	if err != nil {
		c.Error(err)
		return
	}
	// TODO:
	// change to send a message to the cluster to depin
	qm, err := queue.Initialize(queue.IpfsQueue)
	if err != nil {
		c.Error(err)
		return
	}
	qm.PublishMessage(hash)
	c.JSON(http.StatusOK, gin.H{"deleted": hash})
}

// removePinFromCluster is used to remove a pin from the cluster global state
// this will mean that all nodes in the cluster will no longer track the pin
// TODO: fully implement
func removePinFromCluster(c *gin.Context) {
	hash := c.Param("hash")
	manager := rtfs_cluster.Initialize()
	err := manager.RemovePinFromCluster(hash)
	if err != nil {
		c.Error(err)
		return
	}
	qm, err := queue.Initialize(queue.IpfsQueue)
	if err != nil {
		c.Error(err)
		return
	}
	qm.PublishMessage(hash)
	c.JSON(http.StatusOK, gin.H{"deleted": hash})
}

func runTestGarbageCollection(c *gin.Context) {
	db := database.OpenDBConnection()
	um := models.NewUploadManager(db)
	deletedUploads := um.RunTestDatabaseGarbageCollection()
	c.JSON(http.StatusOK, gin.H{"deleted": deletedUploads})
}

// getLocalStatusForClusterPin is used to get teh localnode's cluster status for a particular pin
func getLocalStatusForClusterPin(c *gin.Context) {
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

// getGlobalStatusForClusterPin is used to get the global cluster status for a particular pin
func getGlobalStatusForClusterPin(c *gin.Context) {
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

// fetchLocalClusterStatus is used to fetch the status of the localhost's
// cluster state, and not the rest of the cluster
// TODO: cleanup
func fetchLocalClusterStatus(c *gin.Context) {
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

// getLocalPins is used to get the pins tracked by the local ipfs node
func getLocalPins(c *gin.Context) {
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

// getUploadsFromDatabase is used to read a list of uploads from our database
// TODO: cleanup
func getUploadsFromDatabase(c *gin.Context) {
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

// getUploadsForAddress is used to read a list of uploads from a particular eth address
// TODO: cleanup
func getUploadsForAddress(c *gin.Context) {
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

func getObjectStatForIpfs(c *gin.Context) {
	key := c.Param("key")
	manager := rtfs.Initialize("")
	stats, err := manager.ObjectStat(key)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func checkLocalNodeForPin(c *gin.Context) {
	hash := c.Param("hash")
	manager := rtfs.Initialize("")
	present, err := manager.ParseLocalPinsForHash(hash)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"present": present})
}
