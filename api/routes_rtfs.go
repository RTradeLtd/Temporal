package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/gin-gonic/gin"
)

// PinHashLocally is used to pin a hash to the local ipfs node
func PinHashLocally(c *gin.Context) {
	contextCopy := c.Copy()
	hash := contextCopy.Param("hash")
	uploadAddress := contextCopy.PostForm("eth_address")
	holdTimeInMonths := contextCopy.PostForm("hold_time")
	holdTimeInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	go func() {
		// currently after it is pinned, it is sent to the cluster to be pinned
		manager := rtfs.Initialize("")
		// before exiting, it is pinned to the cluster
		err = manager.Pin(hash)
		if err != nil {
			fmt.Println(err)
		}
	}()
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// publish the message, if there was an error finish processing
	err = qm.PublishMessage(dpa)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	qm.Close()
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uploaderAddress := c.PostForm("eth_ddress")
	holdTimeinMonths := c.PostForm("hold_time")
	holdTimeinMonthsInt, err := strconv.ParseInt(holdTimeinMonths, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("opening file")
	// open the file
	openFile, err := fileHandler.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("initializing manager")
	// initialize a connection to the local ipfs node
	manager := rtfs.Initialize("")
	// pin the file
	fmt.Println("adding file")
	resp, err := manager.Shell.Add(openFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// publish the message
	err = qm.PublishMessage(dfa)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": resp})
}

// IpfsPubSubPublish is used to publish a pubsub msg
func IpfsPubSubPublish(c *gin.Context) {
	topic := c.Param("topic")
	message := c.PostForm("message")
	manager := rtfs.Initialize("")
	err := manager.PublishPubSubMessage(topic, message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"topic": topic, "message": message})
}

// IpfsPubSubTest runs a pubsub test
func IpfsPubSubTest(c *gin.Context) {
	manager := rtfs.Initialize("")
	err := manager.PublishPubSubTest(manager.PubTopic)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO:
	// add in appropriate rabbitmq processing to delete from database
	qm.PublishMessage(hash)
	c.JSON(http.StatusOK, gin.H{"deleted": hash})
}

// GetLocalPins is used to get the pins tracked by the local ipfs node
func GetLocalPins(c *gin.Context) {
	// initialize a connection toe the local ipfs node
	manager := rtfs.Initialize("")
	// get all the known local pins
	// WARNING: THIS COULD BE A VERY LARGE LIST
	pinInfo, err := manager.Shell.Pins()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pins": pinInfo})
}

// GetObjectStatForIpfs is used to get the
// particular object state from the local
// ipfs node
func GetObjectStatForIpfs(c *gin.Context) {
	key := c.Param("key")
	manager := rtfs.Initialize("")
	stats, err := manager.ObjectStat(key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"present": present})
}
