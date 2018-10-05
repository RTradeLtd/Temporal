package api

import (
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/utils"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

// PinHashToCluster is used to trigger a cluster pin of a particular CID
func (api *API) pinHashToCluster(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		api.LogError(err, IPFSConnectionError)(c, http.StatusInternalServerError)
		return
	}
	holdTime, exists := c.GetPostForm("hold_time")
	if !exists {
		FailWithMissingField(c, "hold_time")
		return
	}

	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	cost, err := utils.CalculatePinCost(hash, holdTimeInt, manager.Shell, false)
	if err != nil {
		api.LogError(err, CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	mqURL := api.cfg.RabbitMQ.URL

	qm, err := queue.Initialize(queue.IpfsClusterPinQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c)
		api.refundUserCredits(username, "cluster-pin", cost)
		return
	}

	ipfsClusterPin := queue.IPFSClusterPin{
		CID:              hash,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       cost,
	}

	if err = qm.PublishMessage(ipfsClusterPin); err != nil {
		api.LogError(err, QueuePublishError)(c)
		api.refundUserCredits(username, "cluster-pin", cost)
		return
	}

	api.LogWithUser(username).Info("cluster pin request sent to backend")
	Respond(c, http.StatusOK, gin.H{"response": "cluster pin request sent to backend"})
}

// SyncClusterErrorsLocally is used to parse through the local cluster state and sync any errors that are detected.
func (api *API) syncClusterErrorsLocally(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	// initialize a conection to the cluster
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		api.LogError(err, IPFSConnectionError)(c)
		return
	}
	// parse the local cluster status, and sync any errors, retunring the cids that were in an error state
	syncedCids, err := manager.ParseLocalStatusAllAndSync()
	if err != nil {
		api.LogError(err, IPFSClusterStatusError)(c)
		return
	}

	api.LogWithUser(username).Info("local cluster errors parsed")
	Respond(c, http.StatusOK, gin.H{"response": syncedCids})
}

// GetLocalStatusForClusterPin is used to get teh localnode's cluster status for a particular pin
func (api *API) getLocalStatusForClusterPin(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// initialize a connection to the cluster
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		api.LogError(err, IPFSClusterConnectionError)(c)
		return
	}
	// get the cluster status for the cid only asking the local cluster node
	status, err := manager.GetStatusForCidLocally(hash)
	if err != nil {
		api.LogError(err, IPFSClusterStatusError)(c)
		return
	}

	api.LogWithUser(username).Info("local cluster status for pin requested")

	Respond(c, http.StatusOK, gin.H{"response": status})
}

// GetGlobalStatusForClusterPin is used to get the global cluster status for a particular pin
func (api *API) getGlobalStatusForClusterPin(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// initialize a connection to the cluster
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		api.LogError(err, IPFSClusterConnectionError)(c)
		return
	}
	// get teh cluster wide status for this particular pin
	status, err := manager.GetStatusForCidGlobally(hash)
	if err != nil {
		api.LogError(err, IPFSClusterStatusError)(c)
		return
	}

	api.LogWithUser(username).Info("global cluster status for pin requested")

	Respond(c, http.StatusOK, gin.H{"response": status})
}

// FetchLocalClusterStatus is used to fetch the status of the localhost's cluster state, and not the rest of the cluster
func (api *API) fetchLocalClusterStatus(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	// this will hold all the retrieved content hashes
	var cids []*gocid.Cid
	// this will hold all the statuses of the content hashes
	var statuses []string
	// initialize a connection to the cluster
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		api.LogError(err, IPFSClusterConnectionError)(c)
		return
	}
	// fetch a map of all the statuses
	maps, err := manager.FetchLocalStatus()
	if err != nil {
		api.LogError(err, IPFSClusterStatusError)(c)
		return
	}
	// parse the maps
	for k, v := range maps {
		cids = append(cids, k)
		statuses = append(statuses, v)
	}

	api.LogWithUser(username).Info("local cluster state fetched")
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"cids": cids, "statuses": statuses}})
}
