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
	username := GetAuthenticatedUserFromContext(c)

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

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("password change requested")

	um := models.NewUserManager(api.DBM.DB)
	suceeded, err := um.ChangePassword(username, oldPassword, newPassword)
	if err != nil {
		api.Logger.WithFields(log.Fields{
			"service": "api",
			"user":    username,
			"error":   err.Error(),
		}).Warn("password change failed")
		FailOnError(c, err)
		return
	}
	if !suceeded {
		msg := fmt.Sprintf("password changed failed for user %s to due an unspecified error", username)
		api.Logger.Error(msg)
		FailOnError(c, errors.New("password change failed but no error occured"))
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": api,
		"user":    username,
	}).Info("password changed")

	Respond(c, http.StatusOK, gin.H{"response": "password changed"})
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
	if ethAddress == "" {
		ethAddress = username
	}
	api.Logger.WithFields(log.Fields{
		"service": "api",
	}).Info("user account registration detected")

	userManager := models.NewUserManager(api.DBM.DB)
	userModel, err := userManager.NewUserAccount(ethAddress, username, password, email, false)
	if err != nil {
		api.Logger.WithFields(log.Fields{
			"service": "api",
			"error":   err.Error(),
			"user":    username,
		}).Error("user account registration failed")
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("user account registered")

	userModel.HashedPassword = "scrubbed"
	Respond(c, http.StatusOK, gin.H{"response": userModel})
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
		UserName:    username,
		Name:        keyName,
		Type:        keyType,
		Size:        bitsInt,
		NetworkName: "public",
	}

	mqConnectionURL := api.TConfig.RabbitMQ.URL

	qm, err := queue.Initialize(queue.IpfsKeyCreationQueue, mqConnectionURL, true, false)
	if err != nil {
		api.Logger.WithFields(log.Fields{
			"service": "api",
			"user":    username,
			"error":   err.Error(),
		}).Error("failed to initialize queue")
		FailOnError(c, err)
		return
	}

	err = qm.PublishMessageWithExchange(key, queue.IpfsKeyExchange)
	if err != nil {
		api.Logger.WithFields(log.Fields{
			"service": "api",
			"user":    username,
			"error":   err.Error(),
		}).Error("failed to publish message")
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("key creation request sent to backend")

	Respond(c, http.StatusOK, gin.H{"response": "key creation sent to backend"})
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

	Respond(c, http.StatusOK, gin.H{"response": gin.H{"key_names": keys["key_names"], "key_ids": keys["key_ids"]}})
}

// ChangeEthereumAddress is used to change a user's ethereum address
func (api *API) changeEthereumAddress(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	ethAddress, exists := c.GetPostForm("eth_address")
	if !exists {
		FailNoExistPostForm(c, "eth_address")
		return
	}
	um := models.NewUserManager(api.DBM.DB)
	if _, err := um.ChangeEthereumAddress(username, ethAddress); err != nil {
		api.Logger.WithFields(log.Fields{
			"service": "api",
			"user":    username,
			"error":   err.Error(),
		}).Info("ethereum address change failed")
		FailOnError(c, err)
		return
	}
	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("ethereum address changed")

	Respond(c, http.StatusOK, gin.H{"response": "address change successful"})
}
