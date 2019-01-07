package v2

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	mnemonics "github.com/RTradeLtd/entropy-mnemonics"
	pb "github.com/RTradeLtd/grpc/krab"
	"github.com/RTradeLtd/rtfs"
	"github.com/RTradeLtd/rtfs/beam"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

// SystemsCheck is a basic check of system integrity
func (api *API) SystemsCheck(c *gin.Context) {
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
	forms := api.extractPostForms(c, "source_network", "destination_network", "content_hash")
	if len(forms) == 0 {
		return
	}
	var source, dest string

	if forms["source_network"] == "public" {
		source = api.cfg.IPFS.APIConnection.Host + ":" + api.cfg.IPFS.APIConnection.Port
	} else {
		if err := CheckAccessForPrivateNetwork(username, forms["source_network"], api.dbm.DB); err != nil {
			api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
			return
		}
		url, err := api.nm.GetAPIURLByName(forms["source_network"])
		if err != nil {
			api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
			return
		}
		source = url
	}

	if forms["destination_network"] == "public" {
		dest = api.cfg.IPFS.APIConnection.Host + ":" + api.cfg.IPFS.APIConnection.Port
	} else {
		if err := CheckAccessForPrivateNetwork(username, forms["destination_network"], api.dbm.DB); err != nil {
			api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
			return
		}
		url, err := api.nm.GetAPIURLByName(forms["destination_network"])
		if err != nil {
			api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
			return
		}
		dest = url
	}
	if passphrase := c.PostForm("passphrase"); passphrase != "" {
		// connect to the source network
		net1Conn, err := rtfs.NewManager(source, nil, time.Minute*10)
		if err != nil {
			api.LogError(c, err, eh.IPFSConnectionError)(http.StatusBadRequest)
			return
		}
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
	laserBeam, err := beam.NewLaser(source, dest)
	if err != nil {
		api.LogError(c, err, "failed to initialize laser beam")(http.StatusBadRequest)
		return
	}
	// initiate the content transfer
	if err := laserBeam.BeamFromSource(forms["content_hash"]); err != nil {
		api.LogError(c, err, "failed to tranfer content")(http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"status": "content transferred", "content_hash": forms["content_hash"]}})
}

// ExportKey is used to export an ipfs key as a mnemonic phrase
func (api *API) exportKey(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	keyName := c.Param("name")
	owns, err := api.um.CheckIfKeyOwnedByUser(username, keyName)
	if err != nil {
		api.LogError(c, err, eh.KeySearchError)(http.StatusBadRequest)
		return
	}
	if !owns {
		api.LogError(c, errors.New(eh.KeyUseError), eh.KeyUseError)(http.StatusBadRequest)
		return
	}

	resp, err := api.keys.GetPrivateKey(context.Background(), &pb.KeyGet{Name: keyName})
	if err != nil {
		api.LogError(c, err, eh.KeyExportError)(http.StatusBadRequest)
		return
	}
	phrase, err := mnemonics.ToPhrase(resp.PrivateKey, mnemonics.English)
	if err != nil {
		api.LogError(c, err, eh.KeyExportError)(http.StatusBadRequest)
		return
	}
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
		apiURL, err := api.nm.GetAPIURLByName(networkName)
		if err != nil {
			api.LogError(c, err, eh.APIURLCheckError)(http.StatusBadRequest)
			return
		}
		// initialize our connection to IPFS
		manager, err = rtfs.NewManager(apiURL, nil, time.Minute*10)
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
	// get the size of hte file in bytes
	stats, err := manager.Stat(contentHash)
	if err != nil {
		api.LogError(c, err, eh.IPFSObjectStatError)(http.StatusBadRequest)
		return
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

	api.l.Infow("private ipfs content download served", "user", username)
	c.DataFromReader(200, int64(stats.CumulativeSize), contentType, reader, extraHeaders)
}
