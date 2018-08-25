package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
		msg := fmt.Sprintf("password change failed due to the following error: %s", err.Error())
		api.Logger.Error(msg)
		FailOnError(c, err)
		return
	}
	if !suceeded {
		msg := fmt.Sprintf("password changed failed for user %s to due an unspecified error", ethAddress)
		api.Logger.Error(msg)
		FailOnError(c, errors.New("password change failed but no error occured"))
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": api,
		"user":    ethAddress,
	}).Info("password changed")

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
		msg := fmt.Sprintf("user account registration failed for user %s due to the following error: %s", username, err.Error())
		api.Logger.Error(msg)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("account registered")

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

	switch keyType {
	case "rsa":
		break
	case "ed25519":
		break
	default:
		err := fmt.Errorf("%s is invalid key type must be rsa, or ed25519", keyType)
		FailOnError(c, err)
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

	qm, err := queue.Initialize(queue.IpfsKeyCreationQueue, mqConnectionURL, true, false)
	if err != nil {
		msg := fmt.Sprintf("queue initialization failed due to the following error: %s", err.Error())
		api.Logger.Error(msg)
		FailOnError(c, err)
		return
	}

	err = qm.PublishMessageWithExchange(key, queue.IpfsKeyExchange)
	if err != nil {
		msg := fmt.Sprintf(
			"publishing to queue %s with exchange %s failed due to the following error: %s",
			queue.IpfsKeyCreationQueue,
			queue.IpfsKeyExchange,
			err.Error())
		api.Logger.Error(msg)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("key creation request sent to backend")

	c.JSON(http.StatusOK, gin.H{"status": "key creation sent to backend"})
}

// GetIPFSKeyNamesForAuthUser is used to get the keys a user has setup
func (api *API) getIPFSKeyNamesForAuthUser(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)

	um := models.NewUserManager(api.DBM.DB)
	keys, err := um.GetKeysForUser(ethAddress)
	if err != nil {
		msg := fmt.Sprintf("key fetch for user %s failed due to error: %s", ethAddress, err.Error())
		api.Logger.Error(msg)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("key name list requested")

	c.JSON(http.StatusOK, gin.H{
		"key_names": keys["key_names"],
		"key_ids":   keys["key_ids"],
	})
}
