package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/queue"
	gocid "github.com/ipfs/go-cid"
	log "github.com/sirupsen/logrus"

	"github.com/RTradeLtd/Temporal/rtns/dlink"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/goamz/aws"
)

// PublishToIPNSDetails is used to publish a record on IPNS with more fine grained control over typical publishing methods
func (api *API) publishToIPNSDetails(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
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

	um := models.NewUserManager(api.dbm.DB)

	ownsKey, err := um.CheckIfKeyOwnedByUser(ethAddress, key)
	if err != nil {
		api.LogError(err, KeySearchError)(c)
		return
	}

	if !ownsKey {
		err = fmt.Errorf("user %s attempted to generate IPFS entry with unowned key", ethAddress)
		api.LogError(err, KeyUseError)(c)
		return
	}
	resolve, err := strconv.ParseBool(resolveString)
	if err != nil {
		Fail(c, err)
		return
	}
	lifetime, err := time.ParseDuration(lifetimeStr)
	if err != nil {
		Fail(c, err)
		return
	}
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		Fail(c, err)
		return
	}

	ie := queue.IPNSEntry{
		CID:         hash,
		LifeTime:    lifetime,
		TTL:         ttl,
		Resolve:     resolve,
		Key:         key,
		UserName:    ethAddress,
		NetworkName: "public",
	}

	fmt.Printf("IPNS Entry struct %+v\n", ie)

	qm, err := queue.Initialize(queue.IpnsEntryQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c)
		return
	}
	// in order to avoid generating too much IPFS dht traffic, we publish round-robin style
	// as we announce the records to the swarm, we will eventually achieve consistency across nodes automatically
	if err = qm.PublishMessage(ie); err != nil {
		api.LogError(err, QueuePublishError)(c)
		return
	}

	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("ipns entry creation request sent to backend")

	Respond(c, http.StatusOK, gin.H{"response": "ipns entry creation sent to backend"})
}

// GenerateDNSLinkEntry is used to generate a DNS link entry
func (api *API) generateDNSLinkEntry(c *gin.Context) {
	authUser := GetAuthenticatedUserFromContext(c)
	if authUser != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
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
		"user":    authUser,
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
