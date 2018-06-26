package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/rtfs"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
)

func PinToHostedIPFSNetwork(c *gin.Context) {
	cC := c.Copy()
	networkName, exists := cC.GetPostForm("network_name")
	if !exists {
		FailNoExist(c, "network_name post form does not exist")
		return
	}
	hash := cC.Param("hash")
	ethAddress := GetAuthenticatedUserFromContext(cC)
	holdTimeInMonths, exists := cC.GetPostForm("hold_time")
	if !exists {
		FailNoExist(c, "hold_time post form param not present")
		return
	}
	_, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, ok := cC.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to load database",
		})
		return
	}

	um := models.NewUserManager(db)
	canUpload, err := um.CheckIfUserHasAccessToNetwork(ethAddress, networkName)
	if err != nil {
		FailOnError(c, err)
		return
	}
	if !canUpload {
		FailNotAuthorized(c, "unauthorized access to private ipfs network")
		return
	}
	im := models.NewHostedIPFSNetworkManager(db)
	pnet, err := im.GetNetworkByName(networkName)
	if err != nil {
		FailOnError(c, err)
		return
	}

	apiURL := pnet.APIURL
	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		FailOnError(c, err)
		return
	}
	go func() {
		err := manager.Pin(hash)
		if err != nil {
			//TODO log and handle
			fmt.Println("Error encountered pinning content to private ipfs node", err.Error())
			return
		}
	}()
	c.JSON(http.StatusOK, gin.H{
		"status": "content pin request sent",
	})
}

func AddFileToHostedIPFSNetwork(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to private network upload")
		return
	}

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExist(c, "network_name post form does not exist")
		return
	}

	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to load database",
		})
		return
	}

	um := models.NewUserManager(db)
	canUpload, err := um.CheckIfUserHasAccessToNetwork(ethAddress, networkName)
	if err != nil {
		FailOnError(c, err)
		return
	}

	if !canUpload {
		FailNotAuthorized(c, "user not authorized to access network")
		return
	}
	holdTimeinMonths, exists := c.GetPostForm("hold_time")
	if !exists {
		FailNoExist(c, "hold_time post form does not exist")
		return
	}
	_, err = strconv.ParseInt(holdTimeinMonths, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}
	im := models.NewHostedIPFSNetworkManager(db)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		FailOnError(c, err)
		return
	}

	ipfsManager, err := rtfs.Initialize("", apiURL)
	if err != nil {
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
	resp, err := ipfsManager.Shell.Add(file)
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("file uploaded")
	c.JSON(http.StatusOK, gin.H{
		"status": resp,
	})
}

func PublishDetailedIPNSToHostedIPFSNetwork(c *gin.Context) {

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExist(c, "network_name post form does not exist")
		return
	}

	ethAddress := GetAuthenticatedUserFromContext(c)

	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to load database",
		})
		return
	}
	um := models.NewUserManager(db)
	canUpload, err := um.CheckIfUserHasAccessToNetwork(ethAddress, networkName)
	if err != nil {
		FailOnError(c, err)
		return
	}
	if !canUpload {
		FailNotAuthorized(c, "unauthorized access to private network")
		return
	}
	hash, present := c.GetPostForm("hash")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "hash post form not present",
		})
		return
	}
	lifetime, present := c.GetPostForm("life_time")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "life_time post form not present",
		})
		return
	}
	ttl, present := c.GetPostForm("ttl")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ttl post form not present",
		})
		return
	}
	key, present := c.GetPostForm("key")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "key post form not present",
		})
		return
	}
	resolveString, present := c.GetPostForm("resolve")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "resolve post form not present",
		})
		return
	}

	ownsKey, err := um.CheckIfKeyOwnedByUser(ethAddress, key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !ownsKey {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "attempting to generate IPNS entry unowned key",
		})
		return
	}

	im := models.NewHostedIPFSNetworkManager(db)
	apiURL, err := im.GetAPIURLByName(networkName)
	if err != nil {
		FailOnError(c, err)
		return
	}
	manager, err := rtfs.Initialize("", apiURL)
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("creating key store manager")
	err = manager.CreateKeystoreManager()
	if err != nil {
		FailOnError(c, err)
		return
	}
	resolve, err := strconv.ParseBool(resolveString)
	if err != nil {
		FailOnError(c, err)
		return
	}
	prePubTime := time.Now()
	keyID, err := um.GetKeyIDByName(ethAddress, key)
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("publishing to IPNS")
	resp, err := manager.PublishToIPNSDetails(hash, lifetime, ttl, key, keyID, resolve)
	if err != nil {
		FailOnError(c, err)
		return
	}
	postPubTime := time.Now()
	timeDifference := postPubTime.Sub(prePubTime)

	ipnsManager := models.NewIPNSManager(db)
	ipnsEntry, err := ipnsManager.UpdateIPNSEntry(resp.Name, resp.Value, lifetime, ttl, key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"name":                   resp.Name,
		"value":                  resp.Value,
		"time_to_create_minutes": timeDifference.Minutes(),
		"ipns_entry_model":       ipnsEntry,
	})
}

func CreateHostedIPFSNetworkEntryInDatabase(c *gin.Context) {
	// lock down as admin route for now
	cC := c.Copy()
	ethAddress := GetAuthenticatedUserFromContext(cC)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access")
		return
	}

	networkName, exists := cC.GetPostForm("network_name")
	if !exists {
		FailNoExist(c, "network_name post form does not exist")
		return
	}

	apiURL, exists := cC.GetPostForm("api_url")
	if !exists {
		FailNoExist(c, "api_url post form does not exist")
		return
	}

	swarmKey, exists := cC.GetPostForm("swarm_key")
	if !exists {
		FailNoExist(c, "swarm_key post form does not exist")
		return
	}

	bPeers, exists := cC.GetPostFormArray("bootstrap_peers")
	if !exists {
		FailNoExist(c, "boostrap_peers post form array does not exist")
		return
	}
	nodeAddresses, exists := cC.GetPostFormArray("local_node_addresses")
	if !exists {
		FailNoExist(c, "local_node_addresses post form array does not exist")
		return
	}
	users := cC.PostFormArray("users")
	var localNodeAddresses []string
	var bootstrapPeerAddresses []string

	if len(nodeAddresses) != len(bPeers) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "length of local_node_addresses and bootstrap_peers must be equal",
		})
		return
	}
	for k, v := range bPeers {
		addr, err := utils.GenerateMultiAddrFromString(v)
		if err != nil {
			FailOnError(c, err)
			return
		}
		valid, err := utils.ParseMultiAddrForIPFSPeer(addr)
		if err != nil {
			FailOnError(c, err)
			return
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("provided peer %s is not a valid bootstrap peer", addr),
			})
			return
		}
		addr, err = utils.GenerateMultiAddrFromString(nodeAddresses[k])
		if err != nil {
			FailOnError(c, err)
			return
		}
		valid, err = utils.ParseMultiAddrForIPFSPeer(addr)
		if err != nil {
			FailOnError(c, err)
			return
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("provided peer %s is not a valid ipfs peer", addr),
			})
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
	db, ok := cC.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to load database",
		})
		return
	}
	manager := models.NewHostedIPFSNetworkManager(db)
	network, err := manager.CreateHostedPrivateNetwork(networkName, apiURL, swarmKey, args, users)
	if err != nil {
		FailOnError(c, err)
		return
	}
	um := models.NewUserManager(db)

	if len(users) > 0 {
		for _, v := range users {
			err := um.AddIPFSNetworkForUser(v, networkName)
			if err != nil {
				FailOnError(c, err)
				return
			}
		}
	} else {
		err := um.AddIPFSNetworkForUser(AdminAddress, networkName)
		if err != nil {
			FailOnError(c, err)
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{
		"network": network,
	})

}

func GetIPFSPrivateNetworkByName(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access")
		return
	}
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to load database",
		})
		return
	}

	netName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExist(c, "network_name post form does not exist")
		return
	}
	manager := models.NewHostedIPFSNetworkManager(db)
	net, err := manager.GetNetworkByName(netName)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"network": net,
	})
}
