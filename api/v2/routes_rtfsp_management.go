package v2

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/database/models"
	nexus "github.com/RTradeLtd/grpc/nexus"
	"github.com/gin-gonic/gin"
)

// these API calls are used to handle management of private IPFS networks

// CreateIPFSNetwork is used to create an entry in the database for a private ipfs network
func (api *API) createIPFSNetwork(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithMissingField(c, "network_name")
		return
	}
	if strings.ToLower(networkName) == "public" {
		Fail(c, errors.New("network name can't be public, or PUBLIC"))
	}
	logger := api.l.With("user", username, "network_name", networkName)
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

	network, err := api.nm.CreateHostedPrivateNetwork(networkName, swarmKey, bPeers, models.NetworkAccessOptions{Users: users})
	if err != nil {
		api.LogError(c, err, eh.NetworkCreationError)(http.StatusBadRequest)
		return
	}
	logger.With("db_id", network.ID).Info("database entry created")
	if len(users) > 0 {
		for _, v := range users {
			if err := api.um.AddIPFSNetworkForUser(v, networkName); err != nil && err.Error() != "network already configured for user" {
				api.LogError(c, err, eh.NetworkCreationError)(http.StatusBadRequest)
				return
			}
			api.l.With("user", v).Info("network added to user)")
		}
	}

	// request orchestrator to start up network
	resp, err := api.orch.StartNetwork(c, &nexus.NetworkRequest{
		Network: networkName,
	})
	if err != nil {
		api.LogError(c, err, "failed to start private network",
			"network_name", networkName,
		)(http.StatusBadRequest)
		return
	}
	api.l.With("response", resp).Info("network node started")
	// respond with network details
	Respond(c, http.StatusOK, gin.H{
		"response": gin.H{
			"id":           network.ID,
			"peer_id":      resp.GetPeerId(),
			"network_name": networkName,
			"swarm_key":    resp.GetSwarmKey(),
			"users":        network.Users,
		},
	})
}

func (api *API) startIPFSPrivateNetwork(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithMissingField(c, "network_name")
		return
	}
	logger := api.l.With("user", username, "network_name", networkName)
	logger.Info("private ipfs network start requested")

	// verify access to the requested network
	networks, err := api.um.GetPrivateIPFSNetworksForUser(username)
	if err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	var found bool
	for _, network := range networks {
		if network == networkName {
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
	if _, err := api.orch.StartNetwork(c, &nexus.NetworkRequest{
		Network: networkName}); err != nil {
		api.LogError(c, err, "failed to start network")(http.StatusBadRequest)
		return
	}
	logger.Info("network started")
	Respond(c, http.StatusOK, gin.H{
		"response": gin.H{
			"network_name": networkName,
			"state":        "started",
		},
	})
}

func (api *API) stopIPFSPrivateNetwork(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithMissingField(c, "network_name")
		return
	}
	logger := api.l.With("user", username, "network_name", networkName)
	logger.Info("private ipfs network shutdown requested")

	// retrieve authorized networks to check if person has access
	networks, err := api.um.GetPrivateIPFSNetworksForUser(username)
	if err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
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
	fmt.Println(1)
	if _, err := api.orch.StopNetwork(c, &nexus.NetworkRequest{
		Network: networkName}); err != nil {
		api.LogError(c, err, "failed to stop network")(http.StatusBadRequest)
		return
	}
	logger.Info("network stopped")
	Respond(c, http.StatusOK, gin.H{
		"response": gin.H{
			"network_name": networkName,
			"state":        "stopped",
		},
	})
}

func (api *API) removeIPFSPrivateNetwork(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithMissingField(c, "network_name")
		return
	}
	logger := api.l.With("user", username, "network_name", networkName)
	logger.Info("private ipfs network shutdown requested")
	// retrieve authorized networks to check if person has access
	networks, err := api.um.GetPrivateIPFSNetworksForUser(username)
	if err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	var found bool
	for _, n := range networks {
		if n == networkName {
			found = true
			break
		}
	}
	//TODO: make sure that only the creator of the network can remove it
	if !found {
		logger.Info("user not authorized to access network")
		Respond(c, http.StatusUnauthorized, gin.H{
			"response": "user does not have access to requested network",
		})
		return
	}
	// tell orchestrator to remove the network, and all of its data
	if _, err = api.orch.RemoveNetwork(c, &nexus.NetworkRequest{
		Network: networkName}); err != nil {
		api.LogError(c, err, "failed to remove network assets")(http.StatusBadRequest)
		return
	}
	// search for the network to get list of users who have access
	// this allows us to search through the user table, and remove the network from it
	network, err := api.nm.GetNetworkByName(networkName)
	if err != nil {
		api.LogError(c, err, eh.NetworkSearchError)(http.StatusBadRequest)
		return
	}
	// remove network from database
	if err = api.nm.Delete(networkName); err != nil {
		api.LogError(c, err, "failed to remove network from database")(http.StatusBadRequest)
		return
	}
	// remove network from users authorized networks
	for _, v := range network.Users {
		if err = api.um.RemoveIPFSNetworkForUser(v, networkName); err != nil {
			api.LogError(c, err, "failed to remove network from users")(http.StatusBadRequest)
			return
		}
	}
	logger.Info("network removed")
	Respond(c, http.StatusOK, gin.H{
		"response": gin.H{
			"network_name": networkName,
			"state":        "removed",
		},
	})
}

// GetIPFSPrivateNetworkByName is used to get connection information for a private ipfs network
func (api *API) getIPFSPrivateNetworkByName(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	netName := c.Param("name")
	networks, err := api.um.GetPrivateIPFSNetworksForUser(username)
	if err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	var found bool
	for _, v := range networks {
		if v == netName {
			found = true
			break
		}
	}
	if !found {
		Fail(c, errors.New(eh.PrivateNetworkAccessError))
		return
	}
	logger := api.l.With("user", username, "network_name", netName)
	logger.Info("private ipfs network by name requested")
	net, err := api.nm.GetNetworkByName(netName)
	if err != nil {
		api.LogError(c, err, eh.NetworkSearchError)(http.StatusBadRequest)
		return
	}

	// retrieve additional stats if requested
	if c.Param("stats") == "true" {
		logger.Info("retrieving additional stats from orchestrator")
		stats, err := api.orch.NetworkStats(c, &nexus.NetworkRequest{Network: netName})
		if err != nil {
			api.LogError(c, err, eh.NetworkSearchError)(http.StatusBadRequest)
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

// GetAuthorizedPrivateNetworks is used to retrieve authorized private networks
// an authorized private network is defined as a network a user has API access to
func (api *API) getAuthorizedPrivateNetworks(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	networks, err := api.um.GetPrivateIPFSNetworksForUser(username)
	if err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}

	api.l.Infow("authorized private ipfs network listing requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": networks})
}
