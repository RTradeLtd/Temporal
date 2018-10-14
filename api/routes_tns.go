package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RTradeLtd/Temporal/tns"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/gin-gonic/gin"
	peer "github.com/libp2p/go-libp2p-peer"
)

// PerformZoneRequest is used to perform a zone request lookup
func (api *API) performZoneRequest(c *gin.Context) {
	userToQuery, exists := c.GetPostForm("user_name")
	if !exists {
		FailWithMissingField(c, "user_name")
		return
	}
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
	peerID, exists := c.GetPostForm("peer_id")
	if !exists {
		FailWithMissingField(c, "peer_id")
		return
	}
	req := tns.ZoneRequest{
		UserName:           userToQuery,
		ZoneName:           zoneName,
		ZoneManagerKeyName: zoneManagerKeyName,
	}
	client, err := tns.GenerateTNSClient(true, nil)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	if err = client.MakeHost(client.PrivateKey, nil); err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	id, err := client.AddPeerToPeerStore(peerID)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	fmt.Println("querying tns")
	resp, err := client.QueryTNS(id, "zone-request", req)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	fmt.Println("tns queried")
	Respond(c, http.StatusOK, gin.H{"response": resp})
}

// AddRecordToZone is used to an a record to a TNS zone
func (api *API) addRecordToZone(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	zoneName, exists := c.GetPostForm("zone_name")
	if !exists {
		FailWithMissingField(c, "zone_name")
		return
	}
	recordName, exists := c.GetPostForm("record_name")
	if !exists {
		FailWithMissingField(c, "record_name")
		return
	}
	recordKeyName, exists := c.GetPostForm("record_key_name")
	if !exists {
		FailWithMissingField(c, "record_key_name")
		return
	}
	metadata, exists := c.GetPostForm("meta_data")
	var intf map[string]interface{}
	if exists {
		marshaled, err := json.Marshal(metadata)
		if err != nil {
			Fail(c, err, http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(marshaled, &intf); err != nil {
			Fail(c, err, http.StatusBadRequest)
			return
		}
	}
	req := queue.RecordCreation{
		ZoneName:      zoneName,
		RecordName:    recordName,
		RecordKeyName: recordKeyName,
		UserName:      username,
	}
	if len(intf) > 0 {
		req.MetaData = intf
	}
	mqURL := api.cfg.RabbitMQ.URL
	qm, err := queue.Initialize(queue.RecordCreationQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c, http.StatusBadRequest)
		return
	}
	if err = qm.PublishMessage(req); err != nil {
		api.LogError(err, QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "record creation request sent to backend"})
}

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
	queueManager, err := queue.Initialize(queue.ZoneCreationQueue, api.cfg.RabbitMQ.URL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c, http.StatusBadRequest)
		return
	}
	zoneCreation := queue.ZoneCreation{
		Name:           zone.Name,
		ManagerKeyName: zoneManagerKeyName,
		ZoneKeyName:    zoneKeyName,
		UserName:       username,
	}
	if err = queueManager.PublishMessage(zoneCreation); err != nil {
		api.LogError(err, QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": zone})
}
