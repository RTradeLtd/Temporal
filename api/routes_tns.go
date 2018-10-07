package api

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/gin-gonic/gin"
	peer "github.com/libp2p/go-libp2p-peer"
)

// CreateZone is used to create a TNS zone
func (api *API) CreateZone(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	zoneName, exists := c.GetPostForm("zone_name")
	if !exists {
		FailWithMissingField(c, "zone_name")
		return
	}
	zoneManagerKeyName, exists := c.GetPostForm("zone_manager_key_name")
	if !exists {
		FailWithMissingField(c, "zone_manager_key_name")
		return
	}
	zoneKeyName, exists := c.GetPostForm("zone_key_name")
	if !exists {
		FailWithMissingField(c, "zone_key_name")
		return
	}
	rManager, err := rtfs.Initialize("", "")
	if err != nil {
		api.LogError(err, IPFSConnectionError)(c, http.StatusBadRequest)
		return
	}
	if err = rManager.CreateKeystoreManager(); err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	valid, err := api.um.CheckIfKeyOwnedByUser(username, zoneManagerKeyName)
	if err != nil {
		api.LogError(err, KeySearchError)(c, http.StatusBadRequest)
		return
	}
	if !valid {
		api.LogError(err, KeyUseError)(c, http.StatusBadRequest)
		return
	}
	valid, err = api.um.CheckIfKeyOwnedByUser(username, zoneKeyName)
	if err != nil {
		api.LogError(err, KeySearchError)(c, http.StatusBadRequest)
		return
	}
	if !valid {
		api.LogError(err, KeyUseError)(c, http.StatusBadRequest)
		return
	}
	zoneManagerPK, err := rManager.KeystoreManager.GetPrivateKeyByName(zoneManagerKeyName)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	zonePK, err := rManager.KeystoreManager.GetPrivateKeyByName(zoneKeyName)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	zonePublicKeyID, err := peer.IDFromPublicKey(zonePK.GetPublic())
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	zoneManagerPublicKeyID, err := peer.IDFromPublicKey(zoneManagerPK.GetPublic())
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	zone, err := api.zm.NewZone(
		username,
		zoneName,
		zoneManagerPublicKeyID.String(),
		zonePublicKeyID.String(),
		"qm..",
	)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": zone})
}
