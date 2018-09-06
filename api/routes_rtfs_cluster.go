package api

import (
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
	log "github.com/sirupsen/logrus"
)

// PinHashToCluster is used to trigger a cluster pin of a particular CID
func (api *API) pinHashToCluster(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	hash := c.Param("hash")
	holdTime, exists := c.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}

	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}

	mqURL := api.TConfig.RabbitMQ.URL

	qm, err := queue.Initialize(queue.IpfsClusterPinQueue, mqURL, true, false)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	ipfsClusterPin := queue.IPFSClusterPin{
		CID:              hash,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
	}

	err = qm.PublishMessage(ipfsClusterPin)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("cluster pin request sent to backend")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": "cluster pin request sent to backend",
	})
}

// SyncClusterErrorsLocally is used to parse through the local cluster state and sync any errors that are detected.
func (api *API) syncClusterErrorsLocally(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	// initialize a conection to the cluster
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// parse the local cluster status, and sync any errors, retunring the cids that were in an error state
	syncedCids, err := manager.ParseLocalStatusAllAndSync()
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("local cluster errors parsed")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": syncedCids,
	})
}

// RemovePinFromCluster is used to remove a pin from the cluster global state
// this will mean that all nodes in the cluster will no longer track the pin
func (api *API) removePinFromCluster(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to cluster removal")
		return
	}
	hash := c.Param("hash")
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	err = manager.RemovePinFromCluster(hash)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("pin removal request sent to cluster")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": "pin removal request sent to cluster",
	})
}

// GetLocalStatusForClusterPin is used to get teh localnode's cluster status for a particular pin
func (api *API) getLocalStatusForClusterPin(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	hash := c.Param("hash")
	// initialize a connection to the cluster
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// get the cluster status for the cid only asking the local cluster node
	status, err := manager.GetStatusForCidLocally(hash)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("local cluster status for pin requested")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": status,
	})
}

// GetGlobalStatusForClusterPin is used to get the global cluster status for a particular pin
func (api *API) getGlobalStatusForClusterPin(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to cluster status")
		return
	}
	hash := c.Param("hash")
	// initialize a connection to the cluster
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// get teh cluster wide status for this particular pin
	status, err := manager.GetStatusForCidGlobally(hash)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("global cluster status for pin requested")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": status,
	})
}

// FetchLocalClusterStatus is used to fetch the status of the localhost's cluster state, and not the rest of the cluster
func (api *API) fetchLocalClusterStatus(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	// this will hold all the retrieved content hashes
	var cids []*gocid.Cid
	// this will hold all the statuses of the content hashes
	var statuses []string
	// initialize a connection to the cluster
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// fetch a map of all the statuses
	maps, err := manager.FetchLocalStatus()
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	// parse the maps
	for k, v := range maps {
		cids = append(cids, k)
		statuses = append(statuses, v)
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("local cluster state fetched")

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"response": gin.H{
			"cids":     cids,
			"statuses": statuses,
		},
	})
}
