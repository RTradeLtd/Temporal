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
	// get hash to pin
	hash := c.Param("hash")
	// ensure validity of hash
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// extract post forms
	forms := api.extractPostForms(c, "network_name", "hold_time")
	if len(forms) == 0 {
		return
	}
	// ensure user has access to network
	if err = CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// parse hold time
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// create pin message
	ip := queue.IPFSPin{
		CID:              hash,
		NetworkName:      forms["network_name"],
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       0,
		JWT:              GetAuthToken(c),
	}
	// send message for processing
	if err = api.queues.pin.PublishMessage(ip); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// log and return
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
	// extract post forms
	forms := api.extractPostForms(c, "network_name", "hold_time")
	if len(forms) == 0 {
		return
	}
	// validate access to private network
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// retrieve file handler
	fileHandler, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}
	// validate size of file is within limits
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		Fail(c, err)
		return
	}
	// open file into memory
	openFile, err := fileHandler.Open()
	if err != nil {
		api.LogError(c, err, eh.FileOpenError)
		Fail(c, err)
		return
	}
	api.l.Debug("opening file")
	api.l.Debug("file opened")
	// generate object name for object name in temporary storage
	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", username, randString)
	// store object in minio
	if _, err = api.mini.PutObject(objectName, openFile, fileHandler.Size, mini.PutObjectOptions{
		Bucket:            FilesUploadBucket,
		EncryptPassphrase: html.UnescapeString(c.PostForm("passphrase")),
	}); err != nil {
		api.LogError(c, err, eh.MinioPutError)(http.StatusBadRequest)
		return
	}
	api.l.Debugf("%s stored in minio", objectName)
	// construct ipfs file upload message
	ifp := queue.IPFSFile{
		MinioHostIP:      api.cfg.MINIO.Connection.IP,
		FileName:         fileHandler.Filename,
		FileSize:         fileHandler.Size,
		BucketName:       FilesUploadBucket,
		ObjectName:       objectName,
		UserName:         username,
		NetworkName:      forms["network_name"],
		HoldTimeInMonths: forms["hold_time"],
		CreditCost:       0,
		JWT:              GetAuthToken(c),
	}
	// send message for processing
	if err = api.queues.file.PublishMessage(ifp); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// log and return
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
	// extract post forms
	forms := api.extractPostForms(c, "network_name", "hold_time")
	if len(forms) == 0 {
		return
	}
	// verify user has access to private network
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)
		Fail(c, err)
		return
	}
	// parse hold time
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// fetch the file, and create a handler to interact with it
	fileHandler, err := c.FormFile("file")
	if err != nil {
		// user error, do not log
		Fail(c, err)
		return
	}
	// validate file size
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		Fail(c, err)
		return
	}
	// open file into memory
	file, err := fileHandler.Open()
	if err != nil {
		api.LogError(c, err, eh.FileOpenError)(http.StatusBadRequest)
		return
	}
	// format a url to connect to for private network
	apiURL := api.GetIPFSEndpoint(forms["network_name"])
	// connect to private ifps network
	ipfsManager, err := rtfs.NewManager(apiURL, GetAuthToken(c), time.Minute*10, true)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	// add file to ipfs
	resp, err := ipfsManager.Add(file)
	if err != nil {
		api.LogError(c, err, eh.IPFSAddError)(http.StatusBadRequest)
		return
	}
	// create database update message
	dfa := queue.DatabaseFileAdd{
		Hash:             resp,
		HoldTimeInMonths: holdTimeInt,
		UserName:         username,
		NetworkName:      forms["network_name"],
		CreditCost:       0,
	}
	// send message for processing
	if err = api.queues.database.PublishMessage(dfa); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// log and return
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
	// get the topic to publish a message too
	topic := c.Param("topic")
	// extract post forms
	forms := api.extractPostForms(c, "network_name", "message")
	if len(forms) == 0 {
		return
	}
	// validate access to private network
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// format a url to connect too
	apiURL := api.GetIPFSEndpoint(forms["network_name"])
	// connect to private ipfs network
	manager, err := rtfs.NewManager(apiURL, GetAuthToken(c), time.Minute*10, true)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	// publish the actual message
	if err = manager.PubSubPublish(topic, forms["message"]); err != nil {
		api.LogError(c, err, eh.IPFSPubSubPublishError)(http.StatusBadRequest)
		return
	}
	// log and return
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
	// get the hash to retrieve stats for
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	//get the network to connect to
	networkName := c.Param("networkName")
	// validate access to network
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// format a url to connect to
	apiURL := api.GetIPFSEndpoint(networkName)
	// connect to private ipfs network
	manager, err := rtfs.NewManager(apiURL, GetAuthToken(c), time.Minute*10, true)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	// get stats for object
	stats, err := manager.Stat(hash)
	if err != nil {
		api.LogError(c, err, eh.IPFSObjectStatError)(http.StatusBadRequest)
		return
	}
	// log and return
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
	// hash to check pin for
	hash := c.Param("hash")
	// validate the type of hash
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// network to connect to
	networkName := c.Param("networkName")
	// validate access to network
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// format a url to connect to
	apiURL := api.GetIPFSEndpoint(networkName)
	// connect to the actual private network
	manager, err := rtfs.NewManager(apiURL, GetAuthToken(c), time.Minute*10, true)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	// check node for pin
	present, err := manager.CheckPin(hash)
	if err != nil {
		api.LogError(c, err, eh.IPFSPinParseError)(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.Infow("private ipfs pin check requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": present})
}

// GetDagObject is used to retrieve an IPLD object from ipfs
func (api *API) getDagObjectForHostedIPFSNetwork(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get the hash to retrieve dag object for
	hash := c.Param("hash")
	// validate type of hash
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// get network to connect to
	networkName := c.Param("networkName")
	// validate the user has access to the private network
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// format a url to connect to
	apiURL := api.GetIPFSEndpoint(networkName)
	// connect to the private ipfs network
	im, err := rtfs.NewManager(apiURL, GetAuthToken(c), time.Minute*10, true)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	// retrieve the dag object
	var out interface{}
	if err := im.DagGet(hash, &out); err != nil {
		api.LogError(c, err, eh.IPFSDagGetError)(http.StatusBadRequest)
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": out})
}
