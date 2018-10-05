package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	gocid "github.com/ipfs/go-cid"
	log "github.com/sirupsen/logrus"

	"github.com/RTradeLtd/Temporal/rtns/dlink"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/goamz/aws"
)

// PublishToIPNSDetails is used to publish a record on IPNS with more fine grained control over typical publishing methods
func (api *API) publishToIPNSDetails(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	hash, present := c.GetPostForm("hash")
	if !present {
		FailWithMissingField(c, "hash")
		return
	}
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	lifetimeStr, present := c.GetPostForm("life_time")
	if !present {
		FailWithMissingField(c, "lifetime")
		return
	}
	ttlStr, present := c.GetPostForm("ttl")
	if !present {
		FailWithMissingField(c, "ttl")
		return
	}
	key, present := c.GetPostForm("key")
	if !present {
		FailWithMissingField(c, "key")
		return
	}
	resolveString, present := c.GetPostForm("resolve")
	if !present {
		FailWithMissingField(c, "resolve")
		return
	}

	mqURL := api.cfg.RabbitMQ.URL

	cost, err := utils.CalculateAPICallCost("ipns", false)
	if err != nil {
		api.LogError(err, CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil {
		api.LogError(err, InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	ownsKey, err := api.um.CheckIfKeyOwnedByUser(username, key)
	if err != nil {
		api.LogError(err, KeySearchError)(c)
		api.refundUserCredits(username, "ipns", cost)
		return
	}

	if !ownsKey {
		err = fmt.Errorf("user %s attempted to generate IPFS entry with unowned key", username)
		api.LogError(err, KeyUseError)(c)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	resolve, err := strconv.ParseBool(resolveString)
	if err != nil {
		Fail(c, err)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	lifetime, err := time.ParseDuration(lifetimeStr)
	if err != nil {
		Fail(c, err)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		Fail(c, err)
		api.refundUserCredits(username, "ipns", cost)
		return
	}

	ie := queue.IPNSEntry{
		CID:         hash,
		LifeTime:    lifetime,
		TTL:         ttl,
		Resolve:     resolve,
		Key:         key,
		UserName:    username,
		NetworkName: "public",
		CreditCost:  cost,
	}

	qm, err := queue.Initialize(queue.IpnsEntryQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c)
		api.refundUserCredits(username, "ipns", cost)
		return
	}
	// in order to avoid generating too much IPFS dht traffic, we publish round-robin style
	// as we announce the records to the swarm, we will eventually achieve consistency across nodes automatically
	if err = qm.PublishMessage(ie); err != nil {
		api.LogError(err, QueuePublishError)(c)
		api.refundUserCredits(username, "ipns", cost)
		return
	}

	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("ipns entry creation request sent to backend")

	Respond(c, http.StatusOK, gin.H{"response": "ipns entry creation sent to backend"})
}

// GenerateDNSLinkEntry is used to generate a DNS link entry
func (api *API) generateDNSLinkEntry(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	recordName, exists := c.GetPostForm("record_name")
	if !exists {
		FailWithMissingField(c, "record_name")
		return
	}

	recordValue, exists := c.GetPostForm("record_value")
	if !exists {
		FailWithMissingField(c, "record_value")
		return
	}

	awsZone, exists := c.GetPostForm("aws_zone")
	if !exists {
		FailWithMissingField(c, "aws_zone")
		return
	}

	regionName, exists := c.GetPostForm("region_name")
	if !exists {
		FailWithMissingField(c, "region_name")
		return
	}

	aKey := api.cfg.AWS.KeyID
	aSecret := api.cfg.AWS.Secret

	var region aws.Region
	switch regionName {
	case "us-west-1":
		region = aws.USWest
	default:
		// user error, do not log
		Fail(c, errors.New("invalid region_name"))
		return
	}

	awsManager, err := dlink.GenerateAwsLinkManager("get", aKey, aSecret, awsZone, region)
	if err != nil {
		api.LogError(err, DNSLinkManagerError)
		Fail(c, err)
		return
	}

	resp, err := awsManager.AddDNSLinkEntry(recordName, recordValue)
	if err != nil {
		api.LogError(err, DNSLinkEntryError)
		Fail(c, err)
		return
	}

	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("dnslink entry created")

	Respond(c, http.StatusOK, gin.H{"response": gin.H{
		"record_name":  recordName,
		"record_value": recordValue,
		"zone_name":    awsZone,
		"manager":      fmt.Sprintf("%+v", awsManager),
		"region":       aws.USWest.Name,
		"resp":         resp,
	},
	})
}
