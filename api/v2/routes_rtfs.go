package v2

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/c2h5oh/datasize"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

// PinHashLocally is used to pin a hash to the local ipfs node
func (api *API) pinHashLocally(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// validate hash
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// extract post forms
	forms := api.extractPostForms(c, "hold_time")
	if len(forms) == 0 {
		return
	}
	// parse hold time
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// get object size
	stats, err := api.ipfs.Stat(hash)
	if err != nil {
		api.LogError(c, err, eh.IPFSObjectStatError)(http.StatusBadRequest)
		return
	}
	// check to make sure they can upload an object of this size
	if err := api.usage.CanUpload(username, uint64(stats.CumulativeSize)); err != nil {
		api.LogError(c, err, eh.CantUploadError)(http.StatusBadRequest)
	}
	// determine cost of upload
	cost, err := utils.CalculatePinCost(hash, holdTimeInt, api.ipfs, false)
	if err != nil {
		api.LogError(c, err, eh.PinCostCalculationError)(http.StatusBadRequest)
		return
	}
	// validate, and deduct credits if they can upload
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	// update their data usage
	if err := api.usage.UpdateDataUsage(username, uint64(stats.CumulativeSize)); err != nil {
		api.LogError(c, err, eh.DataUsageUpdateError)(http.StatusBadRequest)
		api.refundUserCredits(username, "pin", cost)
		return
	}
	// construct pin message
	ip := queue.IPFSPin{
		CID:              hash,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       cost,
	}
	// sent pin message
	if err = api.queues.pin.PublishMessageWithExchange(ip, queue.PinExchange); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		api.refundUserCredits(username, "pin", cost)
		//TODO: reduce data used if this fails
		return
	}
	// log success and return
	api.l.Infow("ipfs pin request sent to backend", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": "pin request sent to backend"})
}

// AddFileLocallyAdvanced is used to upload a file in a more resilient
// and efficient manner than our traditional simple upload. Note that
// it does not give the user a content hash back immediately
func (api *API) addFileLocallyAdvanced(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// create formatted log
	logger := api.l.With("user", username)
	logger.Debug("file upload request received from user")
	// extract post forms
	forms := api.extractPostForms(c, "hold_time")
	if len(forms) == 0 {
		return
	}
	// parse hold time
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// extract file handler
	fileHandler, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}
	// check that the size of upload is within limits
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		Fail(c, err)
		return
	}
	// validate that they can upload a file of this size
	if err := api.usage.CanUpload(username, uint64(fileHandler.Size)); err != nil {
		api.LogError(c, err, eh.CantUploadError)(http.StatusBadRequest)
		return
	}
	// calculate the cost of this upload
	cost := utils.CalculateFileCost(holdTimeInt, fileHandler.Size, false)
	// validate they have enough credits to pay for the upload
	if err = api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	logger.Debug("opening file")
	// open file into memory
	openFile, err := fileHandler.Open()
	if err != nil {
		api.LogError(c, err, eh.FileOpenError, "user", username)(http.StatusBadRequest)
		api.refundUserCredits(username, "file", cost)
		return
	}
	logger.Debug("file opened")
	// generate random name for object in temporary storage
	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", username, randString)
	logger.Debugf("storing file in minio as %s", objectName)
	// store object in minio, with optional encryption
	if _, err = api.mini.PutObject(objectName, openFile, fileHandler.Size,
		mini.PutObjectOptions{
			Bucket:            FilesUploadBucket,
			EncryptPassphrase: html.UnescapeString(c.PostForm("passphrase")),
		}); err != nil {
		api.LogError(c, err, eh.MinioPutError,
			"user", username)(http.StatusBadRequest)
		api.refundUserCredits(username, "file", cost)
		return
	}
	// update their data usage
	if err := api.usage.UpdateDataUsage(username, uint64(fileHandler.Size)); err != nil {
		api.LogError(c, err, eh.DataUsageUpdateError)(http.StatusBadRequest)
		api.refundUserCredits(username, "file", cost)
		return
	}
	logger.Debugf("file %s stored in minio", objectName)
	// construct file upload message
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
	// send file upload message
	if err = api.queues.file.PublishMessage(ifp); err != nil {
		api.LogError(c, err, eh.QueuePublishError, "user", username)(http.StatusBadRequest)
		api.refundUserCredits(username, "file", cost)
		return
	}
	// log and return
	logger.With("request", ifp).Info("advanced file upload requested")
	Respond(c, http.StatusOK, gin.H{"response": "file upload request sent to backend"})
}

// AddFileLocally is used to add a file to our local ipfs node in a simple manner
// this route gives the user back a content hash for their file immedaitely
func (api *API) addFileLocally(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract post forms
	forms := api.extractPostForms(c, "hold_time")
	if len(forms) == 0 {
		return
	}
	// parse hold time
	holdTimeinMonthsInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// fetch the file, and create a handler to interact with it
	fileHandler, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}
	// validate the size of upload is within limits
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		Fail(c, err)
		return
	}
	// format size of file into gigabytes
	fileSizeInGB := uint64(fileHandler.Size) / datasize.GB.Bytes()
	api.l.Debug("user", username, "file_size_in_gb", fileSizeInGB)
	// validate if they can upload an object of this size
	if err := api.usage.CanUpload(username, fileSizeInGB); err != nil {
		api.LogError(c, err, eh.CantUploadError)(http.StatusBadRequest)
		return
	}
	// calculate code of upload
	cost := utils.CalculateFileCost(holdTimeinMonthsInt, fileHandler.Size, false)
	// validate they have enough credits to pay for the upload
	if err = api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	// update their data usage
	if err := api.usage.UpdateDataUsage(username, uint64(fileHandler.Size)); err != nil {
		usage, _ := api.usage.FindByUserName(username)
		api.l.Debug("monthly_usage", usage.MonthlyDataLimitGB)
		api.l.Debug("current_usage", usage.CurrentDataUsedGB)
		api.LogError(c, err, eh.DataUsageUpdateError)(http.StatusBadRequest)
		api.refundUserCredits(username, "file", cost)
		return
	}
	api.l.Debug("opening file")
	// open file into memory
	openFile, err := fileHandler.Open()
	if err != nil {
		api.LogError(c, err, eh.FileOpenError)(http.StatusBadRequest)
		api.refundUserCredits(username, "file", cost)
		return
	}
	api.l.Debug("adding file...")
	// add file to ipfs
	resp, err := api.ipfs.Add(openFile)
	if err != nil {
		api.LogError(c, err, eh.IPFSAddError)(http.StatusBadRequest)
		api.refundUserCredits(username, "file", cost)
		return
	}
	api.l.Debug("file added")
	// construct database update message
	dfa := queue.DatabaseFileAdd{
		Hash:             resp,
		HoldTimeInMonths: holdTimeinMonthsInt,
		UserName:         username,
		NetworkName:      "public",
		CreditCost:       0,
	}
	// send message to rabbitmq
	if err = api.queues.database.PublishMessage(dfa); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.Infow("simple ipfs file upload processed", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": resp})
}

// IpfsPubSubPublish is used to publish a pubsub msg
func (api *API) ipfsPubSubPublish(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// topic is the topic which the pubsub message will be addressed to
	topic := c.Param("topic")
	// extract post form
	forms := api.extractPostForms(c, "message")
	if len(forms) == 0 {
		return
	}
	// calculate cost of api call
	cost, err := utils.CalculateAPICallCost("pubsub", false)
	if err != nil {
		api.LogError(c, err, eh.CallCostCalculationError)(http.StatusBadRequest)
		return
	}
	// ensure user has enough credits to pay for
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	// validate they can submit pubsub message calls
	if canUpload, err := api.usage.CanPublishPubSub(username); err != nil {
		api.LogError(c, err, "an error occurred looking up pubsub counts in database")(http.StatusBadRequest)
		api.refundUserCredits(username, "pubsub", cost)
		return
	} else if !canUpload {
		Fail(c, errors.New("sending a pubsub message will go over your monthly limit"))
		return
	}
	// publish the actual message
	if err = api.ipfs.PubSubPublish(topic, forms["message"]); err != nil {
		api.LogError(c, err, eh.IPFSPubSubPublishError)(http.StatusBadRequest)
		api.refundUserCredits(username, "pubsub", cost)
		return
	}
	// update pubsub message usage
	if err := api.usage.IncrementPubSubUsage(username, 1); err != nil {
		api.LogError(c, err, "failed to increment pubsub usage counter")(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.Infow("ipfs pub sub message published", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"topic": topic, "message": forms["message"]}})
}

// GetObjectStatForIpfs is used to get the object stats for the particular cid
func (api *API) getObjectStatForIpfs(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// hash is the object to retrieve stats for
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// retrieve stats for the object
	stats, err := api.ipfs.Stat(hash)
	if err != nil {
		api.LogError(c, err, eh.IPFSObjectStatError)
		Fail(c, err)
		return
	}
	// log and return
	api.l.Infow("ipfs object stat requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": stats})
}

// GetDagObject is used to retrieve an IPLD object from ipfs
func (api *API) getDagObject(c *gin.Context) {
	// hash to retrieve dag for
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	var out interface{}
	if err := api.ipfs.DagGet(hash, &out); err != nil {
		api.LogError(c, err, eh.IPFSDagGetError)(http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": out})
}
