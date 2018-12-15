package v2

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	gocid "github.com/ipfs/go-cid"

	"github.com/gin-gonic/gin"
)

// PublishToIPNSDetails is used to publish a record on IPNS with more fine grained control over typical publishing methods
func (api *API) publishToIPNSDetails(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
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
		api.LogError(err, eh.CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	ownsKey, err := api.um.CheckIfKeyOwnedByUser(username, forms["key"])
	if err != nil {
		api.LogError(err, eh.KeySearchError)(c)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	if !ownsKey {
		err = fmt.Errorf("user %s attempted to generate IPFS entry with unowned key", username)
		api.LogError(err, eh.KeyUseError)(c)
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
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
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
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "network_name", "hash", "life_time", "ttl", "key", "resolve")
	if len(forms) == 0 {
		return
	}
	cost, err := utils.CalculateAPICallCost("ipns", true)
	if err != nil {
		api.LogError(err, eh.CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(err, eh.PrivateNetworkAccessError)(c)
		return
	}
	if _, err := gocid.Decode(forms["hash"]); err != nil {
		Fail(c, err)
		return
	}
	ownsKey, err := api.um.CheckIfKeyOwnedByUser(username, forms["key"])
	if err != nil {
		api.LogError(err, eh.KeySearchError)(c)
		return
	}
	if !ownsKey {
		err = fmt.Errorf("unauthorized access to key by user %s", username)
		api.LogError(err, eh.KeyUseError)(c)
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
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	api.l.Infow("private ipns entry creation request sent to backend", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": "ipns entry creation request sent to backend"})
}

// getIPNSRecordsPublishedByUser is used to fetch IPNS records published by a user
func (api *API) getIPNSRecordsPublishedByUser(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	records, err := api.im.FindByUserName(username)
	if err != nil {
		api.LogError(err, eh.IpnsRecordSearchError)(c, http.StatusBadRequest)
		return
	}
	// check if records is nil, or no entries. For len we must dereference first
	if records == nil || len(*records) == 0 {
		Respond(c, http.StatusOK, gin.H{"response": "no ipns records found"})
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": records})
}
