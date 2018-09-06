package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/minio/minio-go"
	log "github.com/sirupsen/logrus"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/gin-gonic/gin"
)

// CalculateContentHashForFile is used to calculate the content hash
// for a particular file, without actually storing it or processing it
func (api *API) calculateContentHashForFile(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	fileHandler, err := c.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}

	reader, err := fileHandler.Open()
	if err != nil {
		FailOnError(c, err)
		return
	}
	defer reader.Close()
	hash, err := utils.GenerateIpfsMultiHashForFile(reader)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("content hash calculation for file requested")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": hash,
	})
}

// PinHashLocally is used to pin a hash to the local ipfs node
func (api *API) pinHashLocally(c *gin.Context) {
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

	mqConnectionURL := api.TConfig.RabbitMQ.URL

	qm, err := queue.Initialize(queue.IpfsPinQueue, mqConnectionURL, true, false)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	err = qm.PublishMessageWithExchange(ip, queue.PinExchange)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("ipfs pin request sent to backend")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": "pin request sent to backend",
	})
}

// GetFileSizeInBytesForObject is used to retrieve the size of an object in bytes
func (api *API) getFileSizeInBytesForObject(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	key := c.Param("key")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		api.Logger.Error(err)
		FailOnServerError(c, err)
		return
	}
	sizeInBytes, err := manager.GetObjectFileSizeInBytes(key)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("ipfs object file size requested")

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"response": gin.H{
			"object":        key,
			"size_in_bytes": sizeInBytes,
		},
	})

}

// AddFileLocallyAdvanced is used to upload a file in a more resilient
// and efficient manner than our traditional simple upload. Note that
// it does not give the user a content hash back immediately
func (api *API) addFileLocallyAdvanced(c *gin.Context) {
	holdTimeInMonths, exists := c.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}

	accessKey := api.TConfig.MINIO.AccessKey
	secretKey := api.TConfig.MINIO.SecretKey
	endpoint := fmt.Sprintf("%s:%s", api.TConfig.MINIO.Connection.IP, api.TConfig.MINIO.Connection.Port)

	mqURL := api.TConfig.RabbitMQ.URL

	miniManager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, false)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	fileHandler, err := c.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("opening file")
	openFile, err := fileHandler.Open()
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	fmt.Println("file opened")
	username := GetAuthenticatedUserFromContext(c)

	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", username, randString)
	fmt.Println("storing file in minio")
	_, err = miniManager.PutObject(FilesUploadBucket, objectName, openFile, fileHandler.Size, minio.PutObjectOptions{})
	if err != nil {
		api.Logger.Error(err)
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
	qm, err := queue.Initialize(queue.IpfsFileQueue, mqURL, true, false)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	err = qm.PublishMessage(ifp)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("advanced ipfs file upload requested")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": "file upload request sent to backend",
	})
}

// AddFileLocally is used to add a file to our local ipfs node in a simple manner
// this route gives the user back a content hash for their file immedaitely
func (api *API) addFileLocally(c *gin.Context) {
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
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// pin the file
	fmt.Println("adding file")
	resp, err := manager.Add(openFile)
	if err != nil {
		api.Logger.Error(err)
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
	mqConnectionURL := api.TConfig.RabbitMQ.URL
	// initialize a connectino to rabbitmq
	qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqConnectionURL, true, false)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	// Consider whether or not we should trigger a cluster pin here

	// publish the database file add message
	err = qm.PublishMessage(dfa)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	pin := queue.IPFSPin{
		CID:              resp,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeinMonthsInt,
	}

	qm, err = queue.Initialize(queue.IpfsPinQueue, mqConnectionURL, true, false)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	err = qm.PublishMessageWithExchange(pin, queue.PinExchange)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("simple ipfs file upload processed")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": resp,
	})
}

// IpfsPubSubPublish is used to publish a pubsub msg
func (api *API) ipfsPubSubPublish(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	topic := c.Param("topic")
	message, present := c.GetPostForm("message")
	if !present {
		FailNoExistPostForm(c, "message")
		return
	}
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	err = manager.PublishPubSubMessage(topic, message)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("ipfs pub sub message published")

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"response": gin.H{
			"topic":   topic,
			"message": message,
		},
	})
}

// RemovePinFromLocalHost is used to remove a pin from the  ipfs node
func (api *API) removePinFromLocalHost(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if username != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to removal route")
		return
	}
	hash := c.Param("hash")
	mqURL := api.TConfig.RabbitMQ.URL

	qm, err := queue.Initialize(queue.IpfsPinRemovalQueue, mqURL, true, false)
	if err != nil {
		api.Logger.Error(err)
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
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("ipfs pin removal request sent to backend")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": "pin removal sent to backend",
	})
}

// GetLocalPins is used to get the pins tracked by the serving ipfs node
// This is admin locked to avoid peformance penalties from looking up the pinset
func (api *API) getLocalPins(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	// initialize a connection toe the local ipfs node
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// get all the known local pins
	// WARNING: THIS COULD BE A VERY LARGE LIST
	pinInfo, err := manager.Shell.Pins()
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("ipfs pin list requested")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": pinInfo,
	})
}

// GetObjectStatForIpfs is used to get the object stats for the particular cid
func (api *API) getObjectStatForIpfs(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	key := c.Param("key")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	stats, err := manager.ObjectStat(key)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("ipfs object stat requested")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": stats,
	})
}

// CheckLocalNodeForPin is used to check whether or not the serving node is tacking the particular pin
func (api *API) checkLocalNodeForPin(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	hash := c.Param("hash")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	present, err := manager.ParseLocalPinsForHash(hash)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("ipfs pin check requested")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": present,
	})
}

// DownloadContentHash is used to download a particular content hash from the network
func (api *API) downloadContentHash(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
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
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// read the contents of the file
	reader, err := manager.Shell.Cat(contentHash)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// get the size of hte file in bytes
	sizeInBytes, err := manager.GetObjectFileSizeInBytes(contentHash)
	if err != nil {
		api.Logger.Error(err)
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

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("ipfs content download requested")

	// send them the file
	c.DataFromReader(200, int64(sizeInBytes), contentType, reader, extraHeaders)
}
