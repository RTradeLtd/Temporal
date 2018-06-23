package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/rtns/dlink"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/goamz/aws"
)

func PublishToIPNSDetails(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	hash, present := c.GetPostForm("hash")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "hash post form not present",
		})
		return
	}
	lifetime, present := c.GetPostForm("life_time")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "life_time post form not present",
		})
		return
	}
	ttl, present := c.GetPostForm("ttl")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ttl post form not present",
		})
		return
	}
	key, present := c.GetPostForm("key")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "key post form not present",
		})
		return
	}
	resolveString, present := c.GetPostForm("resolve")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "resolve post form not present",
		})
		return
	}

	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error loading db middleware",
		})
		return
	}
	um := models.NewUserManager(db)

	ownsKey, err := um.CheckIfKeyOwnedByUser(ethAddress, key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !ownsKey {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "attempting to generate IPNS entry unowned key",
		})
		return
	}
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to establish connection with ipfs",
		})
		return
	}
	fmt.Println("creating key store manager")
	err = manager.CreateKeystoreManager()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}
	resolve, nil := strconv.ParseBool(resolveString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}
	fmt.Println("publishing to IPNS")
	resp, err := manager.PublishToIPNSDetails(hash, lifetime, ttl, key, resolve)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unable to create ipns record %s", err),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"name":  resp.Name,
		"value": resp.Value,
	})
}

// GenerateDNSLinkEntry is used to generate a DNS link entry
func GenerateDNSLinkEntry(c *gin.Context) {
	authUser := GetAuthenticatedUserFromContext(c)
	if authUser != AdminAddress {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unauthorized access",
		})
		return
	}

	recordName, exists := c.GetPostForm("record_name")
	if !exists {
		FailNoExist(c, "record_name post form does not exist")
		return
	}

	recordValue, exists := c.GetPostForm("record_value")
	if !exists {
		FailNoExist(c, "record_value post form does not exist")
		return
	}

	awsZone, exists := c.GetPostForm("aws_zone")
	if !exists {
		FailNoExist(c, "aws_zone post form does not exist")
		return
	}

	regionName, exists := c.GetPostForm("region_name")
	if !exists {
		FailNoExist(c, "region_name post form does not exist")
		return
	}

	aKey := c.MustGet("aws_key").(string)
	aSecret := c.MustGet("aws_secret").(string)

	var region aws.Region
	switch regionName {
	case "us-west-1":
		region = aws.USWest
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "region_name post form not valid",
		})
		return
	}

	awsManager, err := dlink.GenerateAwsLinkManager("get", aKey, aSecret, awsZone, region)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := awsManager.AddDNSLinkEntry(recordName, recordValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
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
