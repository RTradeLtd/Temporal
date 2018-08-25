package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/gin-gonic/gin"
)

// ChangeAccountPassword is used to change a users password
func (api *API) changeAccountPassword(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)

	oldPassword, exists := c.GetPostForm("old_password")
	if !exists {
		FailNoExistPostForm(c, "old_password")
		return
	}

	newPassword, exists := c.GetPostForm("new_password")
	if !exists {
		FailNoExistPostForm(c, "new_password")
		return
	}

	um := models.NewUserManager(api.DBM.DB)

	suceeded, err := um.ChangePassword(ethAddress, oldPassword, newPassword)
	if err != nil {
		FailOnError(c, err)
		return
	}
	if !suceeded {
		FailOnError(c, errors.New("password change failed but no error occured"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "password changed",
	})
}

// RegisterUserAccount is used to sign up with temporal
func (api *API) registerUserAccount(c *gin.Context) {
	ethAddress := c.PostForm("eth_address")

	username, exists := c.GetPostForm("username")
	if !exists {
		FailNoExistPostForm(c, "username")
		return
	}
	password, exists := c.GetPostForm("password")
	if !exists {
		FailNoExistPostForm(c, "password")
		return
	}
	email, exists := c.GetPostForm("email_address")
	if !exists {
		FailNoExistPostForm(c, "email_address")
		return
	}

	userManager := models.NewUserManager(api.DBM.DB)
	userModel, err := userManager.NewUserAccount(ethAddress, username, password, email, false)
	if err != nil {
		FailOnError(c, err)
		return
	}
	userModel.HashedPassword = "scrubbed"
	c.JSON(http.StatusCreated, gin.H{"user": userModel})
	return
}

// CreateIPFSKey is used to create an IPFS key
func (api *API) createIPFSKey(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)

	keyType, exists := c.GetPostForm("key_type")
	if !exists {
		FailNoExistPostForm(c, "key_type")
		return
	}
	keyBits, exists := c.GetPostForm("key_bits")
	if !exists {
		FailNoExistPostForm(c, "key_bits")
		return
	}

	keyName, exists := c.GetPostForm("key_name")
	if !exists {
		FailNoExistPostForm(c, "key_name")
		return
	}

	bitsInt, err := strconv.Atoi(keyBits)
	if err != nil {
		FailOnError(c, err)
		return
	}

	key := queue.IPFSKeyCreation{
		UserName: username,
		Name:     keyName,
		Type:     keyType,
		Size:     bitsInt,
	}

	mqConnectionURL := api.TConfig.RabbitMQ.URL

	qm, err := queue.Initialize(queue.IpfsKeyCreationQueue, mqConnectionURL, true)
	if err != nil {
		FailOnError(c, err)
		return
	}

	err = qm.PublishMessageWithExchange(key, queue.IpfsKeyExchange)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "key creation sent to backend"})
}

// GetIPFSKeyNamesForAuthUser is used to get the keys a user has setup
func (api *API) getIPFSKeyNamesForAuthUser(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)

	um := models.NewUserManager(api.DBM.DB)
	keys, err := um.GetKeysForUser(ethAddress)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"key_names": keys["key_names"],
		"key_ids":   keys["key_ids"],
	})
}
