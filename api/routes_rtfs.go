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
	uploadAddress, exists := contextCopy.GetPostForm("eth_address")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "eth_address post form does not exist",
		})
		return
	}
	holdTimeInMonths, exists := contextCopy.GetPostForm("hold_time")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "hold_time post form does not exist",
		})
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	go func() {
		// currently after it is pinned, it is sent to the cluster to be pinned
		manager, err := rtfs.Initialize("")
		if err != nil {
			fmt.Println(err)
			return
		}
		// before exiting, it is pinned to the cluster
		err = manager.Pin(hash)
		if err != nil {
			fmt.Println(err)
			return
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

// GetFileSizeInBytesForObject is used to retrieve the size of an object in bytes
func GetFileSizeInBytesForObject(c *gin.Context) {
	key := c.Param("key")
	manager, err := rtfs.Initialize("")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	sizeInBytes, err := manager.GetObjectFileSizeInBytes(key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"object":        key,
		"size_in_bytes": sizeInBytes,
	})

}

// AddFileLocally is used to add a file to our local ipfs node
// this will have to be done first before pushing any file's to the cluster
// this needs to be optimized so that the process doesn't "hang" while uploading
func AddFileLocally(c *gin.Context) {
	//var dir string
	// look into saving uploaded file
	//c.SaveUploadedFile(fileHandler, dir)
	fmt.Println("fetching file")
	// fetch the file, and create a handler to interact with it
	fileHandler, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uploaderAddress, present := c.GetPostForm("eth_address")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "eth_address post form param not present",
		})
		return
	}
	holdTimeinMonths, present := c.GetPostForm("hold_time")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "hold_time post form param not present",
		})
		return
	}
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
	manager, err := rtfs.Initialize("")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
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
	icp := queue.IpfsClusterPin{
		CID: resp,
	}
	mqConnectionURL := c.MustGet("mq_conn_url").(string)
	// initialize a connectino to rabbitmq
	qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqConnectionURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error connectingto rabbitmq": err.Error()})
		return
	}
	// publish the message
	err = qm.PublishMessage(dfa)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error publishing database file add to rabbit mq": err.Error()})
		return
	}
	qm, err = queue.Initialize(queue.IpfsClusterQueue, mqConnectionURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error connectingto rabbitmq": err.Error()})
		return
	}
	err = qm.PublishMessage(icp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error publishing ipfs cluster pin to rabbit mq": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": resp})
}

// IpfsPubSubPublish is used to publish a pubsub msg
func IpfsPubSubPublish(c *gin.Context) {
	topic := c.Param("topic")
	message, present := c.GetPostForm("message")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "message post form param is not present",
		})
		return
	}
	manager, err := rtfs.Initialize("")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	err = manager.PublishPubSubMessage(topic, message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"topic": topic, "message": message})
}

// IpfsPubSubConsume is used to consume pubsub messages
func IpfsPubSubConsume(c *gin.Context) {
	contextCopy := c.Copy()
	topic := contextCopy.Param("topic")

	go func() {
		manager, err := rtfs.Initialize("")
		if err != nil {
			fmt.Println(err)
			return
		}
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

	manager, err := rtfs.Initialize("")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// remove the file from the local ipfs state
	// TODO: implement some kind of error handling and notification
	err = manager.Shell.Unpin(hash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("error unpinning hash %s", err.Error()),
		})
		return
	}

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
	manager, err := rtfs.Initialize("")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
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
	manager, err := rtfs.Initialize("")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
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
	manager, err := rtfs.Initialize("")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	present, err := manager.ParseLocalPinsForHash(hash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"present": present})
}
