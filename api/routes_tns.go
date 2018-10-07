package api

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/gin-gonic/gin"
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
	fmt.Println(1)
	valid, err := api.um.CheckIfKeyOwnedByUser(zoneManagerKeyName, username)
	if err != nil {
		api.LogError(err, KeySearchError)(c, http.StatusBadRequest)
		return
	}
	if !valid {
		api.LogError(err, KeyUseError)(c, http.StatusBadRequest)
		return
	}
	fmt.Println(2)
	valid, err = api.um.CheckIfKeyOwnedByUser(zoneKeyName, username)
	if err != nil {
		api.LogError(err, KeySearchError)(c, http.StatusBadRequest)
		return
	}
	if !valid {
		api.LogError(err, KeyUseError)(c, http.StatusBadRequest)
		return
	}
	fmt.Println(3)
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
	zonePublicKeyBytes, err := zonePK.GetPublic().Bytes()
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	zoneManagerPublicKeyBytes, err := zoneManagerPK.GetPublic().Bytes()
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	zone, err := api.zm.NewZone(
		username,
		zoneName,
		hex.EncodeToString(zoneManagerPublicKeyBytes),
		hex.EncodeToString(zonePublicKeyBytes),
		"",
	)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": zone})
}
