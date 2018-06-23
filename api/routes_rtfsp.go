package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
)

func CreateIPFSNetworkEntryInDatabase(c *gin.Context) {
	// lock down as admin route for now
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access")
		return
	}

	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExist(c, "network_name post form does not exist")
		return
	}

	apiURL, exists := c.GetPostForm("api_url")
	if !exists {
		FailNoExist(c, "api_url post form does not exist")
		return
	}

	swarmKey, exists := c.GetPostForm("swarm_key")
	if !exists {
		FailNoExist(c, "swarm_key post form does not exist")
		return
	}

	isHosted, exists := c.GetPostForm("is_hosted")
	if !exists {
		FailNoExist(c, "is_hosted post form does not exist")
		return
	}

	bPeers, exists := c.GetPostFormArray("bootstrap_peers")
	if !exists {
		FailNoExist(c, "boostrap_peers post form array does not exist")
		return
	}
	nodeIPAddresses, exists := c.GetPostFormArray("local_node_ip_addresses")
	if !exists {
		FailNoExist(c, "local_node_ip_addresses post form array does not exist")
		return
	}
	switch isHosted {
	case "true":
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
			_, err = utils.GenerateMultiAddrFromString(nodeIPAddresses[k])
			if err != nil {
				FailOnError(c, err)
				return
			}
		}
	case "false":
		break
	default:
		FailOnError(c, errors.New("is_hosted must be `true` or `false`"))
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"network_name":            networkName,
		"api_url":                 apiURL,
		"swarm_key":               swarmKey,
		"is_hosted":               isHosted,
		"bootstrap_peers":         bPeers,
		"local_node_ip_addresses": nodeIPAddresses,
	})
}
