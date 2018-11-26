package api

import (
	"bytes"
	"net/http"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/rtfs"
	"github.com/RTradeLtd/rtfs/beam"
	"github.com/gin-gonic/gin"
)

// SystemsCheck is a basic check of system integrity
func (api *API) SystemsCheck(c *gin.Context) {
	Respond(c, http.StatusOK, gin.H{"response": "systems online"})
}

// BeamContent is used to beam content from one network to another
func (api *API) beamContent(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
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
			api.LogError(err, eh.PrivateNetworkAccessError)(c, http.StatusBadRequest)
			return
		}
		url, err := api.nm.GetAPIURLByName(forms["source_network"])
		if err != nil {
			api.LogError(err, eh.PrivateNetworkAccessError)(c, http.StatusBadRequest)
			return
		}
		source = url
	}

	if forms["destination_network"] == "public" {
		dest = api.cfg.IPFS.APIConnection.Host + ":" + api.cfg.IPFS.APIConnection.Port
	} else {
		if err := CheckAccessForPrivateNetwork(username, forms["destination_network"], api.dbm.DB); err != nil {
			api.LogError(err, eh.PrivateNetworkAccessError)(c, http.StatusBadRequest)
			return
		}
		url, err := api.nm.GetAPIURLByName(forms["destination_network"])
		if err != nil {
			api.LogError(err, eh.PrivateNetworkAccessError)(c, http.StatusBadRequest)
			return
		}
		dest = url
	}
	if passphrase := c.PostForm("passphrase"); passphrase != "" {
		// connect to the source network
		net1Conn, err := rtfs.NewManager(source, nil, time.Minute*10)
		if err != nil {
			api.LogError(err, eh.IPFSConnectionError)(c, http.StatusBadRequest)
			return
		}
		// encrypt the file file
		data, err := net1Conn.Cat(forms["content_hash"])
		if err != nil {
			api.LogError(err, eh.IPFSCatError)(c, http.StatusBadRequest)
			return
		}
		// re-add the encrypted content to the source network
		newCid, err := net1Conn.Add(bytes.NewReader(data))
		if err != nil {
			api.LogError(err, eh.IPFSAddError)(c, http.StatusBadRequest)
			return
		}
		// update the content hash to beam
		forms["content_hash"] = newCid
	}
	// create our dual network connection
	laserBeam, err := beam.NewLaser(source, dest)
	if err != nil {
		api.LogError(err, "failed to initialize laser beam")(c, http.StatusBadRequest)
		return
	}
	// initiate the content transfer
	if err := laserBeam.BeamFromSource(forms["content_hash"]); err != nil {
		api.LogError(err, "failed to tranfer content")(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"status": "content transferred", "content_hash": forms["content_hash"]}})
}
