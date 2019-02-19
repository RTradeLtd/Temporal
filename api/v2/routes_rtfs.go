package v2

import (
	"bytes"
	"errors"
	"html"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/c2h5oh/datasize"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/RTradeLtd/crypto"
	"github.com/RTradeLtd/database/models"
	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
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
	upload, err := api.upm.FindUploadByHashAndUserAndNetwork(username, hash, "public")
	// by this conditional if statement passing, it means the user has
	// upload content matching this hash before, and we don't want to charge them
	// so we should gracefully abort further processing
	if err == nil || upload != nil {
		Respond(
			c,
			http.StatusOK,
			gin.H{
				"response": "it seems like you have uploaded content matching this hash already. To save your credits, no charge was placed and the call was gracefully aborted. Please contact support@rtradetechnologies.com if you believe this is an issue",
			},
		)
		return
	}
	// extract post forms
	forms := api.extractPostForms(c, "hold_time")
	if len(forms) == 0 {
		return
	}
	// parse hold time
	holdTimeInt, err := api.validateHoldTime(username, forms["hold_time"])
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
		usages, err := api.usage.FindByUserName(username)
		if err != nil {
			api.LogError(c, err, eh.CantUploadError)(http.StatusBadRequest)
			return
		}
		api.LogError(c, err,
			api.formatUploadErrorMessage(hash, usages.CurrentDataUsedBytes, usages.MonthlyDataLimitBytes),
		)
		return
	}
	// determine cost of upload
	cost, err := utils.CalculatePinCost(username, hash, holdTimeInt, api.ipfs, api.usage)
	if err != nil {
		api.LogError(c, err, eh.CostCalculationError)(http.StatusBadRequest)
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
	qp := queue.IPFSClusterPin{
		CID:              hash,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		Size:             int64(stats.CumulativeSize),
		CreditCost:       cost,
	}
	// sent pin message
	if err = api.queues.cluster.PublishMessage(qp); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		api.refundUserCredits(username, "pin", cost)
		api.usage.ReduceDataUsage(username, uint64(stats.CumulativeSize))
		return
	}
	// log success and return
	api.l.Infow("ipfs pin request sent to backend", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": "pin request sent to backend"})
}

// AddFile is used to add a file to ipfs with optional encryption
func (api *API) addFile(c *gin.Context) {
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
	holdTimeInMonthsInt, err := api.validateHoldTime(username, forms["hold_time"])
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
	api.l.Debug("opening file")
	openFile, err := fileHandler.Open()
	if err != nil {
		api.LogError(c, err, eh.FileOpenError)(http.StatusBadRequest)
		return
	}
	/* THIS IS TEMPORARILY DISABLED FOR FILE ADDS AS USING ANY EXTRA OPTIONS
	WILL RESULT IN A COMPLETELY DIFFERENT CONTENT HASH
	*/
	// lets make sure the user hasn't actually uploaded this file
	// unfortunately the OnlyHash option is resulting in a different hash than when we add the file
	// since adding a file that's already been added wont actually store any extra data
	// we should
	fileBytes, err := ioutil.ReadAll(openFile)
	if err != nil {
		Fail(c, err)
		return
	}
	hash, err := api.ipfs.Add(bytes.NewReader(fileBytes), ipfsapi.OnlyHash(true))
	if err != nil {
		api.LogError(c, err, eh.IPFSAddError)(http.StatusBadRequest)
		return
	}
	upload, err := api.upm.FindUploadByHashAndUserAndNetwork(username, hash, "public")
	// by this conditional if statement passing, it means the user has
	// upload content matching this hash before, and we don't want to charge them
	// so we should gracefully abort further processing
	if err == nil || upload != nil {
		Respond(
			c,
			http.StatusOK,
			gin.H{
				"response": "it seems like you have uploaded content matching this hash already. To save your credits, no charge was placed and the call was gracefully aborted. Please contact support@rtradetechnologies.com if you believe this is an issue",
			},
		)
		return
	}
	// format size of file into gigabytes
	fileSizeInGB := uint64(fileHandler.Size) / datasize.GB.Bytes()
	api.l.Debug("user", username, "file_size_in_gb", fileSizeInGB)
	// validate if they can upload an object of this size
	if err := api.usage.CanUpload(username, fileSizeInGB); err != nil {
		usages, err := api.usage.FindByUserName(username)
		if err != nil {
			api.LogError(c, err, eh.CantUploadError)(http.StatusBadRequest)
			return
		}
		api.LogError(c, err,
			api.formatUploadErrorMessage(fileHandler.Filename, usages.CurrentDataUsedBytes, usages.MonthlyDataLimitBytes),
		)
		return
	}
	// calculate code of upload
	cost, err := utils.CalculateFileCost(username, holdTimeInMonthsInt, fileHandler.Size, api.usage)
	if err != nil {
		api.LogError(c, err, eh.CostCalculationError)(http.StatusBadRequest)
		return
	}
	// validate they have enough credits to pay for the upload
	if err = api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	// update their data usage
	if err := api.usage.UpdateDataUsage(username, uint64(fileHandler.Size)); err != nil {
		api.LogError(c, err, eh.DataUsageUpdateError)(http.StatusBadRequest)
		api.refundUserCredits(username, "file", cost)
		return
	}
	var reader io.Reader
	// encrypt file is passphrase is given
	if c.PostForm("passphrase") != "" {
		userUsage, err := api.usage.FindByUserName(username)
		if err != nil {
			api.LogError(c, err, eh.UserSearchError)(http.StatusBadRequest)
			return
		}
		// if the user is within the free tier, then we throttle on-demand encryption
		// free accounts are limited to a file upload size of 275MB when performing
		// on-demand encryption. Non free accounts do not have this limit
		if userUsage.Tier == models.Free {
			megabytesUint := datasize.MB.Bytes()
			maxSize := megabytesUint * 275
			if fileHandler.Size > int64(maxSize) {
				Fail(c, errors.New("free accounts are limited to a max file size of 275MB when using on-demand encryption"))
				api.refundUserCredits(username, "file", cost)
				api.usage.ReduceDataUsage(username, uint64(fileHandler.Size))
				return
			}
		}
		// html decode strings
		decodedPassPhrase := html.UnescapeString(c.PostForm("passphrase"))
		encrypted, err := crypto.NewEncryptManager(decodedPassPhrase).Encrypt(bytes.NewReader(fileBytes))
		if err != nil {
			api.LogError(c, err, eh.EncryptionError)(http.StatusBadRequest)
			api.refundUserCredits(username, "file", cost)
			api.usage.ReduceDataUsage(username, uint64(fileHandler.Size))
			return
		}
		reader = bytes.NewReader(encrypted)
		// generate an encryption manager and encrypt
	} else {
		reader = bytes.NewReader(fileBytes)
	}
	api.l.Debug("adding file...")
	resp, err := api.ipfs.Add(reader)
	if err != nil {
		api.LogError(c, err, eh.IPFSAddError)(http.StatusBadRequest)
		api.refundUserCredits(username, "file", cost)
		api.usage.ReduceDataUsage(username, uint64(fileHandler.Size))
		return
	}
	// if this was an encrypted upload we need to update the encrypted upload table
	// ipfs cluster pin handles updating the regular uploads table
	if c.PostForm("passphrase") != "" {
		if _, err := api.ue.NewUpload(username, fileHandler.Filename, "public", resp); err != nil {
			api.LogError(c, err, eh.DatabaseUpdateError)(http.StatusBadRequest)
			// dont refund here as the data is already available on ipfs
			return
		}
	}
	api.l.Debug("file uploaded to ipfs")
	qp := queue.IPFSClusterPin{
		CID:              resp,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeInMonthsInt,
	}
	// send message to rabbitmq
	if err = api.queues.cluster.PublishMessage(qp); err != nil {
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
	// validate they can submit pubsub message calls
	if err := api.usage.CanPublishPubSub(username); err != nil {
		api.LogError(c, err, "sending a pubsub message will go over your monthly limit")(http.StatusBadRequest)
		return
	}
	// publish the actual message
	if err = api.ipfs.PubSubPublish(topic, forms["message"]); err != nil {
		api.LogError(c, err, eh.IPFSPubSubPublishError)(http.StatusBadRequest)
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
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
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
	api.l.Infow("ipfs dag get requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": out})
}

// extendPin is used to extend the lifetime of a pin by a certain number of months
func (api *API) extendPin(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// hash to retrieve dag for
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// extract needed post forms
	forms := api.extractPostForms(c, "hold_time")
	if len(forms) == 0 {
		return
	}
	// validate the hold time
	holdTimeInt, err := api.validateHoldTime(username, forms["hold_time"])
	if err != nil {
		Fail(c, err)
		return
	}
	// find usage model
	usage, err := api.usage.FindByUserName(username)
	if err != nil {
		api.LogError(c, err, eh.UserSearchError)(http.StatusBadRequest)
		return
	}
	// make sure they aren't a free account
	if usage.Tier == models.Free {
		Fail(c, errors.New("free accounts are not allowed to extend pin times"))
		return
	}
	// find upload
	upload, err := api.upm.FindUploadByHashAndUserAndNetwork(username, hash, "public")
	if err != nil {
		api.LogError(c, err, eh.UploadSearchError)(http.StatusBadRequest)
		return
	}
	// ensure even with pin time extension, it wont breach two year limit
	if err := api.ensureTwoYearMax(upload, holdTimeInt); err != nil {
		Fail(c, err)
		return
	}
	// calculate cost of hold time extension
	cost, err := utils.CalculatePinCost(username, hash, holdTimeInt, api.ipfs, api.usage)
	if err != nil {
		api.LogError(c, err, eh.CostCalculationError)(http.StatusBadRequest)
		return
	}
	// validate they have enough credits
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	// extend garbage collection period
	if err := api.upm.ExtendGarbageCollectionPeriod(username, hash, "public", int(holdTimeInt)); err != nil {
		api.LogError(c, err, eh.PinExtendError)(http.StatusBadRequest)
		api.refundUserCredits(username, "pin", cost)
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": "pin time successfully extended"})
}
