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
	"github.com/gin-gonic/gin"
)

// CalculateContentHashForFile is used to calculate the content hash
// for a particular file, without actually storing it or providing it
func calculateContentHashForFile(c *gin.Context) {
	fileHandler, err := c.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}

	reader, err := fileHandler.Open()
	if err != nil {
		FailOnError(c, err)
	}
	defer reader.Close()
	hash, err := utils.GenerateIpfsMultiHashForFile(reader)
	if err != nil {
		FailOnError(c, err)
	}
	c.JSON(http.StatusOK, gin.H{"hash": hash})
}

// PinHashLocally is used to pin a hash to the local ipfs node
func pinHashLocally(c *gin.Context) {
	hash := c.Param("hash")
	username := GetAuthenticatedUserFromContext(c)
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
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
	}

	mqConnectionURL, ok := c.MustGet("mq_conn_url").(string)
	if !ok {
		FailOnError(c, errors.New("unable to load rabbitmq"))
		return
	}

	qm, err := queue.Initialize(queue.IpfsPinQueue, mqConnectionURL, true)
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
func getFileSizeInBytesForObject(c *gin.Context) {
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

// AddFileLocallyAdvanced is used to upload a file in a more resilient
// and efficient manner than our traditional simple upload. Note that
// it does not give the user a content hash back immediately and will be sent
// via email (eventually we will have a notification system for the interface)
func addFileLocallyAdvanced(c *gin.Context) {
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
	username := GetAuthenticatedUserFromContext(cC)

	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", username, randString)
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
		UserName:         username,
		NetworkName:      "public",
		HoldTimeInMonths: holdTimeInMonths,
	}
	qm, err := queue.Initialize(queue.IpfsFileQueue, mqURL, true)
	if err != nil {
		FailOnError(c, err)
		return
	}

	err = qm.PublishMessage(ifp)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "file upload request sent to backend"})
}

// AddFileLocally is used to add a file to our local ipfs node
// this will have to be done first before pushing any file's to the cluster
// this needs to be optimized so that the process doesn't "hang" while uploading
func addFileLocally(c *gin.Context) {
	fmt.Println("fetching file")
	// fetch the file, and create a handler to interact with it
	fileHandler, err := c.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}

	username := GetAuthenticatedUserFromContext(c)

	holdTimeinMonths, present := c.GetPostForm("hold_time")
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
	resp, err := manager.Add(openFile)
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("file added")
	// construct a message to rabbitmq to upad the database
	dfa := queue.DatabaseFileAdd{
		Hash:             resp,
		HoldTimeInMonths: holdTimeinMonthsInt,
		UserName:         username,
		NetworkName:      "public",
	}
	mqConnectionURL := c.MustGet("mq_conn_url").(string)
	// initialize a connectino to rabbitmq
	qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqConnectionURL, true)
	if err != nil {
		FailOnError(c, err)
		return
	}

	// Consider whether or not we should trigger a cluster pin here

	// publish the database file add message
	err = qm.PublishMessage(dfa)
	if err != nil {
		FailOnError(c, err)
		return
	}

	pin := queue.IPFSPin{
		CID:              resp,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeinMonthsInt,
	}

	qm, err = queue.Initialize(queue.IpfsPinQueue, mqConnectionURL, true)
	if err != nil {
		FailOnError(c, err)
		return
	}
	err = qm.PublishMessageWithExchange(pin, queue.PinExchange)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": resp})
}

// IpfsPubSubPublish is used to publish a pubsub msg
func ipfsPubSubPublish(c *gin.Context) {
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

// RemovePinFromLocalHost is used to remove a pin from the ipfs instance
func removePinFromLocalHost(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if username != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to removal route")
		return
	}
	hash := c.Param("hash")
	mqURL, ok := c.MustGet("mq_conn_url").(string)
	if !ok {
		FailedToLoadMiddleware(c, "rabbit mq")
		return
	}
	qm, err := queue.Initialize(queue.IpfsPinRemovalQueue, mqURL, true)
	if err != nil {
		FailOnError(c, err)
		return
	}
	rm := queue.IPFSPinRemoval{
		ContentHash: hash,
		NetworkName: "public",
		UserName:    username,
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
func getLocalPins(c *gin.Context) {
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
func getObjectStatForIpfs(c *gin.Context) {
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
func checkLocalNodeForPin(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
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
func downloadContentHash(c *gin.Context) {
	var contentType string
	// fetch the specified content type from the user
	contentType, exists := c.GetPostForm("content_type")
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
	if len(exHeaders) > 0 {
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
