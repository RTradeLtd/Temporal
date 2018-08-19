package api

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

// PinHashToCluster is used to trigger a cluster pin of a particular CID
func PinHashToCluster(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to cluster pin")
	}
	hash := c.Param("hash")

	mqURL, ok := c.MustGet("mq_conn_url").(string)
	if !ok {
		FailedToLoadMiddleware(c, "rabbitmq")
		return
	}
	qm, err := queue.Initialize(queue.IpfsClusterPinQueue, mqURL)
	if err != nil {
		FailOnError(c, err)
		return
	}

	ipfsClusterPin := queue.IPFSClusterPin{
		CID:         hash,
		NetworkName: "public",
	}

	err = qm.PublishMessage(ipfsClusterPin)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "content hash pin request sent to cluster"})
}

// SyncClusterErrorsLocally is used to parse through the local cluster state
// and sync any errors that are detected.
func SyncClusterErrorsLocally(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	// initialize a conection to the cluster
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	// parse the local cluster status, and sync any errors, retunring the cids that were in an error state
	syncedCids, err := manager.ParseLocalStatusAllAndSync()
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"synced-cids": syncedCids})
}

// RemovePinFromCluster is used to remove a pin from the cluster global state
// this will mean that all nodes in the cluster will no longer track the pin
// TODO: change to use a queue
func RemovePinFromCluster(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to cluster removal")
		return
	}
	hash := c.Param("hash")
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	err = manager.RemovePinFromCluster(hash)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"statsu": "pin removal request sent to cluster"})
}

// GetLocalStatusForClusterPin is used to get teh localnode's cluster status for a particular pin
func GetLocalStatusForClusterPin(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	hash := c.Param("hash")
	// initialize a connection to the cluster
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	// get the cluster status for the cid only asking the local cluster node
	status, err := manager.GetStatusForCidLocally(hash)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusFound, gin.H{"status": status})
}

// GetGlobalStatusForClusterPin is used to get the global cluster status for a particular pin
func GetGlobalStatusForClusterPin(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to cluster status")
		return
	}
	hash := c.Param("hash")
	// initialize a connection to the cluster
	manager, err := rtfs_cluster.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	// get teh cluster wide status for this particular pin
	status, err := manager.GetStatusForCidGlobally(hash)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusFound, gin.H{"status": status})
}

// FetchLocalClusterStatus is used to fetch the status of the localhost's
// cluster state, and not the rest of the cluster
// TODO: cleanup
func FetchLocalClusterStatus(c *gin.Context) {
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
		FailOnError(c, err)
		return
	}
	// fetch a map of all the statuses
	maps, err := manager.FetchLocalStatus()
	if err != nil {
		FailOnError(c, err)
		return
	}
	// parse the maps
	for k, v := range maps {
		cids = append(cids, k)
		statuses = append(statuses, v)
	}
	c.JSON(http.StatusOK, gin.H{"cids": cids, "statuses": statuses})
}
