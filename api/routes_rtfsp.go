package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs"
	gocid "github.com/ipfs/go-cid"
	minio "github.com/minio/minio-go"
	log "github.com/sirupsen/logrus"

	"github.com/RTradeLtd/Temporal/models"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
)

// PinToHostedIPFSNetwork is used to pin content to a private ipfs network
func (api *API) pinToHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}

	err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB)
	if err != nil {
		api.LogError(err, PrivateNetworkAccessError)(c)
		return
	}
	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	url, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.LogError(err, APIURLCheckError)(c, http.StatusBadRequest)
		return
	}
	manager, err := rtfs.Initialize("", url)
	if err != nil {
		api.LogError(err, IPFSConnectionError)(c, http.StatusBadRequest)
		return
	}

	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
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
	cost, err := utils.CalculatePinCost(hash, holdTimeInt, manager.Shell, true)
	if err != nil {
		api.LogError(err, CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	ip := queue.IPFSPin{
		CID:              hash,
		NetworkName:      networkName,
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       cost,
	}

	mqConnectionURL := api.cfg.RabbitMQ.URL

	qm, err := queue.Initialize(queue.IpfsPinQueue, mqConnectionURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c)
		api.refundUserCredits(username, "private-pin", cost)
		return
	}

	if err = qm.PublishMessageWithExchange(ip, queue.PinExchange); err != nil {
		api.LogError(err, QueuePublishError)(c)
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

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}

	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(err, PrivateNetworkAccessError)(c, http.StatusBadRequest)
		return
	}

	holdTimeInMonths, exists := c.GetPostForm("hold_time")
	if !exists {
		FailWithBadRequest(c, "hold_time")
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		Fail(c, err, http.StatusBadRequest)
		return
	}
	accessKey := api.cfg.MINIO.AccessKey
	secretKey := api.cfg.MINIO.SecretKey
	endpoint := fmt.Sprintf("%s:%s", api.cfg.MINIO.Connection.IP, api.cfg.MINIO.Connection.Port)

	mqURL := api.cfg.RabbitMQ.URL

	miniManager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, false)
	if err != nil {
		api.LogError(err, MinioConnectionError)
		Fail(c, err)
		return
	}
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
		api.LogError(err, InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	fmt.Println("opening file")
	openFile, err := fileHandler.Open()
	if err != nil {
		api.LogError(err, FileOpenError)
		api.refundUserCredits(username, "private-file", cost)
		Fail(c, err)
		return
	}
	fmt.Println("file opened")
	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", username, randString)
	fmt.Println("storing file in minio")
	if _, err = miniManager.PutObject(FilesUploadBucket, objectName, openFile, fileHandler.Size, minio.PutObjectOptions{}); err != nil {
		api.LogError(err, MinioPutError)
		api.refundUserCredits(username, "private-file", cost)
		Fail(c, err)
		return
	}
	fmt.Println("file stored in minio")
	ifp := queue.IPFSFile{
		BucketName:       FilesUploadBucket,
		ObjectName:       objectName,
		UserName:         username,
		NetworkName:      networkName,
		HoldTimeInMonths: holdTimeInMonths,
		CreditCost:       cost,
	}
	qm, err := queue.Initialize(queue.IpfsFileQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)
		api.refundUserCredits(username, "private-file", cost)
		Fail(c, err)
		return
	}
	// we don't use an exchange for file publishes so that rabbitmq distributes round robin
	if err = qm.PublishMessage(ifp); err != nil {
		api.LogError(err, QueuePublishError)
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

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}

	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(err, PrivateNetworkAccessError)
		Fail(c, err)
		return
	}

	mqURL := api.cfg.RabbitMQ.URL

	holdTimeinMonths, exists := c.GetPostForm("hold_time")
	if !exists {
		FailWithBadRequest(c, "hold_time")
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTimeinMonths, 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}

	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.LogError(err, APIURLCheckError)(c)
		return
	}

	ipfsManager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.LogError(err, IPFSConnectionError)(c)
		return
	}
	qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c)
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
		api.LogError(err, InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	file, err := fileHandler.Open()
	if err != nil {
		api.LogError(err, FileOpenError)(c)
		api.refundUserCredits(username, "private-file", cost)
		return
	}
	resp, err := ipfsManager.Add(file)
	if err != nil {
		api.LogError(err, IPFSAddError)(c)
		api.refundUserCredits(username, "private-file", cost)
		return
	}
	fmt.Println("file uploaded")
	dfa := queue.DatabaseFileAdd{
		Hash:             resp,
		HoldTimeInMonths: holdTimeInt,
		UserName:         username,
		NetworkName:      networkName,
		CreditCost:       0,
	}
	if err = qm.PublishMessage(dfa); err != nil {
		api.LogError(err, QueuePublishError)(c)
		return
	}

	pin := queue.IPFSPin{
		CID:              resp,
		NetworkName:      networkName,
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       0,
	}

	qm, err = queue.Initialize(queue.IpfsPinQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c)
		return
	}
	if err = qm.PublishMessageWithExchange(pin, queue.PinExchange); err != nil {
		api.LogError(err, QueuePublishError)(c)
		return
	}

	api.LogWithUser(username).Info("simple private ipfs file upload processed")

	Respond(c, http.StatusOK, gin.H{"response": resp})
}

// IpfsPubSubPublishToHostedIPFSNetwork is used to publish a pubsub message to a private ipfs network
func (api *API) ipfsPubSubPublishToHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(err, PrivateNetworkAccessError)(c)
		return
	}
	cost, err := utils.CalculateAPICallCost("pubsub", true)
	if err != nil {
		api.LogError(err, CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.LogError(err, APIURLCheckError)(c)
		return
	}
	topic := c.Param("topic")
	message, present := c.GetPostForm("message")
	if !present {
		FailWithBadRequest(c, "message")
		return
	}
	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.LogError(err, IPFSConnectionError)(c)
		return
	}
	if err = manager.PublishPubSubMessage(topic, message); err != nil {
		api.LogError(err, IPFSPubSubPublishError)(c)
		return
	}

	api.LogWithUser(username).Info("private ipfs pub sub message published")

	Respond(c, http.StatusOK, gin.H{"response": gin.H{"topic": topic, "message": message}})
}

// GetLocalPinsForHostedIPFSNetwork is used to get local pins from the serving private ipfs node
func (api *API) getLocalPinsForHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(err, PrivateNetworkAccessError)(c)
		return
	}
	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.LogError(err, APIURLCheckError)(c)
		return
	}
	// initialize a connection toe the local ipfs node
	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.LogError(err, IPFSConnectionError)(c)
		return
	}
	// get all the known local pins
	// WARNING: THIS COULD BE A VERY LARGE LIST
	pinInfo, err := manager.Shell.Pins()
	if err != nil {
		api.LogError(err, IPFSPinParseError)(c)
		return
	}

	api.LogWithUser(username).Info("private ipfs pin list requested")
	Respond(c, http.StatusOK, gin.H{"response": pinInfo})
}

// GetObjectStatForIpfsForHostedIPFSNetwork is  used to get object stats from a private ipfs network
func (api *API) getObjectStatForIpfsForHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(err, PrivateNetworkAccessError)(c)
		return
	}

	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.LogError(err, APIURLCheckError)(c)
		return
	}
	key := c.Param("key")
	if _, err := gocid.Decode(key); err != nil {
		Fail(c, err)
		return
	}
	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.LogError(err, IPFSConnectionError)(c)
		return
	}
	stats, err := manager.ObjectStat(key)
	if err != nil {
		api.LogError(err, IPFSObjectStatError)(c)
		return
	}

	api.LogWithUser(username).Info("private ipfs object stat requested")
	Respond(c, http.StatusOK, gin.H{"response": stats})
}

// CheckLocalNodeForPinForHostedIPFSNetwork is used to check the serving node for a pin
func (api *API) checkLocalNodeForPinForHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}

	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(err, PrivateNetworkAccessError)(c)
		return
	}
	im := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.LogError(err, APIURLCheckError)(c)
		return
	}
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.LogError(err, IPFSConnectionError)(c)
		return
	}
	present, err := manager.ParseLocalPinsForHash(hash)
	if err != nil {
		api.LogError(err, IPFSPinParseError)(c)
		return
	}

	api.LogWithUser(username).Info("private ipfs pin check requested")
	Respond(c, http.StatusOK, gin.H{"response": present})
}

// PublishDetailedIPNSToHostedIPFSNetwork is used to publish an IPNS record to a private network with fine grained control
func (api *API) publishDetailedIPNSToHostedIPFSNetwork(c *gin.Context) {

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}
	username := GetAuthenticatedUserFromContext(c)
	cost, err := utils.CalculateAPICallCost("ipns", true)
	if err != nil {
		api.LogError(err, CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	mqURL := api.cfg.RabbitMQ.URL

	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(err, PrivateNetworkAccessError)(c)
		return
	}

	qm, err := queue.Initialize(queue.IpnsEntryQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c)
		return
	}
	hash, present := c.GetPostForm("hash")
	if !present {
		FailWithBadRequest(c, "hash")
		return
	}
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	lifetimeStr, present := c.GetPostForm("life_time")
	if !present {
		FailWithBadRequest(c, "lifetime")
		return
	}
	ttlStr, present := c.GetPostForm("ttl")
	if !present {
		FailWithBadRequest(c, "ttl")
		return
	}
	key, present := c.GetPostForm("key")
	if !present {
		FailWithBadRequest(c, "key")
		return
	}
	resolveString, present := c.GetPostForm("resolve")
	if !present {
		FailWithBadRequest(c, "resolve")
		return
	}

	ownsKey, err := api.um.CheckIfKeyOwnedByUser(username, key)
	if err != nil {
		api.LogError(err, KeySearchError)(c)
		return
	}

	if !ownsKey {
		err = fmt.Errorf("unauthorized access to key by user %s", username)
		api.LogError(err, KeyUseError)(c)
		return
	}

	resolve, err := strconv.ParseBool(resolveString)
	if err != nil {
		// user error, dont log
		Fail(c, err)
		return
	}
	lifetime, err := time.ParseDuration(lifetimeStr)
	if err != nil {
		// user error, dont log
		Fail(c, err)
		return
	}
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		// user error, dont log
		Fail(c, err)
		return
	}
	ipnsUpdate := queue.IPNSEntry{
		CID:         hash,
		LifeTime:    lifetime,
		TTL:         ttl,
		Key:         key,
		Resolve:     resolve,
		NetworkName: networkName,
		UserName:    username,
		CreditCost:  cost,
	}
	if err := qm.PublishMessage(ipnsUpdate); err != nil {
		api.LogError(err, QueuePublishError)(c)
		return
	}

	api.LogWithUser(username).Info("private ipns entry creation request sent to backend")

	Respond(c, http.StatusOK, gin.H{"response": "ipns entry creation request sent to backend"})
}

// CreateHostedIPFSNetworkEntryInDatabase is used to create an entry in the database for a private ipfs network
// TODO: make bootstrap peers and related config optional
func (api *API) createHostedIPFSNetworkEntryInDatabase(c *gin.Context) {
	// lock down as admin route for now
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}

	apiURL, exists := c.GetPostForm("api_url")
	if !exists {
		FailWithBadRequest(c, "api_url")
		return
	}

	swarmKey, exists := c.GetPostForm("swarm_key")
	if !exists {
		FailWithBadRequest(c, "swarm_key")
		return
	}

	bPeers, exists := c.GetPostFormArray("bootstrap_peers")
	if !exists {
		FailWithBadRequest(c, "bootstrap_peers")
		return
	}
	nodeAddresses, exists := c.GetPostFormArray("local_node_addresses")
	if !exists {
		FailWithBadRequest(c, "local_node_addresses")
		return
	}
	users := c.PostFormArray("users")
	var localNodeAddresses []string
	var bootstrapPeerAddresses []string

	if len(nodeAddresses) != len(bPeers) {
		Fail(c, errors.New("length of local_node_addresses and bootstrap_peers must be equal"))
		return
	}
	for k, v := range bPeers {
		addr, err := utils.GenerateMultiAddrFromString(v)
		if err != nil {
			// this is entirely on the user, so lets not bother logging as it will just make noise
			Fail(c, err)
			return
		}
		valid, err := utils.ParseMultiAddrForIPFSPeer(addr)
		if err != nil {
			// this is entirely on the user, so lets not bother logging as it will just make noise
			Fail(c, err)
			return
		}
		if !valid {
			api.l.Errorf("provided peer %s is not a valid bootstrap peer", addr)
			Fail(c, fmt.Errorf("provided peer %s is not a valid bootstrap peer", addr))
			return
		}
		addr, err = utils.GenerateMultiAddrFromString(nodeAddresses[k])
		if err != nil {
			// this is entirely on the user, so lets not bother logging as it will just make noise
			Fail(c, err)
			return
		}
		valid, err = utils.ParseMultiAddrForIPFSPeer(addr)
		if err != nil {
			// this is entirely on the user, so lets not bother logging as it will just make noise
			Fail(c, err)
			return
		}
		if !valid {
			// this is entirely on the user, so lets not bother logging as it will just make noise
			Fail(c, fmt.Errorf("provided peer %s is not a valid ipfs peer", addr))
			return
		}
		bootstrapPeerAddresses = append(bootstrapPeerAddresses, v)
		localNodeAddresses = append(localNodeAddresses, nodeAddresses[k])
	}
	// previously we were initializing like `var args map[string]*[]string` which was causing some issues.
	args := make(map[string][]string)
	args["local_node_peer_addresses"] = localNodeAddresses
	if len(bootstrapPeerAddresses) > 0 {
		args["bootstrap_peer_addresses"] = bootstrapPeerAddresses
	}
	manager := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	network, err := manager.CreateHostedPrivateNetwork(networkName, apiURL, swarmKey, args, users)
	if err != nil {
		api.LogError(err, NetworkCreationError)(c)
		return
	}

	if len(users) > 0 {
		for _, v := range users {
			if err := api.um.AddIPFSNetworkForUser(v, networkName); err != nil {
				api.LogError(err, NetworkCreationError)(c)
				return
			}
		}
	} else {
		if err := api.um.AddIPFSNetworkForUser(username, networkName); err != nil {
			api.LogError(err, NetworkCreationError)(c)
			return
		}
	}

	api.LogWithUser(username).Info("private ipfs network created")
	Respond(c, http.StatusOK, gin.H{"response": network})

}

// GetIPFSPrivateNetworkByName is used to get connection information for a priavate ipfs network
func (api *API) getIPFSPrivateNetworkByName(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	netName := c.Param("name")
	manager := models.NewHostedIPFSNetworkManager(api.dbm.DB)
	net, err := manager.GetNetworkByName(netName)
	if err != nil {
		api.LogError(err, NetworkSearchError)(c)
		return
	}

	api.LogWithUser(username).Info("private ipfs network by name requested")
	Respond(c, http.StatusOK, gin.H{"response": net})
}

// GetAuthorizedPrivateNetworks is used to get the private
// networks a user is authorized for
func (api *API) getAuthorizedPrivateNetworks(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)

	networks, err := api.um.GetPrivateIPFSNetworksForUser(username)
	if err != nil {
		api.LogError(err, PrivateNetworkAccessError)(c)
		return
	}

	api.LogWithUser(username).Info("authorized private ipfs network listing requested")
	Respond(c, http.StatusOK, gin.H{"response": networks})
}

// GetUploadsByNetworkName is used to getu plaods for a network by its name
func (api *API) getUploadsByNetworkName(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithBadRequest(c, "network_name")
		return
	}

	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(err, PrivateNetworkAccessError)(c)
		return
	}

	um := models.NewUploadManager(api.dbm.DB)
	uploads, err := um.FindUploadsByNetwork(networkName)
	if err != nil {
		api.LogError(err, UploadSearchError)(c)
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
		api.LogError(err, PrivateNetworkAccessError)(c)
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
		api.LogError(err, APIURLCheckError)(c)
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
		api.LogError(err, IPFSConnectionError)(c)
		return
	}
	// read the contents of the file
	reader, err := manager.Shell.Cat(contentHash)
	if err != nil {
		api.LogError(err, IPFSCatError)(c)
		return
	}
	// get the size of hte file in bytes
	sizeInBytes, err := manager.GetObjectFileSizeInBytes(contentHash)
	if err != nil {
		api.LogError(err, IPFSObjectStatError)(c)
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
