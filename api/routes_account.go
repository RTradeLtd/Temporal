package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/rtfs"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	ci "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

var nilTime time.Time

func ChangeAccountPassword(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)

	oldPassword, exists := c.GetPostForm("old_password")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "old_password post form param does not exist",
		})
		return
	}

	newPassword, exists := c.GetPostForm("new_password")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "new_password post form param does not exist",
		})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	um := models.NewUserManager(db)

	suceeded, err := um.ChangePassword(ethAddress, oldPassword, newPassword)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("error occured while changing password %s", err.Error()),
		})
		return
	}
	if !suceeded {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "password change failed but no error occured",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "password changed",
	})
}

// RegisterUserAccount is used to sign up with temporal and gain web interface access
// you will not be granted API access however, as that needs to be done manually
func RegisterUserAccount(c *gin.Context) {
	ethAddress, exists := c.GetPostForm("eth_address")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eth_address parameter does not exist"})
		return
	}
	password, exists := c.GetPostForm("password")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password parameter does not exist"})
		return
	}
	db := c.MustGet("db").(*gorm.DB)

	userManager := models.NewUserManager(db)
	userModel, err := userManager.NewUserAccount(ethAddress, password, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userModel.HashedPassword = "scrubbed"
	c.JSON(http.StatusCreated, gin.H{"user": userModel})
	return
}

// RegisterEnterpriseUserAccount is used to register a user account marked as enterprise enabled
func RegisterEnterpriseUserAccount(c *gin.Context) {
	ethAddress, exists := c.GetPostForm("eth_address")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eth_address parameter does not exist"})
		return
	}
	password, exists := c.GetPostForm("password")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password parameter does not exist"})
		return
	}
	db := c.MustGet("db").(*gorm.DB)

	userManager := models.NewUserManager(db)
	userModel, err := userManager.NewUserAccount(ethAddress, password, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userModel.HashedPassword = "scrubbed"
	c.JSON(http.StatusCreated, gin.H{"user": userModel})
	return
}

// GetAuthenticatedUserFromContext is used to pull the eth address of hte user
func GetAuthenticatedUserFromContext(c *gin.Context) string {
	claims := jwt.ExtractClaims(c)
	// this is their eth address
	return claims["id"].(string)
}

// CreateIPFSKey is used to create an IPFS key
// TODO: encrypt key with provided password
func CreateIPFSKey(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)

	keyType, exists := c.GetPostForm("key_type")
	if !exists {
		FailNoExist(c, "key_type post form does not exist")
		return
	}
	keyBits, exists := c.GetPostForm("key_bits")
	if !exists {
		FailNoExist(c, "key_bits post form does not exist")
		return
	}

	keyName, exists := c.GetPostForm("key_name")
	if !exists {
		FailNoExist(c, "key_name post form does not exist")
		return
	}
	var keyTypeInt int
	var bitsInt int
	// currently we support generation of rsa or ed25519 keys
	switch keyType {
	case "rsa":
		keyTypeInt = ci.RSA
		// convert the string bits to int
		bitsInt64, err := strconv.ParseInt(keyBits, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		// right now we wont generate keys larger than 4096 in length
		if bitsInt64 > 4096 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "key bits must be 4096 or less. For larger keys contact your Temporal administrator",
			})
			return
		}
		bitsInt = int(bitsInt64)
	case "ed25519":
		// ed25519 uses a 256bit key length, we just specify the length here for brevity
		keyTypeInt = ci.Ed25519
		bitsInt = 256
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "key_type must be rsa or ed25519",
		})
		return
	}
	// initialize our connection to the ipfs node
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	//  load the key store manager
	err = manager.CreateKeystoreManager()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// prevent key name collision between different users
	//keyName = fmt.Sprintf("%s-%s", ethAddress, keyName)
	// create a key and save it to disk
	pk, err := manager.KeystoreManager.CreateAndSaveKey(keyName, keyTypeInt, bitsInt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// load the database so we can update our models
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error loading database middleware",
		})
		return
	}
	id, err := peer.IDFromPrivateKey(pk)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	um := models.NewUserManager(db)
	// update the user model with the new key
	err = um.AddIPFSKeyForUser(ethAddress, keyName, id.Pretty())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "key created",
		"id":     id.Pretty(),
	})
}

func GetIPFSKeyNamesForAuthUser(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)

	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error laoding database middleware",
		})
		return
	}

	um := models.NewUserManager(db)
	keys, err := um.GetKeysForUser(ethAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"keys": keys,
	})
}
