package v2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/crypto/v2"
	"github.com/RTradeLtd/database/v2/models"
	mnemonics "github.com/RTradeLtd/entropy-mnemonics"
	pb "github.com/RTradeLtd/grpc/krab"
	"github.com/RTradeLtd/rtfs/v2"
	"github.com/RTradeLtd/rtfs/v2/beam"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

// SystemsCheck is a basic check of system integrity
func (api *API) SystemsCheck(c *gin.Context) {
	hostname, err := os.Hostname()
	if err != nil {
		api.LogError(c, err, eh.HostNameNotFoundError)(http.StatusInternalServerError)
		return
	}
	c.Writer.Header().Set("X-API-Hostname", hostname)
	Respond(c, http.StatusOK, gin.H{
		"version":  api.version,
		"response": "systems online",
	})
}

// BeamContent is used to beam content from one network to another
func (api *API) beamContent(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract post forms
	forms, missingField := api.extractPostForms(c, "source_network", "destination_network", "content_hash")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	var (
		source, dest string
		net1Conn     *rtfs.IpfsManager
	)
	// validate the network to connect to
	if forms["source_network"] == "public" {
		source = api.cfg.IPFS.APIConnection.Host + ":" + api.cfg.IPFS.APIConnection.Port
		net1Conn, err = rtfs.NewManager(source, "", time.Minute*60)
	} else {
		// if non public network, validate user has access
		if err := CheckAccessForPrivateNetwork(username, forms["source_network"], api.dbm.DB); err != nil {
			api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
			return
		}
		// format a url to connect to
		source = api.GetIPFSEndpoint(forms["source_network"])
		// connect to the actual network
		net1Conn, err = rtfs.NewManager(source, GetAuthToken(c), time.Minute*60)
	}
	// ensure connection was successful
	if err != nil {
		api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
		return
	}
	// validate the network to connect to
	if forms["destination_network"] == "public" {
		dest = api.cfg.IPFS.APIConnection.Host + ":" + api.cfg.IPFS.APIConnection.Port
	} else {
		// non public network, validate user has access
		if err := CheckAccessForPrivateNetwork(username, forms["destination_network"], api.dbm.DB); err != nil {
			api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
			return
		}
		dest = api.GetIPFSEndpoint(forms["destination_network"])
	}
	// perform optional encryption of content
	if passphrase := c.PostForm("passphrase"); passphrase != "" {
		// encrypt the file file
		data, err := net1Conn.Cat(forms["content_hash"])
		if err != nil {
			api.LogError(c, err, eh.IPFSCatError)(http.StatusBadRequest)
			return
		}
		// re-add the encrypted content to the source network
		newCid, err := net1Conn.Add(bytes.NewReader(data))
		if err != nil {
			api.LogError(c, err, eh.IPFSAddError)(http.StatusBadRequest)
			return
		}
		// update the content hash to beam
		forms["content_hash"] = newCid
	}
	// create our dual network connection
	laserBeam, err := beam.NewLaser(source, dest, GetAuthToken(c))
	if err != nil {
		api.LogError(c, err, "failed to initialize laser beam")(http.StatusBadRequest)
		return
	}
	// initiate the content transfer
	if err := laserBeam.BeamFromSource(forms["content_hash"]); err != nil {
		api.LogError(c, err, "failed to tranfer content")(http.StatusBadRequest)
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"status": "content transferred", "content_hash": forms["content_hash"]}})
}

// ExportKey is used to export an ipfs key as a mnemonic phrase
func (api *API) exportKey(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get the key name
	keyName := c.Param("name")
	// validate user owns key name
	if owns, err := api.um.CheckIfKeyOwnedByUser(username, keyName); err != nil {
		api.LogError(c, err, eh.KeySearchError)(http.StatusBadRequest)
		return
	} else if !owns {
		api.LogError(c, errors.New(eh.KeyUseError), eh.KeyUseError)(http.StatusBadRequest)
		return
	}
	// get private key from krab keystore
	resp, err := api.keys.kb1.GetPrivateKey(context.Background(), &pb.KeyGet{Name: keyName})
	if err != nil {
		api.LogError(c, err, eh.KeyExportError)(http.StatusBadRequest)
		return
	}
	// convert private key to mnemonic phrase
	phrase, err := mnemonics.ToPhrase(resp.PrivateKey, mnemonics.English)
	if err != nil {
		api.LogError(c, err, eh.KeyExportError)(http.StatusBadRequest)
		return
	}
	// after successful parsing delete key from krab primary
	if resp, err := api.keys.kb1.DeletePrivateKey(context.Background(), &pb.KeyDelete{Name: keyName}); err != nil {
		api.l.Warnw("failed to delete key from primary krab", "error", err.Error())
	} else if resp.Status != "private key deleted" {
		api.l.Warnw("bad status from primary krab key delete", "status", resp.Status)
	}
	// only delete from secondary krab keystore if we arent in dev mode
	if !dev {
		// after successful parsing delete key from krab fallback
		if resp, err := api.keys.kb2.DeletePrivateKey(context.Background(), &pb.KeyDelete{Name: keyName}); err != nil {
			api.l.Warnw("failed to delete key from backup krab", "error", err.Error())
		} else if resp.Status != "private key deleted" {
			api.l.Warnw("bad status from backup krab key delete", "status", resp.Status)
		}
	}
	// get key id from database
	keyID, err := api.um.GetKeyIDByName(username, keyName)
	if err != nil {
		api.LogError(c, err, eh.KeySearchError)(http.StatusBadRequest)
		return
	}
	// remove key id from database
	if err := api.um.RemoveIPFSKeyForUser(username, keyName, keyID); err != nil {
		api.LogError(c, err, "failed to remove key from database")(http.StatusBadRequest)
		return
	}
	// decrease key count
	if err := api.usage.ReduceKeyCount(username, 1); err != nil {
		api.LogError(c, err, "failed to decrease key count")(http.StatusBadRequest)
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": phrase})
}

// downloadContentHash is used to download content from  a private ipfs network
func (api *API) downloadContentHash(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get the content hash that is to be downloaded
	contentHash := c.Param("hash")
	// ensure it's a valid content hash
	if _, err := gocid.Decode(contentHash); err != nil {
		Fail(c, err)
		return
	}
	// get the network name, default to public if not specified
	networkName := c.PostForm("network_name")
	var manager rtfs.Manager
	if networkName == "" {
		networkName = "public"
		manager = api.ipfs
	} else if networkName != "public" {
		// validate user access to network
		if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
			api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
			return
		}
		// retrieve api url
		apiURL := api.GetIPFSEndpoint(networkName)
		// initialize our connection to IPFS
		manager, err = rtfs.NewManager(apiURL, GetAuthToken(c), time.Minute*60)
		if err != nil {
			api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
			return
		}
	}
	// fetch the specified content type from the user
	contentType := c.PostForm("content_type")
	// if not specified, provide a default
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// get any extra headers the user might want
	exHeaders := c.PostFormArray("extra_headers")

	// read the contents of the file
	contents, err := manager.Cat(contentHash)
	if err != nil {
		api.LogError(c, err, eh.IPFSCatError)(http.StatusBadRequest)
		return
	}
	reader := bytes.NewReader(contents)
	size := len(contents)
	// decrypt Temporal-encrypted content if key is provided
	decryptKey := c.PostForm("decrypt_key")
	if decryptKey != "" {
		decrypted, err := crypto.NewEncryptManager(decryptKey).Decrypt(reader)
		if err != nil {
			Fail(c, err)
			return
		}
		size = len(decrypted)
		reader = bytes.NewReader(decrypted)
	}

	// parse extra headers if there are any
	extraHeaders := make(map[string]string)
	// only process if there is actual data to process
	if len(exHeaders) > 0 {
		// the array must be of equal length, as a header has two parts
		// the name of the header, and its value
		if len(exHeaders)%2 != 0 {
			FailWithMessage(c, "extra_headers post form is not even in length")
			return
		}
		// parse through the available headers
		for i := 0; i < len(exHeaders); i += 2 {
			if i+1 < len(exHeaders) {
				header := exHeaders[i]
				value := exHeaders[i+1]
				extraHeaders[header] = value
			}
		}
	}
	api.l.Infow("ipfs content download served", "user", username)
	c.DataFromReader(200, int64(size), contentType, reader, extraHeaders)
}

func (api *API) handleUserCreate(c *gin.Context, forms map[string]string, createErr error) {
	if createErr != nil {
		switch createErr.Error() {
		case eh.DuplicateEmailError:
			api.LogError(
				c,
				createErr,
				eh.DuplicateEmailError,
				"email",
				forms["email_address"])(http.StatusBadRequest)
			return
		case eh.DuplicateUserNameError:
			api.LogError(
				c,
				createErr,
				eh.DuplicateUserNameError,
				"username",
				forms["username"])(http.StatusBadRequest)
			return
		default:
			api.LogError(
				c,
				createErr,
				eh.UserAccountCreationError)(http.StatusBadRequest)
			return
		}
	}
	// generate a random token to validate email
	user, err := api.um.GenerateEmailVerificationToken(forms["username"])
	if err != nil {
		api.LogError(c, err, eh.EmailTokenGenerationError)(http.StatusBadRequest)
		return
	}
	// generate a jwt used to trigger email validation
	token, err := api.generateEmailJWTToken(user.UserName, user.EmailVerificationToken)
	if err != nil {
		api.LogError(c, err, "failed to generate email verification jwt")
		return
	}
	var url string
	// format the url the user clicks to activate email
	if dev {
		url = fmt.Sprintf(
			"https://dev.api.temporal.cloud/v2/account/email/verify/%s/%s",
			user.UserName, token,
		)
	} else {
		url = fmt.Sprintf(
			"https://api.temporal.cloud/v2/account/email/verify/%s/%s",
			user.UserName, token,
		)

	}
	// format a link tag
	link := fmt.Sprintf("<a href=\"%s\">link</a>", url)
	emailSubject := fmt.Sprintf(
		"%s Temporal Email Verification", forms["organization_name"],
	)
	// build email message
	es := queue.EmailSend{
		Subject: emailSubject,
		Content: fmt.Sprintf(
			"please click this %s to activate temporal email functionality", link,
		),
		ContentType: "text/html",
		UserNames:   []string{user.UserName},
		Emails:      []string{user.EmailAddress},
	}
	// send email message to queue for processing
	if err = api.queues.email.PublishMessage(es); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// remove hashed password from output
	user.HashedPassword = "scrubbed"
	// remove the verification token from output
	user.EmailVerificationToken = "scrubbed"
	// format a custom response that includes the user model
	// and an additional status field
	var status string
	if dev {
		status = fmt.Sprintf(
			"by continuing to use this service you agree to be bound by the following api terms and service %s",
			devTermsAndServiceURL,
		)
	} else {
		status = fmt.Sprintf(
			"by continuing to use this service you agree to be bound by the following api terms and service %s",
			prodTermsAndServiceURL,
		)
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": struct {
		*models.User
		Status string
	}{
		user, status,
	},
	})
}
