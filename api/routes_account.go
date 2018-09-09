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
		api.LogError(err, PasswordChangeError)
		FailOnError(c, err)
		return
	}
	if !suceeded {
		err = fmt.Errorf("password changed failed for user %s to due an unspecified error", username)
		api.LogError(err, PasswordChangeError)
		FailOnError(c, err)
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
		api.LogError(err, UserAccountCreationError)
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
		// user error, do not log
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
	um := models.NewUserManager(api.DBM.DB)
	keys, err := um.GetKeysForUser(username)
	if err != nil {
		api.LogError(err, KeySearchError)
		FailOnError(c, err)
		return
	}
	keyNamePrefixed := fmt.Sprintf("%s-%s", username, keyName)
	for _, v := range keys["key_names"] {
		if v == keyNamePrefixed {
			err = fmt.Errorf("key with name already exists")
			api.LogError(err, DuplicateKeyCreationError)
			FailOnError(c, err)
			return
		}
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
		api.LogError(err, QueueInitializationError)
		FailOnError(c, err)
		return
	}

	if err = qm.PublishMessageWithExchange(key, queue.IpfsKeyExchange); err != nil {
		api.LogError(err, QueuePublishError)
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
		api.LogError(err, KeySearchError)
		FailOnError(c, err)
		return
	}
	// if the user has no keys, fail with an error
	if len(keys["key_names"]) == 0 || len(keys["key_ids"]) == 0 {
		FailOnError(c, errors.New(NoKeyError))
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
		api.LogError(err, EthAddressChangeError)
		FailOnError(c, err)
		return
	}
	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("ethereum address changed")

	Respond(c, http.StatusOK, gin.H{"response": "address change successful"})
}
