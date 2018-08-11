package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	ci "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

var nilTime time.Time

// ChangeAccountPassword is used to change a users password
func ChangeAccountPassword(c *gin.Context) {
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

	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}
	um := models.NewUserManager(db)

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
func RegisterUserAccount(c *gin.Context) {
	ethAddress, exists := c.GetPostForm("eth_address")
	if !exists {
		FailNoExistPostForm(c, "eth_address")
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
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}

	userManager := models.NewUserManager(db)
	userModel, err := userManager.NewUserAccount(ethAddress, password, email, false)
	if err != nil {
		FailOnError(c, err)
		return
	}
	userModel.HashedPassword = "scrubbed"
	c.JSON(http.StatusCreated, gin.H{"user": userModel})
	return
}

// CreateIPFSKey is used to create an IPFS key
func CreateIPFSKey(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)

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
	var keyTypeInt int
	var bitsInt int
	// currently we support generation of rsa or ed25519 keys
	switch keyType {
	case "rsa":
		keyTypeInt = ci.RSA
		// convert the string bits to int
		bitsInt64, err := strconv.ParseInt(keyBits, 10, 64)
		if err != nil {
			FailOnError(c, err)
			return
		}
		// right now we wont generate keys larger than 4096 in length
		if bitsInt64 > 4096 {
			FailOnError(c, errors.New("key bits must be 4096 or less. For larger keys contact your Temporal administrator"))
			return
		}
		bitsInt = int(bitsInt64)
	case "ed25519":
		// ed25519 uses a 256bit key length, we just specify the length here for brevity
		keyTypeInt = ci.Ed25519
		bitsInt = 256
	default:
		FailOnError(c, errors.New("key_type must be rsa or ed25519"))
		return
	}
	// initialize our connection to the ipfs node
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	//  load the key store manager
	err = manager.CreateKeystoreManager()
	if err != nil {
		FailOnError(c, err)
		return
	}
	// prevent key name collision between different users
	//keyName = fmt.Sprintf("%s-%s", ethAddress, keyName)
	// create a key and save it to disk
	pk, err := manager.KeystoreManager.CreateAndSaveKey(keyName, keyTypeInt, bitsInt)
	if err != nil {
		FailOnError(c, err)
		return
	}
	// load the database so we can update our models
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}
	id, err := peer.IDFromPrivateKey(pk)
	if err != nil {
		FailOnError(c, err)
		return
	}
	um := models.NewUserManager(db)
	// update the user model with the new key
	err = um.AddIPFSKeyForUser(ethAddress, keyName, id.Pretty())
	if err != nil {
		FailOnError(c, err)
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
		FailedToLoadDatabase(c)
		return
	}

	um := models.NewUserManager(db)
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
