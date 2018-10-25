package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/RTradeLtd/crypto"
	gocid "github.com/ipfs/go-cid"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/gin-gonic/gin"
)

// CalculateContentHashForFile is used to calculate the content hash
// for a particular file, without actually storing it or processing it
func (api *API) calculateContentHashForFile(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	fileHandler, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}

	reader, err := fileHandler.Open()
	if err != nil {
		api.LogError(err, eh.FileOpenError)(c)
		return
	}
	defer reader.Close()
	hash, err := utils.GenerateIpfsMultiHashForFile(reader)
	if err != nil {
		api.LogError(err, eh.IPFSMultiHashGenerationError)(c)
		return
	}

	api.LogWithUser(username).Info("content hash calculation for file requested")

	Respond(c, http.StatusOK, gin.H{"response": hash})
}

// PinHashLocally is used to pin a hash to the local ipfs node
func (api *API) pinHashLocally(c *gin.Context) {
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	username := GetAuthenticatedUserFromContext(c)
	holdTimeInMonths, exists := c.GetPostForm("hold_time")
	if !exists {
		FailWithBadRequest(c, "hold_time")
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	cost, err := utils.CalculatePinCost(hash, holdTimeInt, api.ipfs.Shell, false)
	if err != nil {
		api.LogError(err, eh.PinCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	ip := queue.IPFSPin{
		CID:              hash,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       cost,
	}

	mqConnectionURL := api.cfg.RabbitMQ.URL

	qm, err := queue.Initialize(queue.IpfsPinQueue, mqConnectionURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c)
		api.refundUserCredits(username, "pin", cost)
		return
	}

	if err = qm.PublishMessageWithExchange(ip, queue.PinExchange); err != nil {
		api.LogError(err, eh.QueuePublishError)(c)
		api.refundUserCredits(username, "pin", cost)
		return
	}

	api.LogWithUser(username).Info("ipfs pin request sent to backend")
	Respond(c, http.StatusOK, gin.H{"response": "pin request sent to backend"})
}

// AddFileLocallyAdvanced is used to upload a file in a more resilient
// and efficient manner than our traditional simple upload. Note that
// it does not give the user a content hash back immediately
func (api *API) addFileLocallyAdvanced(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	logger := api.LogWithUser(username)
	logger.Debug("file upload request received from user")

	holdTimeInMonths, exists := c.GetPostForm("hold_time")
	if !exists {
		FailWithBadRequest(c, "hold_time")
		return
	}

	accessKey := api.cfg.MINIO.AccessKey
	secretKey := api.cfg.MINIO.SecretKey
	endpoint := fmt.Sprintf("%s:%s", api.cfg.MINIO.Connection.IP, api.cfg.MINIO.Connection.Port)

	mqURL := api.cfg.RabbitMQ.URL

	miniManager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, false)
	if err != nil {
		api.LogError(err, eh.MinioConnectionError)(c)
		return
	}
	fileHandler, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		Fail(c, err)
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	cost := utils.CalculateFileCost(holdTimeInt, fileHandler.Size, false)
	if err = api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	logger.Debug("opening file")
	openFile, err := fileHandler.Open()
	if err != nil {
		api.LogError(err, eh.FileOpenError,
			"user", username)(c)
		api.refundUserCredits(username, "file", cost)
		return
	}
	logger.Debug("file opened")

	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", username, randString)

	logger.Debugf("storing file in minio as %s", objectName)
	if _, err = miniManager.PutObject(objectName, openFile, fileHandler.Size,
		mini.PutObjectOptions{
			Bucket:            FilesUploadBucket,
			EncryptPassphrase: c.PostForm("passphrase"),
		}); err != nil {
		api.LogError(err, eh.MinioPutError,
			"user", username)(c)
		api.refundUserCredits(username, "file", cost)
		return
	}
	logger.Debugf("file %s stored in minio", objectName)

	ifp := queue.IPFSFile{
		MinioHostIP:      api.cfg.MINIO.Connection.IP,
		FileSize:         fileHandler.Size,
		FileName:         fileHandler.Filename,
		BucketName:       FilesUploadBucket,
		ObjectName:       objectName,
		UserName:         username,
		NetworkName:      "public",
		HoldTimeInMonths: holdTimeInMonths,
		CreditCost:       cost,
		// if passphrase was provided, this file is encrypted
		Encrypted: c.PostForm("passphrase") != "",
	}
	qm, err := queue.Initialize(queue.IpfsFileQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError,
			"user", username)(c)
		api.refundUserCredits(username, "file", cost)
		return
	}

	if err = qm.PublishMessage(ifp); err != nil {
		api.LogError(err, eh.QueuePublishError,
			"user", username)(c)
		api.refundUserCredits(username, "file", cost)
		return
	}

	logger.
		WithField("request", ifp).
		Info("advanced ipfs file upload requested")

	Respond(c, http.StatusOK, gin.H{"response": "file upload request sent to backend"})
}

// AddFileLocally is used to add a file to our local ipfs node in a simple manner
// this route gives the user back a content hash for their file immedaitely
func (api *API) addFileLocally(c *gin.Context) {
	// fetch the file, and create a handler to interact with it
	fileHandler, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		Fail(c, err)
		return
	}

	holdTimeinMonths, present := c.GetPostForm("hold_time")
	if !present {
		FailWithBadRequest(c, "post_form")
		return
	}
	holdTimeinMonthsInt, err := strconv.ParseInt(holdTimeinMonths, 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	username := GetAuthenticatedUserFromContext(c)
	cost := utils.CalculateFileCost(holdTimeinMonthsInt, fileHandler.Size, false)
	if err = api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	// open the file
	api.LogDebug("opening file")
	openFile, err := fileHandler.Open()
	if err != nil {
		api.LogError(err, eh.FileOpenError)(c)
		api.refundUserCredits(username, "file", cost)
		return
	}
	api.LogDebug("file opened")
	api.LogDebug("initializing manager")
	// initialize a connection to the local ipfs node
	if err != nil {
		api.LogError(err, eh.IPFSConnectionError)(c)
		api.refundUserCredits(username, "file", cost)
		return
	}
	// pin the file
	api.LogDebug("adding file...")
	resp, err := api.ipfs.Add(openFile)
	if err != nil {
		api.LogError(err, eh.IPFSAddError)(c)
		api.refundUserCredits(username, "file", cost)
		return
	}
	api.LogDebug("file added")

	// construct a message to rabbitmq to upad the database
	dfa := queue.DatabaseFileAdd{
		Hash:             resp,
		HoldTimeInMonths: holdTimeinMonthsInt,
		UserName:         username,
		NetworkName:      "public",
		CreditCost:       0,
	}
	mqConnectionURL := api.cfg.RabbitMQ.URL

	// initialize a connectino to rabbitmq
	qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqConnectionURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c)
		return
	}

	// publish the database file add message
	if err = qm.PublishMessage(dfa); err != nil {
		api.LogError(err, eh.QueuePublishError)(c)
		return
	}

	qm, err = queue.Initialize(queue.IpfsPinQueue, mqConnectionURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c)
		return
	}

	if err = qm.PublishMessageWithExchange(queue.IPFSPin{
		CID:              resp,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeinMonthsInt,
		CreditCost:       0,
	}, queue.PinExchange); err != nil {
		api.LogError(err, eh.QueuePublishError)(c)
		return
	}

	api.LogWithUser(username).Info("simple ipfs file upload processed")
	Respond(c, http.StatusOK, gin.H{"response": resp})
}

// IpfsPubSubPublish is used to publish a pubsub msg
func (api *API) ipfsPubSubPublish(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	topic := c.Param("topic")
	message, present := c.GetPostForm("message")
	if !present {
		FailWithMissingField(c, "message")
		return
	}
	cost, err := utils.CalculateAPICallCost("pubsub", false)
	if err != nil {
		api.LogError(err, eh.CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	if err != nil {
		api.LogError(err, eh.IPFSConnectionError)(c)
		api.refundUserCredits(username, "pubsub", cost)
		return
	}
	if err = api.ipfs.PublishPubSubMessage(topic, message); err != nil {
		api.LogError(err, eh.IPFSPubSubPublishError)(c)
		api.refundUserCredits(username, "pubsub", cost)
		return
	}

	api.LogWithUser(username).Info("ipfs pub sub message published")
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"topic": topic, "message": message}})
}

// GetLocalPins is used to get the pins tracked by the serving ipfs node
// This is admin locked to avoid peformance penalties from looking up the pinset
func (api *API) getLocalPins(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, eh.UnAuthorizedAdminAccess)
		return
	}
	// get all the known local pins
	// WARNING: THIS COULD BE A VERY LARGE LIST
	pinInfo, err := api.ipfs.Shell.Pins()
	if err != nil {
		api.LogError(err, eh.IPFSPinParseError)(c)
		return
	}

	api.LogWithUser(username).Info("ipfs pin list requested")
	Respond(c, http.StatusOK, gin.H{"response": pinInfo})
}

// GetObjectStatForIpfs is used to get the object stats for the particular cid
func (api *API) getObjectStatForIpfs(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	key := c.Param("key")
	if _, err := gocid.Decode(key); err != nil {
		Fail(c, err)
		return
	}
	stats, err := api.ipfs.ObjectStat(key)
	if err != nil {
		api.LogError(err, eh.IPFSObjectStatError)
		Fail(c, err)
		return
	}

	api.LogWithUser(username).Info("ipfs object stat requested")
	Respond(c, http.StatusOK, gin.H{"response": stats})
}

// CheckLocalNodeForPin is used to check whether or not the serving node is tacking the particular pin
func (api *API) checkLocalNodeForPin(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, eh.UnAuthorizedAdminAccess)
		return
	}
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	present, err := api.ipfs.ParseLocalPinsForHash(hash)
	if err != nil {
		api.LogError(err, eh.IPFSPinParseError)(c)
		return
	}

	api.LogWithUser(username).Info("ipfs pin check requested")

	Respond(c, http.StatusOK, gin.H{"response": present})
}

// DownloadContentHash is used to download a particular content hash from the network
func (api *API) downloadContentHash(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)

	// fetch the specified content type from the user
	contentType, exists := c.GetPostForm("content_type")
	if !exists {
		contentType = "application/octet-stream"
	}

	// get any extra headers the user might want
	exHeaders := c.PostFormArray("extra_headers")

	// get the content hash that is to be downloaded
	contentHash := c.Param("hash")
	if _, err := gocid.Decode(contentHash); err != nil {
		Fail(c, err)
		return
	}

	var (
		err    error
		reader io.Reader
		size   int
	)

	// read the contents of the file
	if reader, err = api.ipfs.Shell.Cat(contentHash); err != nil {
		api.LogError(err, eh.IPFSCatError)(c)
		return
	}

	// decrypt Temporal-encrypted content if key is provided
	decryptKey := c.PostForm("decrypt_key")
	if decryptKey != "" {
		decrypted, err := crypto.NewEncryptManager(decryptKey).Decrypt(reader)
		if err != nil {
			Fail(c, err)
			return
		}
		size = len(decrypted)
		reader = bytes.NewReader(decrypted)
	} else {
		// get the size of the file in bytes
		if size, err = api.ipfs.GetObjectFileSizeInBytes(contentHash); err != nil {
			api.LogError(err, eh.IPFSObjectStatError)(c)
			return
		}
	}

	// parse extra headers if there are any
	extraHeaders := make(map[string]string)
	// only process if there is actual data to process
	if len(exHeaders) > 0 {
		// the array must be of equal length, as a header has two parts
		// the name of the header, and its value
		if len(exHeaders)%2 != 0 {
			FailWithMessage(c, "extra_headers post form is not even in length")
			return
		}
		// parse through the available headers
		for i := 0; i < len(exHeaders); i += 2 {
			if i+1 < len(exHeaders) {
				header := exHeaders[i]
				value := exHeaders[i+1]
				extraHeaders[header] = value
			}
		}
	}

	api.LogWithUser(username).Info("ipfs content download requested")

	// send them the file
	c.DataFromReader(200, int64(size), contentType, reader, extraHeaders)
}
