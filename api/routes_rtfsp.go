package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs"
	ipfs_orchestrator "github.com/RTradeLtd/grpc/ipfs-orchestrator"
	gocid "github.com/ipfs/go-cid"
	log "github.com/sirupsen/logrus"

	"github.com/RTradeLtd/database/models"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
)

// PinToHostedIPFSNetwork is used to pin content to a private ipfs network
func (api *API) pinToHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	forms := api.extractPostForms(c, "network_name", "hold_time")
	if len(forms) == 0 {
		return
	}
	err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB)
	if err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c)
		return
	}
	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	url, err := im.GetAPIURLByName(forms["network_name"])
	if err != nil {
		api.LogError(err, eh.APIURLCheckError)(c, http.StatusBadRequest)
		return
	}
	manager, err := rtfs.Initialize("", url)
	if err != nil {
		api.LogError(err, eh.IPFSConnectionError)(c, http.StatusBadRequest)
		return
	}

	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	cost, err := utils.CalculatePinCost(hash, holdTimeInt, manager.Shell, true)
	if err != nil {
		api.LogError(err, eh.CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	ip := queue.IPFSPin{
		CID:              hash,
		NetworkName:      forms["network_name"],
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       cost,
	}

	mqConnectionURL := api.cfg.RabbitMQ.URL

	qm, err := queue.Initialize(queue.IpfsPinQueue, mqConnectionURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c)
		api.refundUserCredits(username, "private-pin", cost)
		return
	}

	if err = qm.PublishMessageWithExchange(ip, queue.PinExchange); err != nil {
		api.LogError(err, eh.QueuePublishError)(c)
		api.refundUserCredits(username, "private-pin", cost)
		return
	}

	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("ipfs pin request for private network sent to backend")

	Respond(c, http.StatusOK, gin.H{"response": "content pin request sent to backend"})
}

// AddFileToHostedIPFSNetworkAdvanced is used to add a file to a private ipfs network in a more advanced and resilient manner
func (api *API) addFileToHostedIPFSNetworkAdvanced(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	forms := api.extractPostForms(c, "network_name", "hold_time")
	if len(forms) == 0 {
		return
	}
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c, http.StatusBadRequest)
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
		api.LogError(err, eh.MinioConnectionError)
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
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	api.LogDebug("opening file")
	openFile, err := fileHandler.Open()
	if err != nil {
		api.LogError(err, eh.FileOpenError)
		api.refundUserCredits(username, "private-file", cost)
		Fail(c, err)
		return
	}
	api.LogDebug("file opened")

	// generate object name
	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", username, randString)
	fmt.Println("storing file in minio")
	if _, err = miniManager.PutObject(objectName, openFile, fileHandler.Size, mini.PutObjectOptions{
		Bucket:            FilesUploadBucket,
		EncryptPassphrase: c.PostForm("passphrase"),
	}); err != nil {
		api.LogError(err, eh.MinioPutError)
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

	// initialize queue
	mqURL := api.cfg.RabbitMQ.URL
	qm, err := queue.Initialize(queue.IpfsFileQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)
		api.refundUserCredits(username, "private-file", cost)
		Fail(c, err)
		return
	}

	// we don't use an exchange for file publishes so that rabbitmq distributes round robin
	if err = qm.PublishMessage(ifp); err != nil {
		api.LogError(err, eh.QueuePublishError)
		api.refundUserCredits(username, "private-file", cost)
		Fail(c, err)
		return
	}

	api.LogWithUser(username).Info("advanced private ipfs file upload requested")
	Respond(c, http.StatusOK, gin.H{"response": "file upload request sent to backend"})
}

// AddFileToHostedIPFSNetwork is used to add a file to a private IPFS network via the simple method
func (api *API) addFileToHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	forms := api.extractPostForms(c, "network_name", "hold_time")
	if len(forms) == 0 {
		return
	}
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)
		Fail(c, err)
		return
	}
	mqURL := api.cfg.RabbitMQ.URL
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}

	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	apiURL, err := im.GetAPIURLByName(forms["network_name"])
	if err != nil {
		api.LogError(err, eh.APIURLCheckError)(c)
		return
	}

	ipfsManager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.LogError(err, eh.IPFSConnectionError)(c)
		return
	}
	qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c)
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
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	file, err := fileHandler.Open()
	if err != nil {
		api.LogError(err, eh.FileOpenError)(c)
		api.refundUserCredits(username, "private-file", cost)
		return
	}
	resp, err := ipfsManager.Add(file)
	if err != nil {
		api.LogError(err, eh.IPFSAddError)(c)
		api.refundUserCredits(username, "private-file", cost)
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
	if err = qm.PublishMessage(dfa); err != nil {
		api.LogError(err, eh.QueuePublishError)(c)
		return
	}

	pin := queue.IPFSPin{
		CID:              resp,
		NetworkName:      forms["network_name"],
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       0,
	}

	qm, err = queue.Initialize(queue.IpfsPinQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c)
		return
	}
	if err = qm.PublishMessageWithExchange(pin, queue.PinExchange); err != nil {
		api.LogError(err, eh.QueuePublishError)(c)
		return
	}

	api.LogWithUser(username).Info("simple private ipfs file upload processed")

	Respond(c, http.StatusOK, gin.H{"response": resp})
}

// IpfsPubSubPublishToHostedIPFSNetwork is used to publish a pubsub message to a private ipfs network
func (api *API) ipfsPubSubPublishToHostedIPFSNetwork(c *gin.Context) {
	topic := c.Param("topic")
	username := GetAuthenticatedUserFromContext(c)
	forms := api.extractPostForms(c, "network_name", "message")
	if len(forms) == 0 {
		return
	}
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c)
		return
	}
	cost, err := utils.CalculateAPICallCost("pubsub", true)
	if err != nil {
		api.LogError(err, eh.CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	apiURL, err := im.GetAPIURLByName(forms["network_name"])
	if err != nil {
		api.LogError(err, eh.APIURLCheckError)(c)
		return
	}
	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.LogError(err, eh.IPFSConnectionError)(c)
		return
	}
	if err = manager.PublishPubSubMessage(topic, forms["message"]); err != nil {
		api.LogError(err, eh.IPFSPubSubPublishError)(c)
		return
	}

	api.LogWithUser(username).Info("private ipfs pub sub message published")

	Respond(c, http.StatusOK, gin.H{"response": gin.H{"topic": topic, "message": forms["message"]}})
}

// GetLocalPinsForHostedIPFSNetwork is used to get local pins from the serving private ipfs node
func (api *API) getLocalPinsForHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, eh.UnAuthorizedAdminAccess)
		return
	}
	forms := api.extractPostForms(c, "network_name")
	if len(forms) == 0 {
		return
	}
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c)
		return
	}
	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	apiURL, err := im.GetAPIURLByName(forms["network_name"])
	if err != nil {
		api.LogError(err, eh.APIURLCheckError)(c)
		return
	}
	// initialize a connection toe the local ipfs node
	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.LogError(err, eh.IPFSConnectionError)(c)
		return
	}
	// get all the known local pins
	// WARNING: THIS COULD BE A VERY LARGE LIST
	pinInfo, err := manager.Shell.Pins()
	if err != nil {
		api.LogError(err, eh.IPFSPinParseError)(c)
		return
	}
	api.LogWithUser(username).Info("private ipfs pin list requested")
	Respond(c, http.StatusOK, gin.H{"response": pinInfo})
}

// GetObjectStatForIpfsForHostedIPFSNetwork is  used to get object stats from a private ipfs network
func (api *API) getObjectStatForIpfsForHostedIPFSNetwork(c *gin.Context) {
	key := c.Param("key")
	if _, err := gocid.Decode(key); err != nil {
		Fail(c, err)
		return
	}
	username := GetAuthenticatedUserFromContext(c)
	forms := api.extractPostForms(c, "network_name")
	if len(forms) == 0 {
		return
	}
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c)
		return
	}

	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	apiURL, err := im.GetAPIURLByName(forms["network_name"])
	if err != nil {
		api.LogError(err, eh.APIURLCheckError)(c)
		return
	}

	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.LogError(err, eh.IPFSConnectionError)(c)
		return
	}
	stats, err := manager.ObjectStat(key)
	if err != nil {
		api.LogError(err, eh.IPFSObjectStatError)(c)
		return
	}
	api.LogWithUser(username).Info("private ipfs object stat requested")
	Respond(c, http.StatusOK, gin.H{"response": stats})
}

// CheckLocalNodeForPinForHostedIPFSNetwork is used to check the serving node for a pin
func (api *API) checkLocalNodeForPinForHostedIPFSNetwork(c *gin.Context) {
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, eh.UnAuthorizedAdminAccess)
		return
	}
	forms := api.extractPostForms(c, "network_name")
	if len(forms) == 0 {
		return
	}
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c)
		return
	}
	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	apiURL, err := im.GetAPIURLByName(forms["network_name"])
	if err != nil {
		api.LogError(err, eh.APIURLCheckError)(c)
		return
	}
	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.LogError(err, eh.IPFSConnectionError)(c)
		return
	}
	present, err := manager.ParseLocalPinsForHash(hash)
	if err != nil {
		api.LogError(err, eh.IPFSPinParseError)(c)
		return
	}
	api.LogWithUser(username).Info("private ipfs pin check requested")
	Respond(c, http.StatusOK, gin.H{"response": present})
}

// PublishDetailedIPNSToHostedIPFSNetwork is used to publish an IPNS record to a private network with fine grained control
func (api *API) publishDetailedIPNSToHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	forms := api.extractPostForms(c, "network_name", "hash", "life_time", "ttl", "key", "resolve")
	if len(forms) == 0 {
		return
	}
	cost, err := utils.CalculateAPICallCost("ipns", true)
	if err != nil {
		api.LogError(err, eh.CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	mqURL := api.cfg.RabbitMQ.URL
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c)
		return
	}
	qm, err := queue.Initialize(queue.IpnsEntryQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c)
		return
	}
	if _, err := gocid.Decode(forms["hash"]); err != nil {
		Fail(c, err)
		return
	}
	ownsKey, err := api.um.CheckIfKeyOwnedByUser(username, forms["key"])
	if err != nil {
		api.LogError(err, eh.KeySearchError)(c)
		return
	}
	if !ownsKey {
		err = fmt.Errorf("unauthorized access to key by user %s", username)
		api.LogError(err, eh.KeyUseError)(c)
		return
	}
	resolve, err := strconv.ParseBool(forms["resolve"])
	if err != nil {
		// user error, dont log
		Fail(c, err)
		return
	}
	lifetime, err := time.ParseDuration(forms["life_time"])
	if err != nil {
		// user error, dont log
		Fail(c, err)
		return
	}
	ttl, err := time.ParseDuration(forms["ttl"])
	if err != nil {
		// user error, dont log
		Fail(c, err)
		return
	}
	ipnsUpdate := queue.IPNSEntry{
		CID:         forms["hash"],
		LifeTime:    lifetime,
		TTL:         ttl,
		Key:         forms["key"],
		Resolve:     resolve,
		NetworkName: forms["network_name"],
		UserName:    username,
		CreditCost:  cost,
	}
	if err := qm.PublishMessage(ipnsUpdate); err != nil {
		api.LogError(err, eh.QueuePublishError)(c)
		return
	}
	api.LogWithUser(username).Info("private ipns entry creation request sent to backend")
	Respond(c, http.StatusOK, gin.H{"response": "ipns entry creation request sent to backend"})
}

// CreateHostedIPFSNetworkEntryInDatabase is used to create an entry in the database for a private ipfs network
func (api *API) createHostedIPFSNetworkEntryInDatabase(c *gin.Context) {
	// retrieve fields
	username := GetAuthenticatedUserFromContext(c)
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}
	logger := api.LogWithUser(username).WithField("network_name", networkName)
	logger.Info("network creation request received")

	// retrieve parameters - thse are all optional
	swarmKey, _ := c.GetPostForm("swarm_key")
	bPeers, _ := c.GetPostFormArray("bootstrap_peers")
	users := c.PostFormArray("users")
	if users == nil {
		users = []string{username}
	} else {
		users = append(users, username)
	}

	// create entry for network
	manager := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	network, err := manager.CreateHostedPrivateNetwork(networkName, swarmKey, bPeers, users)
	if err != nil {
		api.LogError(err, eh.NetworkCreationError)(c)
		return
	}
	logger.WithField("db_id", network.ID).Info("database entry created")

	// add network to users
	if err := api.um.AddIPFSNetworkForUser(username, networkName); err != nil {
		api.LogError(err, eh.NetworkCreationError)(c)
		return
	}
	logger.WithField("user", username).Info("network added to user")
	if len(users) > 0 {
		for _, v := range users {
			if err := api.um.AddIPFSNetworkForUser(v, networkName); err != nil {
				api.LogError(err, eh.NetworkCreationError)(c)
				return
			}
			logger.WithField("user", v).Info("network added to user")
		}
	}

	// request orchestrator to start up network
	resp, err := api.orch.StartNetwork(c, &ipfs_orchestrator.NetworkRequest{
		Network: networkName,
	})
	if err != nil {
		api.LogError(err, "failed to start private network",
			"network_name", networkName,
		)(c)
	}
	logger.WithField("response", resp).Info("network node started")

	// respond with network details
	Respond(c, http.StatusOK, gin.H{
		"response": gin.H{
			"id":           network.ID,
			"network_name": networkName,
			"api_url":      resp.GetApi(),
			"swarm_key":    resp.GetSwarmKey(),
			"users":        network.Users,
		},
	})
}

func (api *API) stopIPFSPrivateNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}
	logger := api.LogWithUser(username).WithField("network_name", networkName)
	logger.Info("private ipfs network shutdown requested")

	// retrieve authorized networks to check if person has access
	networks, err := api.um.GetPrivateIPFSNetworksForUser(username)
	if err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c)
		return
	}
	var found bool
	for _, n := range networks {
		if n == networkName {
			found = true
			break
		}
	}
	if !found {
		logger.Info("user not authorized to access network")
		Respond(c, http.StatusUnauthorized, gin.H{
			"response": "user does not have access to requested network",
		})
		return
	}

	if _, err := api.orch.StopNetwork(c, &ipfs_orchestrator.NetworkRequest{
		Network: networkName}); err != nil {
		api.LogError(err, "failed to stop network")(c)
		return
	}
	logger.Info("network stopped")

	Respond(c, http.StatusOK, gin.H{
		"response": gin.H{
			"network_name": networkName,
		},
	})
}

// GetIPFSPrivateNetworkByName is used to get connection information for a priavate ipfs network
func (api *API) getIPFSPrivateNetworkByName(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, eh.UnAuthorizedAdminAccess)
		return
	}
	netName := c.Param("name")
	logger := api.LogWithUser(username).WithField("netowrk_name", netName)
	logger.Info("private ipfs network by name requested")

	// retrieve details from database
	manager := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	net, err := manager.GetNetworkByName(netName)
	if err != nil {
		api.LogError(err, eh.NetworkSearchError)(c)
		return
	}

	// retrieve additional stats if requested
	if c.Param("stats") == "true" {
		logger.Info("retrieving additional stats from orchestrator")
		stats, err := api.orch.NetworkStats(c, &ipfs_orchestrator.NetworkRequest{Network: netName})
		if err != nil {
			api.LogError(err, eh.NetworkSearchError)(c)
			return
		}

		Respond(c, http.StatusOK, gin.H{"response": gin.H{
			"database":      net,
			"network_stats": stats,
		}})
	} else {
		Respond(c, http.StatusOK, gin.H{"response": gin.H{
			"database": net,
		}})
	}
}

// GetAuthorizedPrivateNetworks is used to get the private
// networks a user is authorized for
func (api *API) getAuthorizedPrivateNetworks(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)

	networks, err := api.um.GetPrivateIPFSNetworksForUser(username)
	if err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c)
		return
	}

	api.LogWithUser(username).Info("authorized private ipfs network listing requested")
	Respond(c, http.StatusOK, gin.H{"response": networks})
}

// getUploadsByNetworkName is used to get uploads for a network by its name
func (api *API) getUploadsByNetworkName(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	networkName := c.Param("network_name")
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c)
		return
	}
	um := models.NewUploadManager(api.dbm.DB)
	uploads, err := um.FindUploadsByNetwork(networkName)
	if err != nil {
		api.LogError(err, eh.UploadSearchError)(c)
		return
	}

	api.LogWithUser(username).Info("uploads forprivate ifps network requested")
	Respond(c, http.StatusOK, gin.H{"response": uploads})
}

// DownloadContentHashForPrivateNetwork is used to download content from  a private ipfs network
func (api *API) downloadContentHashForPrivateNetwork(c *gin.Context) {
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}

	username := GetAuthenticatedUserFromContext(c)

	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c)
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

	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.LogError(err, eh.APIURLCheckError)(c)
		return
	}
	// get the content hash that is to be downloaded
	contentHash := c.Param("hash")
	if _, err := gocid.Decode(contentHash); err != nil {
		Fail(c, err)
		return
	}
	// initialize our connection to IPFS
	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.LogError(err, eh.IPFSConnectionError)(c)
		return
	}
	// read the contents of the file
	reader, err := manager.Shell.Cat(contentHash)
	if err != nil {
		api.LogError(err, eh.IPFSCatError)(c)
		return
	}
	// get the size of hte file in bytes
	sizeInBytes, err := manager.GetObjectFileSizeInBytes(contentHash)
	if err != nil {
		api.LogError(err, eh.IPFSObjectStatError)(c)
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
			Fail(c, errors.New("extra_headers post form is not even in length"))
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

	api.LogWithUser(username).Info("private ipfs content download served")
	c.DataFromReader(200, int64(sizeInBytes), contentType, reader, extraHeaders)
}
