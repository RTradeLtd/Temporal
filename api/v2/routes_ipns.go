package v2

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	path "github.com/ipfs/go-path"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	gocid "github.com/ipfs/go-cid"
)

// PublishToIPNSDetails is used to publish a record on IPNS with more fine grained control over typical publishing methods
func (api *API) publishToIPNSDetails(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract post forms
	forms, missingField := api.extractPostForms(c, "hash", "life_time", "ttl", "key", "resolve")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// validate that the hash is an ipfs one
	if _, err := gocid.Decode(forms["hash"]); err != nil {
		Fail(c, err)
		return
	}
	// ensure user owns the key
	if ownsKey, err := api.um.CheckIfKeyOwnedByUser(username, forms["key"]); err != nil {
		api.LogError(c, err, eh.KeySearchError)(http.StatusBadRequest)
		return
	} else if !ownsKey {
		err = fmt.Errorf("unauthorized access to key by user %s", username)
		api.LogError(c, err, eh.KeyUseError)(http.StatusBadRequest)
		return
	}
	// parse resolve into boolean type
	resolve, err := strconv.ParseBool(forms["resolve"])
	if err != nil {
		Fail(c, err)
		return
	}
	// parse lifetime into time.Duration
	lifetime, err := time.ParseDuration(forms["life_time"])
	if err != nil {
		Fail(c, err)
		return
	}
	// parse ttl into time.Duration
	ttl, err := time.ParseDuration(forms["ttl"])
	if err != nil {
		Fail(c, err)
		return
	}
	if err := api.usage.CanPublishIPNS(username); err != nil {
		api.LogError(c, err, "too many ipns records published this month, please wait until next billing cycle")(http.StatusBadRequest)
		return
	}
	if err := api.usage.IncrementIPNSUsage(username, 1); err != nil {
		api.LogError(c, err, "failed to increment ipns usage")
		return
	}
	// create ipns entry creation message
	ie := queue.IPNSEntry{
		CID:         forms["hash"],
		LifeTime:    lifetime,
		TTL:         ttl,
		Resolve:     resolve,
		Key:         forms["key"],
		UserName:    username,
		NetworkName: "public",
	}
	// send message for processing
	if err = api.queues.ipns.PublishMessage(ie); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.With("user", username).Info("ipns entry creation sent to backend")
	Respond(c, http.StatusOK, gin.H{"response": "ipns entry creation sent to backend"})
}

// getIPNSRecordsPublishedByUser is used to fetch IPNS records published by a user
func (api *API) getIPNSRecordsPublishedByUser(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// search for all records published by this user
	records, err := api.im.FindByUserName(username)
	if err != nil {
		api.LogError(c, err, eh.IpnsRecordSearchError)(http.StatusBadRequest)
		return
	}
	// if they haven't uploaded any records, don't fail just notify them
	if records == nil || len(*records) == 0 {
		Respond(c, http.StatusOK, gin.H{"response": "no ipns records found"})
		return
	}
	// return
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
	// extract post forms
	forms, missingField := api.extractPostForms(c, "hold_time", "ipns_path")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// validate the provided path is legit
	parsedPath := path.FromString(forms["ipns_path"])
	if err := parsedPath.IsValid(); err != nil {
		Fail(c, err, http.StatusBadRequest)
		return
	}
	// parse hold time
	holdTimeInt, err := api.validateHoldTime(username, forms["hold_time"])
	if err != nil {
		Fail(c, err)
		return
	}
	// resolve the ipns path to get the hash
	hashToPin, err := api.ipfs.Resolve(forms["ipns_path"])
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
	upload, err := api.upm.FindUploadByHashAndUserAndNetwork(username, hash, "public")
	// by this conditional if statement passing, it means the user has
	// upload content matching this hash before, and we don't want to charge them
	// so we should gracefully abort further processing
	if err == nil || upload != nil {
		Respond(c, http.StatusBadRequest, gin.H{"response": alreadyUploadedMessage})
		return
	}
	// get size of object
	stats, err := api.ipfs.Stat(hash)
	if err != nil {
		api.LogError(c, err, eh.IPFSObjectStatError)(http.StatusBadRequest)
		return
	}
	// ensure user can upload
	if err := api.usage.CanUpload(username, uint64(stats.CumulativeSize)); err != nil {
		api.LogError(c, err, eh.CantUploadError)(http.StatusBadRequest)
		return
	}
	// get the cost of this object
	cost, err := utils.CalculatePinCost(username, hash, holdTimeInt, api.ipfs, api.usage)
	if err != nil {
		api.LogError(c, err, eh.CostCalculationError)(http.StatusBadRequest)
		return
	}
	// ensure they have enough credits
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	if err := api.usage.UpdateDataUsage(username, uint64(stats.CumulativeSize)); err != nil {
		api.LogError(c, err, eh.DataUsageUpdateError)(http.StatusBadRequest)
		api.refundUserCredits(username, "pin", cost)
		return
	}
	// create pin message
	qp := queue.IPFSClusterPin{
		CID:              hash,
		NetworkName:      "public",
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       cost,
	}
	// send message for processing
	if err = api.queues.cluster.PublishMessage(qp); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		api.refundUserCredits(username, "pin", cost)
		api.usage.ReduceDataUsage(username, uint64(stats.CumulativeSize))
		return
	}
	// log and return
	api.l.Infow("ipfs pin request sent to backend", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": "pin request sent to backend"})
}
