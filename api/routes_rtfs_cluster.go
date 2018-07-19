package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

// PinHashToCluster is used to pin a hash to the local ipfs node
// TODO: INCOMPLETE
func PinHashToCluster(c *gin.Context) {
	contextCopy := c.Copy()
	hash := contextCopy.Param("hash")
	uploadAddress := GetAuthenticatedUserFromContext(contextCopy)

	holdTimeInMonths, exists := contextCopy.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}
	//TODO: CLEANUP AND MAKE MORE RESILIENT
	go func() {
		// currently after it is pinned, it is sent to the cluster to be pinned
		manager, err := rtfs_cluster.Initialize()
		if err != nil {
			fmt.Println(err)
			return
		}
		decodedHash, err := manager.DecodeHashString(hash)
		if err != nil {
			fmt.Println(err)
			return
		}
		// before exiting, it is pinned to the cluster
		err = manager.Pin(decodedHash)
		if err != nil {
			fmt.Println(err)
			// log error
		}
	}()
	// construct the rabbitmq message to add this entry to the database
	dpa := queue.DatabasePinAdd{
		Hash:             hash,
		UploaderAddress:  uploadAddress,
		HoldTimeInMonths: holdTimeInt,
	}
	// assert type assertion retrieving info from middleware
	mqConnectionURL := c.MustGet("mq_conn_url").(string)
	// initialize the queue
	qm, err := queue.Initialize(queue.DatabasePinAddQueue, mqConnectionURL)
	if err != nil {
		FailOnError(c, err)
		return
	}
	// publish the message, if there was an error finish processing
	err = qm.PublishMessage(dpa)
	if err != nil {
		FailOnError(c, err)
		return
	}
	qm.Close()
	c.JSON(http.StatusOK, gin.H{"upload": dpa})
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
	manager, err := rtfs_cluster.Initialize()
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
// TODO: fully implement, add in goroutines
func RemovePinFromCluster(c *gin.Context) {
	hash := c.Param("hash")
	manager, err := rtfs_cluster.Initialize()
	if err != nil {
		FailOnError(c, err)
		return
	}
	err = manager.RemovePinFromCluster(hash)
	if err != nil {
		FailOnError(c, err)
		return
	}
	mqConnectionURL := c.MustGet("mq_conn_url").(string)
	qm, err := queue.Initialize(queue.IpfsQueue, mqConnectionURL)
	if err != nil {
		FailOnError(c, err)
		return
	}
	qm.PublishMessage(hash)
	c.JSON(http.StatusOK, gin.H{"deleted": hash})
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
	manager, err := rtfs_cluster.Initialize()
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
	hash := c.Param("hash")
	// initialize a connection to the cluster
	manager, err := rtfs_cluster.Initialize()
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
	manager, err := rtfs_cluster.Initialize()
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
