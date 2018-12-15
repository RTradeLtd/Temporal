package v2

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/Temporal/utils"
	gocid "github.com/ipfs/go-cid"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/gin-gonic/gin"
)

// PinHashLocally is used to pin a hash to the local ipfs node
func (api *API) pinHashLocally(c *gin.Context) {
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "hold_time")
	if len(forms) == 0 {
		return
	}
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	cost, err := utils.CalculatePinCost(hash, holdTimeInt, api.ipfs, false)
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
	if err = api.queues.pin.PublishMessageWithExchange(ip, queue.PinExchange); err != nil {
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
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	logger := api.l.With("user", username)
	logger.Debug("file upload request received from user")
	forms := api.extractPostForms(c, "hold_time")
	if len(forms) == 0 {
		return
	}
	accessKey := api.cfg.MINIO.AccessKey
	secretKey := api.cfg.MINIO.SecretKey
	endpoint := fmt.Sprintf("%s:%s", api.cfg.MINIO.Connection.IP, api.cfg.MINIO.Connection.Port)
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
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
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
		HoldTimeInMonths: forms["hold_time1"],
		CreditCost:       cost,
		// if passphrase was provided, this file is encrypted
		Encrypted: c.PostForm("passphrase") != "",
	}
	if err = api.queues.file.PublishMessage(ifp); err != nil {
		api.LogError(err, eh.QueuePublishError,
			"user", username)(c)
		api.refundUserCredits(username, "file", cost)
		return
	}
	logger.With("request", ifp).Info("advanced file upload requested")
	Respond(c, http.StatusOK, gin.H{"response": "file upload request sent to backend"})
}

// AddFileLocally is used to add a file to our local ipfs node in a simple manner
// this route gives the user back a content hash for their file immedaitely
func (api *API) addFileLocally(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "hold_time")
	if len(forms) == 0 {
		return
	}
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
	holdTimeinMonthsInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
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
	if err = api.queues.pin.PublishMessageWithExchange(queue.IPFSPin{
		CID:              resp,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeinMonthsInt,
		CreditCost:       0,
	}, queue.PinExchange); err != nil {
		api.LogError(err, eh.QueuePublishError)(c)
		return
	}
	// construct a message to rabbitmq to upad the database
	dfa := queue.DatabaseFileAdd{
		Hash:             resp,
		HoldTimeInMonths: holdTimeinMonthsInt,
		UserName:         username,
		NetworkName:      "public",
		CreditCost:       0,
	}
	if err = api.queues.database.PublishMessage(dfa); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	api.LogWithUser(username).Info("simple ipfs file upload processed")
	Respond(c, http.StatusOK, gin.H{"response": resp})
}

// IpfsPubSubPublish is used to publish a pubsub msg
func (api *API) ipfsPubSubPublish(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	topic := c.Param("topic")
	forms := api.extractPostForms(c, "message")
	if len(forms) == 0 {
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
	if err = api.ipfs.PubSubPublish(topic, forms["message"]); err != nil {
		api.LogError(err, eh.IPFSPubSubPublishError)(c)
		api.refundUserCredits(username, "pubsub", cost)
		return
	}

	api.LogWithUser(username).Info("ipfs pub sub message published")
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"topic": topic, "message": forms["message"]}})
}

// GetObjectStatForIpfs is used to get the object stats for the particular cid
func (api *API) getObjectStatForIpfs(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	key := c.Param("key")
	if _, err := gocid.Decode(key); err != nil {
		Fail(c, err)
		return
	}
	stats, err := api.ipfs.Stat(key)
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
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, eh.UnAuthorizedAdminAccess)
		return
	}
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	present, err := api.ipfs.CheckPin(hash)
	if err != nil {
		api.LogError(err, eh.IPFSPinParseError)(c)
		return
	}

	api.LogWithUser(username).Info("ipfs pin check requested")

	Respond(c, http.StatusOK, gin.H{"response": present})
}

// GetDagObject is used to retrieve an IPLD object from ipfs
func (api *API) getDagObject(c *gin.Context) {
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	var out interface{}
	if err := api.ipfs.DagGet(hash, &out); err != nil {
		api.LogError(err, eh.IPFSDagGetError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": out})
}
