package v2

import (
	"errors"
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
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract network name
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithMissingField(c, "network_name")
		return
	}
	// make sure the name is something other than public
	if strings.ToLower(networkName) == "public" {
		Fail(c, errors.New("network name can't be public, or PUBLIC"))
	}
	// retrieve parameters - thse are all optional
	swarmKey, _ := c.GetPostForm("swarm_key")
	bPeers, _ := c.GetPostFormArray("bootstrap_peers")
	users := c.PostFormArray("users")
	if users == nil {
		users = []string{username}
	} else {
		users = append(users, username)
	}
	// create the network in our database
	network, err := api.nm.CreateHostedPrivateNetwork(networkName, swarmKey, bPeers, models.NetworkAccessOptions{Users: users, Owner: username})
	if err != nil {
		api.LogError(c, err, eh.NetworkCreationError)(http.StatusBadRequest)
		return
	}
	// request orchestrator to start up network and create it after registering it in the database
	resp, err := api.orch.StartNetwork(c, &nexus.NetworkRequest{
		Network: networkName,
	})
	if err != nil {
		api.LogError(c, err, "failed to start private network",
			"network_name", networkName,
		)(http.StatusBadRequest)
		return
	}
	logger := api.l.With("user", username, "network_name", networkName)
	logger.Info("network creation request received")
	logger.With("db_id", network.ID).Info("database entry created")
	// update allows users who can access the network
	if len(users) > 0 {
		for _, v := range users {
			if err := api.um.AddIPFSNetworkForUser(v, networkName); err != nil && err.Error() != "network already configured for user" {
				api.LogError(c, err, eh.NetworkCreationError)(http.StatusBadRequest)
				return
			}
			api.l.With("user", v).Info("network added to user)")
		}
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
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get network name
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithMissingField(c, "network_name")
		return
	}
	logger := api.l.With("user", username, "network_name", networkName)
	logger.Info("private ipfs network start requested")
	if err := api.isNetworkOwner(networkName, username); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusUnauthorized)
		return
	}
	if _, err := api.orch.StartNetwork(c, &nexus.NetworkRequest{
		Network: networkName}); err != nil {
		api.LogError(c, err, "failed to start network")(http.StatusBadRequest)
		return
		return
	}
	// log and return
	logger.Info("network started")
	Respond(c, http.StatusOK, gin.H{
		"response": gin.H{
			"network_name": networkName,
			"state":        "started",
		},
	})
}

func (api *API) stopIPFSPrivateNetwork(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get network name
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithMissingField(c, "network_name")
		return
	}
	logger := api.l.With("user", username, "network_name", networkName)
	logger.Info("private ipfs network shutdown requested")
	// verify admin access to network
	if err := api.isNetworkOwner(networkName, username); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusUnauthorized)
		return
	}
	// send a stop network request
	if _, err := api.orch.StopNetwork(c, &nexus.NetworkRequest{
		Network: networkName}); err != nil {
		api.LogError(c, err, "failed to stop network")(http.StatusBadRequest)
		return
	}
	// log and return
	logger.Info("network stopped")
	Respond(c, http.StatusOK, gin.H{
		"response": gin.H{
			"network_name": networkName,
			"state":        "stopped",
		},
	})
}

func (api *API) removeIPFSPrivateNetwork(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get the network name
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailWithMissingField(c, "network_name")
		return
	}
	logger := api.l.With("user", username, "network_name", networkName)
	logger.Info("private ipfs network shutdown requested")
	// verify admin access to network
	network, err := api.nm.GetNetworkByName(networkName)
	if err != nil {
		api.LogError(c, err, eh.NetworkSearchError)(http.StatusInternalServerError)
		return
	}
	if err := api.isNetworkOwner(networkName, username); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusUnauthorized)
		return
	}
	// send node removal request, removing all data stored
	// this is a DESTRUCTIVE action
	if _, err = api.orch.RemoveNetwork(c, &nexus.NetworkRequest{
		Network: networkName}); err != nil {
		api.LogError(c, err, "failed to remove network assets")(http.StatusBadRequest)
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
	// log and return
	logger.Info("network removed")
	Respond(c, http.StatusOK, gin.H{
		"response": gin.H{
			"network_name": networkName,
			"state":        "removed",
		},
	})
}

// GetIPFSPrivateNetworkByName is used to private ipfs network information
func (api *API) getIPFSPrivateNetworkByName(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get the network name
	netName := c.Param("name")
	// get all networks user has access to
	networks, err := api.um.GetPrivateIPFSNetworksForUser(username)
	if err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// ensure they can access this network
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
	// get network details from database
	net, err := api.nm.GetNetworkByName(netName)
	if err != nil {
		api.LogError(c, err, eh.NetworkSearchError)(http.StatusBadRequest)
		return
	}
	// retrieve additional stats if requested
	// otherwise send generic information from the database directly
	if c.Param("stats") == "true" {
		logger.Info("retrieving additional stats from orchestrator")
		stats, err := api.orch.NetworkStats(c, &nexus.NetworkRequest{Network: netName})
		if err != nil {
			api.LogError(c, err, eh.NetworkSearchError)(http.StatusBadRequest)
			return
		}
		// return
		Respond(c, http.StatusOK, gin.H{"response": gin.H{
			"database":      net,
			"network_stats": stats,
		}})
	} else {
		// return
		Respond(c, http.StatusOK, gin.H{"response": gin.H{
			"database": net,
		}})
	}
}

// GetAuthorizedPrivateNetworks is used to retrieve authorized private networks
// an authorized private network is defined as a network a user has API access to
func (api *API) getAuthorizedPrivateNetworks(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get all networks the user has access too
	networks, err := api.um.GetPrivateIPFSNetworksForUser(username)
	if err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.Infow("authorized private ipfs network listing requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": networks})
}

// addUsersToNetwork is used to add a user to the list of authorized users
// for a given private network.
func (api *API) addUsersToNetwork(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
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
	users, exists := c.GetPostFormArray("users")
	if !exists {
		FailWithMissingField(c, "users")
		return
	}
	network, err := api.nm.GetNetworkByName(networkName)
	if err != nil {
		api.LogError(c, err, eh.NetworkSearchError)(http.StatusInternalServerError)
		return
	}
	if err := api.isNetworkOwner(networkName, username); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusUnauthorized)
		return
	}
	// make sure the user accounts exist
	for _, user := range users {
		if _, err := api.um.FindByUserName(user); err != nil {
			api.LogError(c, err, eh.UserSearchError)(http.StatusInternalServerError)
			return
		}
	}
	// combine both the currrent list of authorized users, and the list of users to add
	network.Users = append(network.Users, users...)
	// update the users field of the database model only
	if err := api.nm.UpdateNetworkByName(networkName, map[string]interface{}{"users": network.Users}); err != nil {
		api.LogError(c, err, "failed to update authorized users for network")
		return
	}
	for _, user := range users {
		if err := api.um.AddIPFSNetworkForUser(user, networkName); err != nil {
			api.LogError(c, err, "failed to update network for user")(http.StatusInternalServerError)
			return
		}
	}
	Respond(c, http.StatusOK, gin.H{"response": "authorized user list updated"})
}

// removeUsersFromNetwork is used to remove a user from being able
// to access a network.
func (api *API) removeUsersFromNetwork(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
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
	users, exists := c.GetPostFormArray("users")
	if !exists {
		FailWithMissingField(c, "users")
		return
	}
	network, err := api.nm.GetNetworkByName(networkName)
	if err != nil {
		api.LogError(c, err, eh.NetworkSearchError)(http.StatusInternalServerError)
		return
	}
	if err := api.isNetworkOwner(networkName, username); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusUnauthorized)
		return
	}
	var (
		usersToRemove map[string]bool
		newUsers      []string
	)
	for _, user := range users {
		usersToRemove[user] = true
	}
	// iterate over all current users
	// and compare against users to remove list
	// if they aren't found, then we add them
	// to the new users list.
	for _, user := range network.Users {
		if !usersToRemove[user] {
			newUsers = append(newUsers, user)
		}
	}
	if err := api.nm.UpdateNetworkByName(networkName, map[string]interface{}{"users": newUsers}); err != nil {
		api.LogError(c, err, "failed to update authorized users for network")
		return
	}
	for user := range usersToRemove {
		if err := api.um.RemoveIPFSNetworkForUser(user, networkName); err != nil {
			api.LogError(c, err, "failed to remove network from user")
			return
		}
	}
	Respond(c, http.StatusOK, gin.H{"response": "authorized user list updated"})
}

func (api *API) addOwnersToNetwork(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
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
	owners, exists := c.GetPostFormArray("owners")
	if !exists {
		FailWithMissingField(c, "owners")
		return
	}
	if err := api.isNetworkOwner(networkName, username); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusUnauthorized)
		return
	}
	// make sure the user accounts exist
	for _, owner := range owners {
		if _, err := api.um.FindByUserName(owner); err != nil {
			api.LogError(c, err, eh.UserSearchError)(http.StatusInternalServerError)
			return
		}
	}
	network, err := api.nm.GetNetworkByName(networkName)
	if err != nil { // remove network from users authorized networks
		api.LogError(c, err, eh.NetworkSearchError)(http.StatusInternalServerError)
		return
	}
	for _, v := range network.Users {
		if err = api.um.RemoveIPFSNetworkForUser(v, networkName); err != nil {
			api.LogError(c, err, "failed to remove network from users")(http.StatusBadRequest)
			return
		}
	}

	network.Owners = append(network.Owners, owners...)
	if err := api.nm.UpdateNetworkByName(networkName, map[string]interface{}{"owners": network.Owners}); err != nil {
		api.LogError(c, err, "failed to update network owners")
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "network owners updated"})
}

func (api *API) isNetworkOwner(network, username string) error {
	n, err := api.nm.GetNetworkByName(network)
	if err != nil {
		return err
	}
	for _, owner := range n.Owners {
		if owner == username {
			return nil
		}
	}
	return errors.New("user is not owner")
}
