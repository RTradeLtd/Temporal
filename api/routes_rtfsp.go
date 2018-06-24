package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
)

func UploadToHostedIPFSNetwork(c *gin.Context) {
	cC := c.Copy()
	networkName, exists := cC.GetPostForm("network_name")
	hash := cC.Param("hash")
	ethAddress := GetAuthenticatedUserFromContext(cC)
	holdTimeInMonths, exists := cC.GetPostForm("hold_time")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "hold_time post form does not exist",
		})
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"hash":                hash,
		"eth_address":         ethAddress,
		"hold_time_in_months": holdTimeInt,
		"network_name":        networkName,
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
