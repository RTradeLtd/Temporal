package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/minio/minio-go"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"
	"github.com/gin-gonic/gin"
)

// PinHashLocally is used to pin a hash to the local ipfs node
func PinHashLocally(c *gin.Context) {
	// check if its for a private network
	_, exists := c.GetPostForm("use_private_network")
	if exists {
		PinToHostedIPFSNetwork(c)
		return
	}
	hash := c.Param("hash")
	uploadAddress := GetAuthenticatedUserFromContext(c)
	holdTimeInMonths, exists := c.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}

	ip := queue.IPFSPin{
		CID:              hash,
		NetworkName:      "public",
		EthAddress:       uploadAddress,
		HoldTimeInMonths: holdTimeInt,
	}

	mqConnectionURL, ok := c.MustGet("mq_conn_url").(string)
	if !ok {
		FailOnError(c, errors.New("unable to load rabbitmq"))
		return
	}

	qm, err := queue.Initialize(queue.IpfsPinQueue, mqConnectionURL)
	if err != nil {
		FailOnError(c, err)
		return
	}

	err = qm.PublishMessageWithExchange(ip, queue.PinExchange)
	if err != nil {
		FailOnError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "pin request sent to backend"})
}

// GetFileSizeInBytesForObject is used to retrieve the size of an object in bytes
func GetFileSizeInBytesForObject(c *gin.Context) {
	key := c.Param("key")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	sizeInBytes, err := manager.GetObjectFileSizeInBytes(key)
	if err != nil {
		FailOnError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"object":        key,
		"size_in_bytes": sizeInBytes,
	})

}

func AddFileLocallyAdvanced(c *gin.Context) {
	_, exists := c.GetPostForm("use_private_network")
	if exists {
		//TODO need to create another function to add file with no response
		AddFileToHostedIPFSNetwork(c)
		return
	}

	cC := c.Copy()

	holdTimeInMonths, exists := cC.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}

	credentials, ok := cC.MustGet("minio_credentials").(map[string]string)
	if !ok {
		FailedToLoadMiddleware(c, "minio credentials")
		return
	}
	secure, ok := cC.MustGet("minio_secure").(bool)
	if !ok {
		FailedToLoadMiddleware(c, "minio secure")
		return
	}
	endpoint, ok := cC.MustGet("minio_endpoint").(string)
	if !ok {
		FailedToLoadMiddleware(c, "minio endpoint")
		return
	}
	mqURL, ok := c.MustGet("mq_conn_url").(string)
	if !ok {
		FailedToLoadMiddleware(c, "rabbitmq")
		return
	}

	miniManager, err := mini.NewMinioManager(endpoint, credentials["access_key"], credentials["secret_key"], secure)
	if err != nil {
		FailOnError(c, err)
		return
	}
	fileHandler, err := cC.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("opening file")
	openFile, err := fileHandler.Open()
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("file opened")
	ethAddress := GetAuthenticatedUserFromContext(cC)

	holdTimeInMonthsInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}

	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", ethAddress, randString)
	fmt.Println("storing file in minio")
	_, err = miniManager.PutObject(FilesUploadBucket, objectName, openFile, fileHandler.Size, minio.PutObjectOptions{})
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("file stored in minio")
	ifp := queue.IPFSFile{
		BucketName:       FilesUploadBucket,
		ObjectName:       objectName,
		EthAddress:       ethAddress,
		NetworkName:      "public",
		HoldTimeInMonths: holdTimeInMonths,
	}
	qm, err := queue.Initialize(queue.IpfsFileQueue, mqURL)
	if err != nil {
		FailOnError(c, err)
		return
	}

	err = qm.PublishMessage(ifp)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"eth_address": ethAddress,
		"hold_time":   holdTimeInMonthsInt,
	})
}

// AddFileLocally is used to add a file to our local ipfs node
// this will have to be done first before pushing any file's to the cluster
// this needs to be optimized so that the process doesn't "hang" while uploading
func AddFileLocally(c *gin.Context) {
	_, exists := c.GetPostForm("use_private_network")
	if exists {
		AddFileToHostedIPFSNetwork(c)
		return
	}
	cC := c.Copy()
	fmt.Println("fetching file")
	// fetch the file, and create a handler to interact with it
	fileHandler, err := cC.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}

	uploaderAddress := GetAuthenticatedUserFromContext(cC)

	holdTimeinMonths, present := cC.GetPostForm("hold_time")
	if !present {
		FailNoExistPostForm(c, "post_form")
		return
	}
	holdTimeinMonthsInt, err := strconv.ParseInt(holdTimeinMonths, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("opening file")
	// open the file
	openFile, err := fileHandler.Open()
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("file opened")
	fmt.Println("initializing manager")
	// initialize a connection to the local ipfs node
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	// pin the file
	fmt.Println("adding file")
	resp, err := manager.Shell.Add(openFile)
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("file added")
	// construct a message to rabbitmq to upad the database
	dfa := queue.DatabaseFileAdd{
		Hash:             resp,
		HoldTimeInMonths: holdTimeinMonthsInt,
		UploaderAddress:  uploaderAddress,
		NetworkName:      "public",
	}
	mqConnectionURL := c.MustGet("mq_conn_url").(string)
	// initialize a connectino to rabbitmq
	qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqConnectionURL)
	if err != nil {
		FailOnError(c, err)
		return
	}
	clusterManager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	decodedHash, err := clusterManager.DecodeHashString(resp)
	if err != nil {
		FailOnError(c, err)
		return
	}
	// spawn a cluster pin as a go-routine
	go func() {
		err := clusterManager.Pin(decodedHash)
		if err != nil {
			//TODO: LOG IT
			fmt.Println("error encountered pinning to cluster ", err)
		}
	}()
	// publish the database file add message
	err = qm.PublishMessage(dfa)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": resp})
}

// IpfsPubSubPublish is used to publish a pubsub msg
func IpfsPubSubPublish(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	topic := c.Param("topic")
	message, present := c.GetPostForm("message")
	if !present {
		FailNoExistPostForm(c, "message")
		return
	}
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	err = manager.PublishPubSubMessage(topic, message)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"topic":   topic,
		"message": message,
	})
}

// IpfsPubSubConsume is used to consume pubsub messages
func IpfsPubSubConsume(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	contextCopy := c.Copy()
	topic := contextCopy.Param("topic")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	go func() {
		manager.SubscribeToPubSubTopic(topic)
		manager.ConsumeSubscription(manager.PubSub)
	}()

	c.JSON(http.StatusOK, gin.H{
		"status": "consuming messages in background",
	})
}

// RemovePinFromLocalHost is used to remove a pin from the ipfs instance
// TODO: fully implement
func RemovePinFromLocalHost(c *gin.Context) {
	// fetch hash param
	hash := c.Param("hash")
	ethAddress := GetAuthenticatedUserFromContext(c)
	rm := queue.IPFSPinRemoval{
		ContentHash: hash,
		NetworkName: "public",
		EthAddress:  ethAddress,
	}
	mqURL, ok := c.MustGet("mq_conn_url").(string)
	if !ok {
		FailedToLoadMiddleware(c, "rabbit mq")
		return
	}
	qm, err := queue.Initialize(queue.IpfsPinRemovalQueue, mqURL)
	if err != nil {
		FailOnError(c, err)
		return
	}
	err = qm.PublishMessageWithExchange(rm, queue.PinRemovalExchange)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "pin removal sent to backend",
	})
}

// GetLocalPins is used to get the pins tracked by the local ipfs node
func GetLocalPins(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	// initialize a connection toe the local ipfs node
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	// get all the known local pins
	// WARNING: THIS COULD BE A VERY LARGE LIST
	pinInfo, err := manager.Shell.Pins()
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"pins": pinInfo})
}

// GetObjectStatForIpfs is used to get the
// particular object state from the local
// ipfs node
func GetObjectStatForIpfs(c *gin.Context) {
	key := c.Param("key")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	stats, err := manager.ObjectStat(key)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// CheckLocalNodeForPin is used to check whether or not
// the local node has pinned the content
func CheckLocalNodeForPin(c *gin.Context) {
	hash := c.Param("hash")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	present, err := manager.ParseLocalPinsForHash(hash)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"present": present})
}

// DownloadContentHash is used to download a particular content hash from the network
//
func DownloadContentHash(c *gin.Context) {
	_, exists := c.GetPostForm("use_private_network")
	if exists {
		DownloadContentHashForPrivateNetwork(c)
		return
	}
	ethAddress := GetAuthenticatedUserFromContext(c)
	// Admin locked for now
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	var contentType string
	// fetch the specified content type from the user
	contentType, exists = c.GetPostForm("content_type")
	// if not specified, provide a default
	if !exists {
		contentType = "application/octet-stream"
	}

	// get any extra headers the user might want
	exHeaders := c.PostFormArray("extra_headers")

	// get the content hash that is to be downloaded
	contentHash := c.Param("hash")
	// initialize our connection to IPFS
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	// read the contents of the file
	reader, err := manager.Shell.Cat(contentHash)
	if err != nil {
		FailOnError(c, err)
		return
	}
	// get the size of hte file in bytes
	sizeInBytes, err := manager.GetObjectFileSizeInBytes(contentHash)
	if err != nil {
		FailOnError(c, err)
		return
	}
	// parse extra headers if there are any
	extraHeaders := make(map[string]string)
	var header string
	var value string
	// only process if there is actual data to process
	// this will always be admin locked
	if len(exHeaders) > 0 && ethAddress == AdminAddress {
		// the array must be of equal length, as a header has two parts
		// the name of the header, and its value
		// this expects the user to have properly formatted the headers
		// we will need to restrict the headers that we process so we don't
		// open ourselves up to being attacked
		if len(exHeaders)%2 != 0 {
			FailOnError(c, errors.New("extra_headers post form is not even in length"))
			return
		}
		// parse through the available headers
		for i := 1; i < len(exHeaders)-1; i += 2 {
			// retrieve header name
			header = exHeaders[i-1]
			// retrieve header value
			value = exHeaders[i]
			// store data
			extraHeaders[header] = value
		}
	}
	// send them the file
	c.DataFromReader(200, int64(sizeInBytes), contentType, reader, extraHeaders)
}
