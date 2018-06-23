package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
)

func CreateIPFSNetworkEntryInDatabase(c *gin.Context) {
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

	isHosted, exists := cC.GetPostForm("is_hosted")
	if !exists {
		FailNoExist(c, "is_hosted post form does not exist")
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

	var hosted bool
	switch isHosted {
	case "true":
		hosted = true
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
			valid, err := utils.ParseMultiAddrForBootstrap(addr)
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
			_, err = utils.GenerateMultiAddrFromString(nodeAddresses[k])
			if err != nil {
				FailOnError(c, err)
				return
			}
			bootstrapPeerAddresses = append(bootstrapPeerAddresses, v)
			localNodeAddresses = append(localNodeAddresses, nodeAddresses[k])
		}
	case "false":
		hosted = false
		for _, v := range nodeAddresses {
			_, err := utils.GenerateMultiAddrFromString(v)
			if err != nil {
				FailOnError(c, err)
				return
			}
			localNodeAddresses = append(localNodeAddresses, v)
		}
	default:
		FailOnError(c, errors.New("is_hosted must be `true` or `false`"))
		return
	}
	// previously we were initializing like `var args map[string]*[]string` which was causing some issues.
	args := make(map[string]*[]string)
	args["local_node_addresses"] = &localNodeAddresses
	if len(bootstrapPeerAddresses) > 0 {
		args["bootstrap_addresses"] = &bootstrapPeerAddresses
	}
	db, ok := cC.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to load database",
		})
		return
	}
	manager := models.NewIPFSNetworkManager(db)
	network, err := manager.CreatePrivateNetwork(networkName, apiURL, swarmKey, hosted, args, users)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"network": network,
	})
}
