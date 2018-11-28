package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	gocid "github.com/ipfs/go-cid"
	log "github.com/sirupsen/logrus"

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
	mqURL := api.cfg.RabbitMQ.URL
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
	qm, err := queue.Initialize(queue.IpnsEntryQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	// in order to avoid generating too much IPFS dht traffic, we publish round-robin style
	// as we announce the records to the swarm, we will eventually achieve consistency across nodes automatically
	if err = qm.PublishMessage(ie); err != nil {
		api.LogError(err, eh.QueuePublishError)(c)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("ipns entry creation request sent to backend")
	Respond(c, http.StatusOK, gin.H{"response": "ipns entry creation sent to backend"})
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
