package v2

import (
	"bytes"
	"errors"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/crypto/v2"
	"github.com/RTradeLtd/database/v2/models"
	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
	"github.com/RTradeLtd/rtfs/v2"
	gocid "github.com/ipfs/go-cid"
	"github.com/jinzhu/gorm"

	"github.com/gin-gonic/gin"
)

// PinToHostedIPFSNetwork is used to pin content to a private ipfs network
func (api *API) pinToHostedIPFSNetwork(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get hash to pin
	hash := c.Param("hash")
	// ensure validity of hash
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// extract post forms
	forms, missingField := api.extractPostForms(c, "network_name", "hold_time")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// ensure user has access to network
	if err = CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// parse hold time
	holdTimeInt, err := api.validateHoldTime(username, forms["hold_time"])
	if err != nil {
		Fail(c, err)
		return
	}
	upload, err := api.upm.FindUploadByHashAndUserAndNetwork(username, hash, forms["network_name"])
	if err == nil || upload != nil {
		Respond(c, http.StatusBadRequest, gin.H{"response": alreadyUploadedMessage})
		return
	}
	// create pin message
	ip := queue.IPFSPin{
		CID:              hash,
		NetworkName:      forms["network_name"],
		UserName:         username,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       0,
		JWT:              GetAuthToken(c),
	}
	// send message for processing
	if err = api.queues.pin.PublishMessage(ip); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.With("user", username).Info("private network pin request sent to backend")
	Respond(c, http.StatusOK, gin.H{"response": "content pin request sent to backend"})
}

// AddFileToHostedIPFSNetwork is used to add a file to a private IPFS network via the simple method
func (api *API) addFileToHostedIPFSNetwork(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract post forms
	forms, missingField := api.extractPostForms(c, "network_name", "hold_time")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// verify user has access to private network
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)
		Fail(c, err)
		return
	}
	// parse hold time
	holdTimeInt, err := api.validateHoldTime(username, forms["hold_time"])
	if err != nil {
		Fail(c, err)
		return
	}
	// fetch the file, and create a handler to interact with it
	fileHandler, err := c.FormFile("file")
	if err != nil {
		// user error, do not log
		Fail(c, err)
		return
	}
	// validate file size
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		Fail(c, err)
		return
	}
	// open file into memory
	file, err := fileHandler.Open()
	if err != nil {
		api.LogError(c, err, eh.FileOpenError)(http.StatusBadRequest)
		return
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		Fail(c, err)
		return
	}
	hash, err := api.ipfs.Add(bytes.NewReader(fileBytes), ipfsapi.OnlyHash(true))
	if err != nil {
		api.LogError(c, err, eh.IPFSAddError)(http.StatusInternalServerError)
		return
	}
	upload, err := api.upm.FindUploadByHashAndUserAndNetwork(username, hash, forms["network_name"])
	if err == nil || upload != nil {
		Respond(c, http.StatusBadRequest, gin.H{"response": alreadyUploadedMessage})
		return
	}
	var reader io.Reader
	// encrypt file if passphrase is given
	if c.PostForm("passphrase") != "" {
		// html decode strings
		decodedPassPhrase := html.UnescapeString(c.PostForm("passphrase"))
		encrypted, err := crypto.NewEncryptManager(decodedPassPhrase).Encrypt(file)
		if err != nil {
			api.LogError(c, err, eh.EncryptionError)(http.StatusBadRequest)
			return
		}
		reader = bytes.NewReader(encrypted)
	} else {
		reader = bytes.NewReader(fileBytes)
	}
	// format a url to connect to for private network
	apiURL := api.GetIPFSEndpoint(forms["network_name"])
	// connect to private ifps network
	ipfsManager, err := rtfs.NewManager(apiURL, GetAuthToken(c), time.Minute*60)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	// add file to ipfs
	resp, err := ipfsManager.Add(reader)
	if err != nil {
		api.LogError(c, err, eh.IPFSAddError)(http.StatusBadRequest)
		return
	}
	// if this was an encrypted upload we need to update the encrypted upload table
	// ipfs cluster pin handles updating the regular uploads table
	if c.PostForm("passphrase") != "" {
		if _, err := api.ue.NewUpload(username, fileHandler.Filename, "public", resp); err != nil {
			api.LogError(c, err, eh.DatabaseUpdateError)(http.StatusBadRequest)
			return
		}
	}
	upload, err = api.upm.FindUploadByHashAndUserAndNetwork(
		username,
		resp,
		forms["network_name"],
	)
	if err != nil && err != gorm.ErrRecordNotFound {
		api.LogError(c, err, eh.UploadSearchError)(http.StatusBadRequest)
		return
	}
	if upload == nil {
		_, err = api.upm.NewUpload(resp, "file", models.UploadOptions{
			NetworkName:      forms["network_name"],
			Username:         username,
			HoldTimeInMonths: holdTimeInt,
		})
	} else {
		_, err = api.upm.UpdateUpload(holdTimeInt, username, resp, forms["network_name"])
	}
	if err != nil {
		api.LogError(c, err, eh.DatabaseUpdateError)(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.Infow("simple private ipfs file upload processed", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": resp})
}

// IpfsPubSubPublishToHostedIPFSNetwork is used to publish a pubsub message to a private ipfs network
func (api *API) ipfsPubSubPublishToHostedIPFSNetwork(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get the topic to publish a message too
	topic := c.Param("topic")
	// extract post forms
	forms, missingField := api.extractPostForms(c, "network_name", "message")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// validate access to private network
	if err := CheckAccessForPrivateNetwork(username, forms["network_name"], api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// format a url to connect too
	apiURL := api.GetIPFSEndpoint(forms["network_name"])
	// connect to private ipfs network
	manager, err := rtfs.NewManager(apiURL, GetAuthToken(c), time.Minute*60)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	// publish the actual message
	if err = manager.PubSubPublish(topic, forms["message"]); err != nil {
		api.LogError(c, err, eh.IPFSPubSubPublishError)(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.Infow("private ipfs pub sub message published", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"topic": topic, "message": forms["message"]}})
}

// GetObjectStatForIpfsForHostedIPFSNetwork is  used to get object stats from a private ipfs network
func (api *API) getObjectStatForIpfsForHostedIPFSNetwork(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get the hash to retrieve stats for
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	//get the network to connect to
	networkName := c.Param("networkName")
	// validate access to network
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// format a url to connect to
	apiURL := api.GetIPFSEndpoint(networkName)
	// connect to private ipfs network
	manager, err := rtfs.NewManager(apiURL, GetAuthToken(c), time.Minute*60)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	// get stats for object
	stats, err := manager.Stat(hash)
	if err != nil {
		api.LogError(c, err, eh.IPFSObjectStatError)(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.Infow("private ipfs object stat requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": stats})
}

// CheckLocalNodeForPinForHostedIPFSNetwork is used to check the serving node for a pin
func (api *API) checkLocalNodeForPinForHostedIPFSNetwork(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// hash to check pin for
	hash := c.Param("hash")
	// validate the type of hash
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// network to connect to
	networkName := c.Param("networkName")
	// validate access to network
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// format a url to connect to
	apiURL := api.GetIPFSEndpoint(networkName)
	// connect to the actual private network
	manager, err := rtfs.NewManager(apiURL, GetAuthToken(c), time.Minute*60)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	// check node for pin
	present, err := manager.CheckPin(hash)
	if err != nil {
		api.LogError(c, err, eh.IPFSPinParseError)(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.Infow("private ipfs pin check requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": present})
}

// GetDagObject is used to retrieve an IPLD object from ipfs
func (api *API) getDagObjectForHostedIPFSNetwork(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("private networks not supported in production, please use https://dev.api.temporal.cloud"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get the hash to retrieve dag object for
	hash := c.Param("hash")
	// validate type of hash
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// get network to connect to
	networkName := c.Param("networkName")
	// validate the user has access to the private network
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	// format a url to connect to
	apiURL := api.GetIPFSEndpoint(networkName)
	// connect to the private ipfs network
	im, err := rtfs.NewManager(apiURL, GetAuthToken(c), time.Minute*60)
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	// retrieve the dag object
	var out interface{}
	if err := im.DagGet(hash, &out); err != nil {
		api.LogError(c, err, eh.IPFSDagGetError)(http.StatusBadRequest)
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": out})
}
