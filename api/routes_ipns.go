package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/queue"
	log "github.com/sirupsen/logrus"

	"github.com/RTradeLtd/Temporal/rtns/dlink"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/goamz/aws"
)

// PublishToIPNSDetails is used to publish a record on IPNS with more fine grained control
func (api *API) publishToIPNSDetails(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	hash, present := c.GetPostForm("hash")
	if !present {
		FailNoExistPostForm(c, "hash")
		return
	}
	lifetimeStr, present := c.GetPostForm("life_time")
	if !present {
		FailNoExistPostForm(c, "lifetime")
		return
	}
	ttlStr, present := c.GetPostForm("ttl")
	if !present {
		FailNoExistPostForm(c, "ttl")
		return
	}
	key, present := c.GetPostForm("key")
	if !present {
		FailNoExistPostForm(c, "key")
		return
	}
	resolveString, present := c.GetPostForm("resolve")
	if !present {
		FailNoExistPostForm(c, "resolve")
		return
	}

	mqURL := api.TConfig.RabbitMQ.URL

	um := models.NewUserManager(api.DBM.DB)

	ownsKey, err := um.CheckIfKeyOwnedByUser(ethAddress, key)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	if !ownsKey {
		api.Logger.Warnf("user %s attempted to generate IPFS entry with unowned key", ethAddress)
		FailOnError(c, errors.New("attempting to generate IPNS entry with unowned key"))
		return
	}
	resolve, err := strconv.ParseBool(resolveString)
	if err != nil {
		FailOnError(c, err)
		return
	}
	lifetime, err := time.ParseDuration(lifetimeStr)
	if err != nil {
		FailOnError(c, err)
		return
	}
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		FailOnError(c, err)
		return
	}

	ie := queue.IPNSEntry{
		CID:         hash,
		LifeTime:    lifetime,
		TTL:         ttl,
		Resolve:     resolve,
		Key:         key,
		EthAddress:  ethAddress,
		NetworkName: "public",
	}

	fmt.Printf("IPNS Entry struct %+v\n", ie)

	qm, err := queue.Initialize(queue.IpnsEntryQueue, mqURL, true)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}
	//TODO move to fanout exchange
	err = qm.PublishMessage(ie)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("ipns entry creation request sent to backend")

	c.JSON(http.StatusOK, gin.H{
		"status": "ipns entry creation sent to backend",
	})
}

// GenerateDNSLinkEntry is used to generate a DNS link entry
// TODO: turn into a queue call
func (api *API) generateDNSLinkEntry(c *gin.Context) {
	authUser := GetAuthenticatedUserFromContext(c)
	if authUser != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}

	recordName, exists := c.GetPostForm("record_name")
	if !exists {
		FailNoExistPostForm(c, "record_name")
		return
	}

	recordValue, exists := c.GetPostForm("record_value")
	if !exists {
		FailNoExistPostForm(c, "record_value")
		return
	}

	awsZone, exists := c.GetPostForm("aws_zone")
	if !exists {
		FailNoExistPostForm(c, "aws_zone")
		return
	}

	regionName, exists := c.GetPostForm("region_name")
	if !exists {
		FailNoExistPostForm(c, "region_name")
		return
	}

	aKey := api.TConfig.AWS.KeyID
	aSecret := api.TConfig.AWS.Secret

	var region aws.Region
	switch regionName {
	case "us-west-1":
		region = aws.USWest
	default:
		FailOnError(c, errors.New("invalid region_name"))
		return
	}

	awsManager, err := dlink.GenerateAwsLinkManager("get", aKey, aSecret, awsZone, region)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	resp, err := awsManager.AddDNSLinkEntry(recordName, recordValue)
	if err != nil {
		api.Logger.Error(err)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    authUser,
	}).Info("dnslink entry created")

	c.JSON(http.StatusOK, gin.H{
		"record_name":  recordName,
		"record_value": recordValue,
		"zone_name":    awsZone,
		"manager":      fmt.Sprintf("%+v", awsManager),
		"region":       aws.USWest.Name,
		"resp":         resp,
	})
}
