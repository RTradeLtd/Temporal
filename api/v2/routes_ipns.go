package v2

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	path "gx/ipfs/QmZErC2Ay6WuGi96CPg316PwitdwgLo6RxZRqVjJjRj2MR/go-path"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
	gocid "github.com/ipfs/go-cid"

	"github.com/gin-gonic/gin"
)

// PublishToIPNSDetails is used to publish a record on IPNS with more fine grained control over typical publishing methods
func (api *API) publishToIPNSDetails(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "hash", "life_time", "ttl", "key", "resolve")
	if len(forms) == 0 {
		return
	}
	if _, err := gocid.Decode(forms["hash"]); err != nil {
		Fail(c, err)
		return
	}
	cost, err := utils.CalculateAPICallCost("ipns", false)
	if err != nil {
		api.LogError(c, err, eh.CallCostCalculationError)(http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	ownsKey, err := api.um.CheckIfKeyOwnedByUser(username, forms["key"])
	if err != nil {
		api.LogError(c, err, eh.KeySearchError)(http.StatusBadRequest)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	if !ownsKey {
		err = fmt.Errorf("user %s attempted to generate IPFS entry with unowned key", username)
		api.LogError(c, err, eh.KeyUseError)(http.StatusBadRequest)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	resolve, err := strconv.ParseBool(forms["resolve"])
	if err != nil {
		Fail(c, err)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	lifetime, err := time.ParseDuration(forms["life_time"])
	if err != nil {
		Fail(c, err)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	ttl, err := time.ParseDuration(forms["ttl"])
	if err != nil {
		Fail(c, err)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	ie := queue.IPNSEntry{
		CID:         forms["hash"],
		LifeTime:    lifetime,
		TTL:         ttl,
		Resolve:     resolve,
		Key:         forms["key"],
		UserName:    username,
		NetworkName: "public",
		CreditCost:  cost,
	}
	if err = api.queues.ipns.PublishMessage(ie); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	api.l.With("user", username).Info("ipns entry creation sent to backend")
	Respond(c, http.StatusOK, gin.H{"response": "ipns entry creation sent to backend"})
}

// PublishDetailedIPNSToHostedIPFSNetwork is used to publish an IPNS record to a private network with fine grained control
func (api *API) publishDetailedIPNSToHostedIPFSNetwork(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "network_name", "hash", "life_time", "ttl", "key", "resolve")
	if len(forms) == 0 {
		return
	}
	cost, err := utils.CalculateAPICallCost("ipns", true)
	if err != nil {
		api.LogError(c, err, eh.CallCostCalculationError)(http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	if _, err := gocid.Decode(forms["hash"]); err != nil {
		Fail(c, err)
		return
	}
	ownsKey, err := api.um.CheckIfKeyOwnedByUser(username, forms["key"])
	if err != nil {
		api.LogError(c, err, eh.KeySearchError)(http.StatusBadRequest)
		return
	}
	if !ownsKey {
		err = fmt.Errorf("unauthorized access to key by user %s", username)
		api.LogError(c, err, eh.KeyUseError)(http.StatusBadRequest)
		return
	}
	resolve, err := strconv.ParseBool(forms["resolve"])
	if err != nil {
		// user error, dont log
		Fail(c, err)
		return
	}
	lifetime, err := time.ParseDuration(forms["life_time"])
	if err != nil {
		// user error, dont log
		Fail(c, err)
		return
	}
	ttl, err := time.ParseDuration(forms["ttl"])
	if err != nil {
		// user error, dont log
		Fail(c, err)
		return
	}
	ipnsUpdate := queue.IPNSEntry{
		CID:         forms["hash"],
		LifeTime:    lifetime,
		TTL:         ttl,
		Key:         forms["key"],
		Resolve:     resolve,
		NetworkName: forms["network_name"],
		UserName:    username,
		CreditCost:  cost,
	}
	if err = api.queues.ipns.PublishMessage(ipnsUpdate); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	api.l.Infow("private ipns entry creation request sent to backend", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": "ipns entry creation request sent to backend"})
}

// getIPNSRecordsPublishedByUser is used to fetch IPNS records published by a user
func (api *API) getIPNSRecordsPublishedByUser(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	records, err := api.im.FindByUserName(username)
	if err != nil {
		api.LogError(c, err, eh.IpnsRecordSearchError)(http.StatusBadRequest)
		return
	}
	// check if records is nil, or no entries. For len we must dereference first
	if records == nil || len(*records) == 0 {
		Respond(c, http.StatusOK, gin.H{"response": "no ipns records found"})
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": records})
}

// PinIPNSHash is used to pin the content referenced by an IPNS record
// only usable by public IPFS.
// The processing logic is as follows:
// 1) parse the path which will be /ipns/hash
// 2) validate that it is a valid path
// 3) resolve the cid referenced by the record
// 4) pop the last segment of the path, which will be the hash we are looking to pin
func (api *API) pinIPNSHash(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "hold_time", "ipns_path")
	if len(forms) == 0 {
		return
	}
	parsedPath := path.FromString(forms["ipns_path"])
	if err := parsedPath.IsValid(); err != nil {
		Fail(c, err, http.StatusBadRequest)
		return
	}
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// extract the hash to pin
	// as an unfortunate work-around we need to establish a seperate shell for this
	// TODO: come back to this and make it work without this janky work-around
	// this will likely need to wait for IPFS Cluster 0.8.0
	shell := ipfsapi.NewShell(api.cfg.IPFS.APIConnection.Host + ":" + api.cfg.IPFS.APIConnection.Port)
	if !shell.IsUp() {
		api.LogError(c, errors.New("node is not online"), eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	hashToPin, err := shell.Resolve(forms["ipns_path"])
	if err != nil {
		api.LogError(c, err, eh.IpnsRecordSearchError)(http.StatusBadRequest)
		return
	}
	// extract the hash to pin by using the cid at the very end of the path
	// if someone is passing in a multi-cid path like /ipfs/cidA/cidB/cidC
	// IPFS will pin cidC, although most of our service will recognize this as a valid hash
	// IPFS Cluster doesn't, so to keep consistency with the rest of our endpoints
	// we should only operate on a bare cidC
	hash := strings.Split(hashToPin, "/")[len(strings.Split(hashToPin, "/"))-1]
	cost, err := utils.CalculatePinCost(hash, holdTimeInt, api.ipfs, false)
	if err != nil {
		api.LogError(c, err, eh.PinCostCalculationError)(http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	ip := queue.IPFSPin{
		CID:              hash,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       cost,
	}
	if err = api.queues.pin.PublishMessageWithExchange(ip, queue.PinExchange); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		api.refundUserCredits(username, "pin", cost)
		return
	}
	api.l.Infow("ipfs pin request sent to backend", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": "pin request sent to backend"})
}
