package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/rtns/dlink"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/goamz/aws"
)

func PublishToIPNSDetails(c *gin.Context) {
	_, exists := c.GetPostForm("network_name")
	if exists {
		PublishDetailedIPNSToHostedIPFSNetwork(c)
		return
	}
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

	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}
	mqURL, ok := c.MustGet("mq_conn_url").(string)
	if !ok {
		FailOnError(c, errors.New("failed to load rabbitmq"))
		return
	}

	um := models.NewUserManager(db)
	qm, err := queue.Initialize(queue.IpnsUpdateQueue, mqURL)
	if err != nil {
		FailOnError(c, err)
		return
	}

	ownsKey, err := um.CheckIfKeyOwnedByUser(ethAddress, key)
	if err != nil {
		FailOnError(c, err)
		return
	}

	if !ownsKey {
		FailOnError(c, errors.New("attempting to generate IPNS entry unowned key"))
		return
	}
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("creating key store manager")
	err = manager.CreateKeystoreManager()
	if err != nil {
		FailOnError(c, err)
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
	prePubTime := time.Now()
	keyID, err := um.GetKeyIDByName(ethAddress, key)
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println(key)
	fmt.Println(keyID)
	fmt.Println("publishing to IPNS")
	resp, err := manager.PublishToIPNSDetails(hash, key, lifetime, ttl, resolve)
	if err != nil {
		FailOnError(c, err)
		return
	}
	postPubTime := time.Now()
	timeDifference := postPubTime.Sub(prePubTime)

	im := models.NewIPNSManager(db)
	ipnsEntry, err := im.UpdateIPNSEntry(resp.Name, resp.Value, key, "public", lifetime, ttl)
	if err != nil {
		FailOnError(c, err)
		return
	}
	ipnsUpdate := queue.IPNSUpdate{
		CID:         hash,
		IPNSHash:    resp.Name,
		LifeTime:    lifetime.String(),
		TTL:         ttl.String(),
		Key:         key,
		Resolve:     resolve,
		EthAddress:  ethAddress,
		NetworkName: "public",
	}
	err = qm.PublishMessage(ipnsUpdate)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"name":                   resp.Name,
		"value":                  resp.Value,
		"time_to_create_minutes": timeDifference.Minutes(),
		"ipns_entry_model":       ipnsEntry,
	})
}

// GenerateDNSLinkEntry is used to generate a DNS link entry
func GenerateDNSLinkEntry(c *gin.Context) {
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

	aKey := c.MustGet("aws_key").(string)
	aSecret := c.MustGet("aws_secret").(string)

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
		FailOnError(c, err)
		return
	}

	resp, err := awsManager.AddDNSLinkEntry(recordName, recordValue)
	if err != nil {
		FailOnError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"record_name":  recordName,
		"record_value": recordValue,
		"zone_name":    awsZone,
		"manager":      fmt.Sprintf("%+v", awsManager),
		"region":       aws.USWest.Name,
		"resp":         resp,
	})
}
