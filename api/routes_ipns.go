package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/gin-gonic/gin"
)

// PublishToIPNS is used to publish a record to ipns
// TODO: make sure the user owns the hash in question
func PublishToIPNS(c *gin.Context) {
	authUser := GetAuthenticatedUserFromContext(c)
	if authUser != AdminAddress {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "unauthorized access",
		})
		return
	}
	hash := c.Param("hash")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}
	fmt.Println("publishing to ipns")
	resp, err := manager.PublishToIPNS(hash)
	if err != nil {
		fmt.Println("error publishing to ipns", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}
	fmt.Println("published to ipns")
	c.JSON(http.StatusOK, gin.H{
		"status": "published",
		"name":   resp.Name,
		"value":  resp.Value,
	})
}

func PublishToIPNSDetails(c *gin.Context) {
	authUser := GetAuthenticatedUserFromContext(c)
	if authUser != AdminAddress {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "unauthorized access",
		})
		return
	}
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
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to establish connection with ipfs",
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

	c.JSON(http.StatusOK, gin.H{
		"record_name":  recordName,
		"record_value": recordValue,
	})
}
