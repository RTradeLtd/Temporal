package v2

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/rtfs"
	gocid "github.com/ipfs/go-cid"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
)

// PinToHostedIPFSNetwork is used to pin content to a private ipfs network
func (api *API) pinToHostedIPFSNetwork(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	forms := api.extractPostForms(c, "network_name", "hold_time")
	if len(forms) == 0 {
		return
	}
	if err = CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	url := fmt.Sprintf("%s/network/%s/api", api.cfg.Orchestrator.Host+":"+api.cfg.Orchestrator.Port, forms["network_name"])
	manager, err := rtfs.NewManager(url, time.Minute*10)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	cost, err := utils.CalculatePinCost(hash, holdTimeInt, manager, true)
	if err != nil {
		api.LogError(c, err, eh.CallCostCalculationError)(http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	ip := queue.IPFSPin{
		CID:              hash,
		NetworkName:      forms["network_name"],
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       cost,
	}
	if err = api.queues.pin.PublishMessageWithExchange(ip, queue.PinExchange); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		api.refundUserCredits(username, "private-pin", cost)
		return
	}
	api.l.With("user", username).Info("private network pin request sent to backend")
	Respond(c, http.StatusOK, gin.H{"response": "content pin request sent to backend"})
}

// AddFileToHostedIPFSNetworkAdvanced is used to add a file to a private ipfs network in a more advanced and resilient manner
func (api *API) addFileToHostedIPFSNetworkAdvanced(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "network_name", "hold_time")
	if len(forms) == 0 {
		return
	}
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err, http.StatusBadRequest)
		return
	}
	accessKey := api.cfg.MINIO.AccessKey
	secretKey := api.cfg.MINIO.SecretKey
	endpoint := fmt.Sprintf("%s:%s", api.cfg.MINIO.Connection.IP, api.cfg.MINIO.Connection.Port)
	miniManager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, false)
	if err != nil {
		api.LogError(c, err, eh.MinioConnectionError)
		Fail(c, err)
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
	cost := utils.CalculateFileCost(holdTimeInt, fileHandler.Size, true)
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	api.l.Debug("opening file")
	openFile, err := fileHandler.Open()
	if err != nil {
		api.LogError(c, err, eh.FileOpenError)
		api.refundUserCredits(username, "private-file", cost)
		Fail(c, err)
		return
	}
	api.l.Debug("file opened")
	// generate object name
	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", username, randString)
	fmt.Println("storing file in minio")
	if _, err = miniManager.PutObject(objectName, openFile, fileHandler.Size, mini.PutObjectOptions{
		Bucket:            FilesUploadBucket,
		EncryptPassphrase: html.UnescapeString(c.PostForm("passphrase")),
	}); err != nil {
		api.LogError(c, err, eh.MinioPutError)
		api.refundUserCredits(username, "private-file", cost)
		Fail(c, err)
		return
	}
	fmt.Println("file stored in minio")
	ifp := queue.IPFSFile{
		MinioHostIP:      api.cfg.MINIO.Connection.IP,
		FileName:         fileHandler.Filename,
		FileSize:         fileHandler.Size,
		BucketName:       FilesUploadBucket,
		ObjectName:       objectName,
		UserName:         username,
		NetworkName:      forms["network_name"],
		HoldTimeInMonths: forms["hold_time"],
		CreditCost:       cost,
	}
	api.l.Debugf("%s stored in minio", objectName)
	if err = api.queues.file.PublishMessage(ifp); err != nil {
		api.LogError(c, err, eh.QueuePublishError)
		api.refundUserCredits(username, "private-file", cost)
		Fail(c, err)
		return
	}
	api.l.Infow("advanced private ipfs file upload requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": "file upload request sent to backend"})
}

// AddFileToHostedIPFSNetwork is used to add a file to a private IPFS network via the simple method
func (api *API) addFileToHostedIPFSNetwork(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "network_name", "hold_time")
	if len(forms) == 0 {
		return
	}
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)
		Fail(c, err)
		return
	}
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	apiURL := fmt.Sprintf("%s/network/%s/api", api.cfg.Orchestrator.Host+":"+api.cfg.Orchestrator.Port, forms["network_name"])
	ipfsManager, err := rtfs.NewManager(apiURL, time.Minute*10)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	fmt.Println("fetching file")
	// fetch the file, and create a handler to interact with it
	fileHandler, err := c.FormFile("file")
	if err != nil {
		// user error, do not log
		Fail(c, err)
		return
	}
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		Fail(c, err)
		return
	}
	cost := utils.CalculateFileCost(holdTimeInt, fileHandler.Size, true)
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	file, err := fileHandler.Open()
	if err != nil {
		api.LogError(c, err, eh.FileOpenError)(http.StatusBadRequest)
		api.refundUserCredits(username, "private-file", cost)
		return
	}
	resp, err := ipfsManager.Add(file)
	if err != nil {
		api.LogError(c, err, eh.IPFSAddError)(http.StatusBadRequest)
		api.refundUserCredits(username, "private-file", cost)
		return
	}
	pin := queue.IPFSPin{
		CID:              resp,
		NetworkName:      forms["network_name"],
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       0,
	}
	if err = api.queues.pin.PublishMessageWithExchange(pin, queue.PinExchange); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	fmt.Println("file uploaded")
	dfa := queue.DatabaseFileAdd{
		Hash:             resp,
		HoldTimeInMonths: holdTimeInt,
		UserName:         username,
		NetworkName:      forms["network_name"],
		CreditCost:       0,
	}
	if err = api.queues.database.PublishMessage(dfa); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	api.l.Infow("simple private ipfs file upload processed", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": resp})
}

// IpfsPubSubPublishToHostedIPFSNetwork is used to publish a pubsub message to a private ipfs network
func (api *API) ipfsPubSubPublishToHostedIPFSNetwork(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	topic := c.Param("topic")
	forms := api.extractPostForms(c, "network_name", "message")
	if len(forms) == 0 {
		return
	}
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	cost, err := utils.CalculateAPICallCost("pubsub", true)
	if err != nil {
		api.LogError(c, err, eh.CallCostCalculationError)(http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	apiURL := fmt.Sprintf("%s/network/%s/api", api.cfg.Orchestrator.Host+":"+api.cfg.Orchestrator.Port, forms["network_name"])
	manager, err := rtfs.NewManager(apiURL, time.Minute*10)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	if err = manager.PubSubPublish(topic, forms["message"]); err != nil {
		api.LogError(c, err, eh.IPFSPubSubPublishError)(http.StatusBadRequest)
		return
	}

	api.l.Infow("private ipfs pub sub message published", "user", username)

	Respond(c, http.StatusOK, gin.H{"response": gin.H{"topic": topic, "message": forms["message"]}})
}

// GetObjectStatForIpfsForHostedIPFSNetwork is  used to get object stats from a private ipfs network
func (api *API) getObjectStatForIpfsForHostedIPFSNetwork(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	key := c.Param("hash")
	if _, err := gocid.Decode(key); err != nil {
		Fail(c, err)
		return
	}
	networkName := c.Param("networkName")
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}

	apiURL := fmt.Sprintf("%s/network/%s/api", api.cfg.Orchestrator.Host+":"+api.cfg.Orchestrator.Port, networkName)
	manager, err := rtfs.NewManager(apiURL, time.Minute*10)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	stats, err := manager.Stat(key)
	if err != nil {
		api.LogError(c, err, eh.IPFSObjectStatError)(http.StatusBadRequest)
		return
	}
	api.l.Infow("private ipfs object stat requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": stats})
}

// CheckLocalNodeForPinForHostedIPFSNetwork is used to check the serving node for a pin
func (api *API) checkLocalNodeForPinForHostedIPFSNetwork(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	networkName := c.Param("networkName")
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	apiURL := fmt.Sprintf("%s/network/%s/api", api.cfg.Orchestrator.Host+":"+api.cfg.Orchestrator.Port, networkName)
	manager, err := rtfs.NewManager(apiURL, time.Minute*10)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	present, err := manager.CheckPin(hash)
	if err != nil {
		api.LogError(c, err, eh.IPFSPinParseError)(http.StatusBadRequest)
		return
	}
	api.l.Infow("private ipfs pin check requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": present})
}

// GetDagObject is used to retrieve an IPLD object from ipfs
func (api *API) getDagObjectForHostedIPFSNetwork(c *gin.Context) {
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	networkName := c.Param("networkName")
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	apiURL := fmt.Sprintf("%s/network/%s/api", api.cfg.Orchestrator.Host+":"+api.cfg.Orchestrator.Port, networkName)
	im, err := rtfs.NewManager(apiURL, time.Minute*10)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	var out interface{}
	if err := im.DagGet(hash, &out); err != nil {
		api.LogError(c, err, eh.IPFSDagGetError)(http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": out})
}
