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
	minio "github.com/minio/minio-go"
	log "github.com/sirupsen/logrus"

	"github.com/RTradeLtd/Temporal/models"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
)

// PinToHostedIPFSNetwork is used to pin content to a private/hosted ipfs network
func (api *API) pinToHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
	}

	err := CheckAccessForPrivateNetwork(username, networkName, api.DBM.DB)
	if err != nil {
		api.Logger.Warn(err)
		FailOnError(c, err)
		return
	}
	hash := c.Param("hash")

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
		NetworkName:      networkName,
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
	}

	mqConnectionURL := api.TConfig.RabbitMQ.URL

	qm, err := queue.Initialize(queue.IpfsPinQueue, mqConnectionURL, true)
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
	}).Info("ipfs pin request for private network sent to backend")

	c.JSON(http.StatusOK, gin.H{
		"status": "content pin request sent to backend",
	})
}

// GetFileSizeInBytesForObjectForHostedIPFSNetwork is used to get file size for an object
// on a private IPFS network
func (api *API) getFileSizeInBytesForObjectForHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}
	err := CheckAccessForPrivateNetwork(username, networkName, api.DBM.DB)
	if err != nil {
		api.Logger.Warn(err)
		FailOnError(c, err)
		return
	}

	im := models.NewHostedIPFSNetworkManager(api.DBM.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	key := c.Param("key")
	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
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
	}).Info("private ipfs object file size requested")

	c.JSON(http.StatusOK, gin.H{
		"object":        key,
		"size_in_bytes": sizeInBytes,
	})

}

// AddFileToHostedIPFSNetworkAdvanced is used to add a file to a hosted ipfs network in a more advanced and resilient manner
func (api *API) addFileToHostedIPFSNetworkAdvanced(c *gin.Context) {

	username := GetAuthenticatedUserFromContext(c)

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}

	err := CheckAccessForPrivateNetwork(username, networkName, api.DBM.DB)
	if err != nil {
		api.Logger.Warn(err)
		FailOnError(c, err)
		return
	}

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
		FailOnError(c, err)
		return
	}
	fmt.Println("file opened")
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
		NetworkName:      networkName,
		HoldTimeInMonths: holdTimeInMonths,
	}
	qm, err := queue.Initialize(queue.IpfsFileQueue, mqURL, true)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// we don't use an exchange for file publishes so that rabbitmq distributes round robin
	err = qm.PublishMessage(ifp)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("advanced private ipfs file upload requested")

	c.JSON(http.StatusOK, gin.H{"status": "file upload request sent to backend"})
}

// AddFileToHostedIPFSNetwork is used to add a file to a private IPFS network
func (api *API) addFileToHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}

	err := CheckAccessForPrivateNetwork(username, networkName, api.DBM.DB)
	if err != nil {
		FailOnError(c, err)
		return
	}

	mqURL := api.TConfig.RabbitMQ.URL

	holdTimeinMonths, exists := c.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTimeinMonths, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}
	im := models.NewHostedIPFSNetworkManager(api.DBM.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	ipfsManager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	qm, err := queue.Initialize(queue.DatabaseFileAddQueue, mqURL, true)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	fmt.Println("fetching file")
	// fetch the file, and create a handler to interact with it
	fileHandler, err := c.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}

	file, err := fileHandler.Open()
	if err != nil {
		FailOnError(c, err)
		return
	}
	resp, err := ipfsManager.Add(file)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	fmt.Println("file uploaded")
	dfa := queue.DatabaseFileAdd{
		Hash:             resp,
		HoldTimeInMonths: holdTimeInt,
		UserName:         username,
		NetworkName:      networkName,
	}
	fmt.Printf("+%v\n", dfa)
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
		HoldTimeInMonths: holdTimeInt,
	}

	qm, err = queue.Initialize(queue.IpfsPinQueue, mqURL, true)
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
	}).Info("simple private ipfs file upload processed")

	c.JSON(http.StatusOK, gin.H{
		"status": resp,
	})
}

// IpfsPubSubPublishToHostedIPFSNetwork is used to publish a pubsub message
// to a private ipfs network
func (api *API) ipfsPubSubPublishToHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}
	err := CheckAccessForPrivateNetwork(username, networkName, api.DBM.DB)
	if err != nil {
		FailOnError(c, err)
	}

	im := models.NewHostedIPFSNetworkManager(api.DBM.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	topic := c.Param("topic")
	message, present := c.GetPostForm("message")
	if !present {
		FailNoExistPostForm(c, "message")
		return
	}
	manager, err := rtfs.Initialize("", apiURL)
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
	}).Info("private ipfs pub sub message published")

	c.JSON(http.StatusOK, gin.H{
		"topic":   topic,
		"message": message,
	})
}

// RemovePinFromLocalHostForHostedIPFSNetwork is used to remove a content hash from a private hosted ipfs network
func (api *API) removePinFromLocalHostForHostedIPFSNetwork(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	hash := c.Param("hash")
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}
	err := CheckAccessForPrivateNetwork(username, networkName, api.DBM.DB)
	if err != nil {
		api.Logger.Warn(err)
		FailOnError(c, err)
	}
	rm := queue.IPFSPinRemoval{
		ContentHash: hash,
		NetworkName: networkName,
		UserName:    username,
	}
	mqConnectionURL := api.TConfig.RabbitMQ.URL
	qm, err := queue.Initialize(queue.IpfsPinRemovalQueue, mqConnectionURL, true)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
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
	}).Info("private ipfs pin removal request sent to backend")

	c.JSON(http.StatusOK, gin.H{
		"status": "pin removal sent to backend",
	})
}

// GetLocalPinsForHostedIPFSNetwork is used to get local pins
// for a private ipfs network
func (api *API) getLocalPinsForHostedIPFSNetwork(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}
	err := CheckAccessForPrivateNetwork(ethAddress, networkName, api.DBM.DB)
	if err != nil {
		api.Logger.Warn(err)
		FailOnError(c, err)
		return
	}
	im := models.NewHostedIPFSNetworkManager(api.DBM.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// initialize a connection toe the local ipfs node
	manager, err := rtfs.Initialize("", apiURL)
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
	}).Info("private ipfs pin list requested")

	c.JSON(http.StatusOK, gin.H{"pins": pinInfo})
}

// GetObjectStatForIpfsForHostedIPFSNetwork is  used to get object
// stats for a private ipfs network
func (api *API) getObjectStatForIpfsForHostedIPFSNetwork(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}
	err := CheckAccessForPrivateNetwork(ethAddress, networkName, api.DBM.DB)
	if err != nil {
		api.Logger.Warn(err)
		FailOnError(c, err)
		return
	}

	im := models.NewHostedIPFSNetworkManager(api.DBM.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	key := c.Param("key")
	manager, err := rtfs.Initialize("", apiURL)
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
		"user":    ethAddress,
	}).Info("private ipfs object stat requested")

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// CheckLocalNodeForPinForHostedIPFSNetwork is used to check a
// private ipfs network for a partilcar pin
func (api *API) checkLocalNodeForPinForHostedIPFSNetwork(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}

	err := CheckAccessForPrivateNetwork(ethAddress, networkName, api.DBM.DB)
	if err != nil {
		api.Logger.Warn(err)
		FailOnError(c, err)
		return
	}
	im := models.NewHostedIPFSNetworkManager(api.DBM.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		FailOnError(c, err)
		return
	}
	hash := c.Param("hash")
	manager, err := rtfs.Initialize("", apiURL)
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
	}).Info("private ipfs pin check requested")

	c.JSON(http.StatusOK, gin.H{"present": present})
}

// PublishDetailedIPNSToHostedIPFSNetwork is used to publish
// an IPNS record to a private network with fine grained control
func (api *API) publishDetailedIPNSToHostedIPFSNetwork(c *gin.Context) {

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}

	ethAddress := GetAuthenticatedUserFromContext(c)

	mqURL := api.TConfig.RabbitMQ.URL

	err := CheckAccessForPrivateNetwork(ethAddress, networkName, api.DBM.DB)
	if err != nil {
		api.Logger.Warn(err)
		FailOnError(c, err)
		return
	}

	um := models.NewUserManager(api.DBM.DB)
	qm, err := queue.Initialize(queue.IpnsEntryQueue, mqURL, true)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	hash, present := c.GetPostForm("hash")
	if !present {
		FailNoExistPostForm(c, "hash")
		return
	}
	lifetimeStr, present := c.GetPostForm("life_time")
	if !present {
		FailNoExistPostForm(c, "lifetime")
		return
	}
	ttlStr, present := c.GetPostForm("ttl")
	if !present {
		FailNoExistPostForm(c, "ttl")
		return
	}
	key, present := c.GetPostForm("key")
	if !present {
		FailNoExistPostForm(c, "key")
		return
	}
	resolveString, present := c.GetPostForm("resolve")
	if !present {
		FailNoExistPostForm(c, "resolve")
		return
	}

	ownsKey, err := um.CheckIfKeyOwnedByUser(ethAddress, key)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	if !ownsKey {
		api.Logger.Warnf("user %s attempting to generate IPNS entry unowned key", ethAddress)
		FailOnError(c, errors.New("attempting to generate IPNS entry unowned key"))
		return
	}

	resolve, err := strconv.ParseBool(resolveString)
	if err != nil {
		FailOnError(c, err)
		return
	}
	lifetime, err := time.ParseDuration(lifetimeStr)
	if err != nil {
		FailOnError(c, err)
		return
	}
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		FailOnError(c, err)
		return
	}
	keyID, err := um.GetKeyIDByName(ethAddress, key)
	fmt.Println("using key id of ", keyID)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	ipnsUpdate := queue.IPNSEntry{
		CID:         hash,
		LifeTime:    lifetime,
		TTL:         ttl,
		Key:         key,
		Resolve:     resolve,
		NetworkName: networkName,
	}
	err = qm.PublishMessage(ipnsUpdate)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("private ipns entry creation request sent to backend")

	c.JSON(http.StatusOK, gin.H{"status": "ipns entry creation request sent to backend"})
}

// CreateHostedIPFSNetworkEntryInDatabase is used to create
// an entry in the database for a private ipfs network
func (api *API) createHostedIPFSNetworkEntryInDatabase(c *gin.Context) {
	// lock down as admin route for now
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access")
		return
	}

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}

	apiURL, exists := c.GetPostForm("api_url")
	if !exists {
		FailNoExistPostForm(c, "api_url")
		return
	}

	swarmKey, exists := c.GetPostForm("swarm_key")
	if !exists {
		FailNoExistPostForm(c, "swarm_key")
		return
	}

	bPeers, exists := c.GetPostFormArray("bootstrap_peers")
	if !exists {
		FailNoExistPostForm(c, "bootstrap_peers")
		return
	}
	nodeAddresses, exists := c.GetPostFormArray("local_node_addresses")
	if !exists {
		FailNoExistPostForm(c, "local_node_addresses")
		return
	}
	users := c.PostFormArray("users")
	var localNodeAddresses []string
	var bootstrapPeerAddresses []string

	if len(nodeAddresses) != len(bPeers) {
		FailOnError(c, errors.New("length of local_node_addresses and bootstrap_peers must be equal"))
		return
	}
	for k, v := range bPeers {
		addr, err := utils.GenerateMultiAddrFromString(v)
		if err != nil {
			api.Logger.Error(err)
			FailOnError(c, err)
			return
		}
		valid, err := utils.ParseMultiAddrForIPFSPeer(addr)
		if err != nil {
			api.Logger.Error(err)
			FailOnError(c, err)
			return
		}
		if !valid {
			api.Logger.Errorf("provided peer %s is not a valid bootstrap peer", addr)
			FailOnError(c, fmt.Errorf("provided peer %s is not a valid bootstrap peer", addr))
			return
		}
		addr, err = utils.GenerateMultiAddrFromString(nodeAddresses[k])
		if err != nil {
			api.Logger.Error(err)
			FailOnError(c, err)
			return
		}
		valid, err = utils.ParseMultiAddrForIPFSPeer(addr)
		if err != nil {
			api.Logger.Error(err)
			FailOnError(c, err)
			return
		}
		if !valid {
			api.Logger.Errorf("provided peer %s is not a valid ipfs peer", addr)
			FailOnError(c, fmt.Errorf("provided peer %s is not a valid ipfs peer", addr))
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
	manager := models.NewHostedIPFSNetworkManager(api.DBM.DB)
	network, err := manager.CreateHostedPrivateNetwork(networkName, apiURL, swarmKey, args, users)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	um := models.NewUserManager(api.DBM.DB)

	if len(users) > 0 {
		for _, v := range users {
			err := um.AddIPFSNetworkForUser(v, networkName)
			if err != nil {
				api.Logger.Error(err)
				FailOnError(c, err)
				return
			}
		}
	} else {
		err := um.AddIPFSNetworkForUser(AdminAddress, networkName)
		if err != nil {
			api.Logger.Error(err)
			FailOnError(c, err)
			return
		}
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("private ipfs netwokr created")

	c.JSON(http.StatusCreated, gin.H{
		"network": network,
	})

}

// GetIPFSPrivateNetworkByName is used to get connection information for a priavate ipfs network
func (api *API) getIPFSPrivateNetworkByName(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access")
		return
	}

	netName := c.Param("name")
	manager := models.NewHostedIPFSNetworkManager(api.DBM.DB)
	net, err := manager.GetNetworkByName(netName)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("private ipfs network by name requested")

	c.JSON(http.StatusOK, gin.H{
		"network": net,
	})
}

// GetAuthorizedPrivateNetworks is used to get the private
// networks a user is authorized for
func (api *API) getAuthorizedPrivateNetworks(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)

	um := models.NewUserManager(api.DBM.DB)
	networks, err := um.GetPrivateIPFSNetworksForUser(ethAddress)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("authorized private ipfs network listing requested")

	c.JSON(http.StatusOK, gin.H{
		"networks": networks,
	})
}

// GetUploadsByNetworkName is used to getu plaods for a network by its name
func (api *API) getUploadsByNetworkName(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}

	err := CheckAccessForPrivateNetwork(ethAddress, networkName, api.DBM.DB)
	if err != nil {
		api.Logger.Warn(err)
		FailOnError(c, err)
		return
	}

	um := models.NewUploadManager(api.DBM.DB)
	uploads, err := um.FindUploadsByNetwork(networkName)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("uploads forprivate ifps network requested")

	c.JSON(http.StatusOK, gin.H{
		"uploads": uploads,
	})
}

// DownloadContentHashForPrivateNetwork is used to download content from  a private ipfs network
func (api *API) downloadContentHashForPrivateNetwork(c *gin.Context) {
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}

	ethAddress := GetAuthenticatedUserFromContext(c)

	err := CheckAccessForPrivateNetwork(ethAddress, networkName, api.DBM.DB)
	if err != nil {
		api.Logger.Warn(err)
		FailOnError(c, err)
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

	im := models.NewHostedIPFSNetworkManager(api.DBM.DB)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// get the content hash that is to be downloaded
	contentHash := c.Param("hash")

	// initialize our connection to IPFS
	manager, err := rtfs.Initialize("", apiURL)
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
		"user":    ethAddress,
	}).Info("private ipfs content download requested")

	// send them the file
	c.DataFromReader(200, int64(sizeInBytes), contentType, reader, extraHeaders)
}
