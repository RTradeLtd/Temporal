package api

import (
	"encoding/json"
	"net/http"

	"github.com/RTradeLtd/Temporal/eh"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/gin-gonic/gin"
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
	zone, err := api.zm.FindZoneByNameAndUser(zoneName, userToQuery)
	if err != nil {
		api.LogError(err, eh.ZoneSearchError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": zone})
}

// PerformRecordRequest is used to perform a request request lookup
func (api *API) performRecordRequest(c *gin.Context) {
	userToQuery, exists := c.GetPostForm("user_name")
	if !exists {
		FailWithMissingField(c, "user_name")
		return
	}
	recordName, exists := c.GetPostForm("record_name")
	if !exists {
		FailWithMissingField(c, "record_name")
		return
	}
	record, err := api.rm.FindRecordByNameAndUser(userToQuery, recordName)
	if err != nil {
		api.LogError(err, eh.RecordSearchError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": record})
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
		api.LogError(err, eh.QueueInitializationError)(c, http.StatusBadRequest)
		return
	}
	if err = qm.PublishMessage(req); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
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
	valid, err := api.um.CheckIfKeyOwnedByUser(username, zoneManagerKeyName)
	if err != nil {
		api.LogError(err, eh.KeySearchError)(c, http.StatusBadRequest)
		return
	}
	if !valid {
		api.LogError(err, eh.KeyUseError)(c, http.StatusBadRequest)
		return
	}
	valid, err = api.um.CheckIfKeyOwnedByUser(username, zoneKeyName)
	if err != nil {
		api.LogError(err, eh.KeySearchError)(c, http.StatusBadRequest)
		return
	}
	if !valid {
		api.LogError(err, eh.KeyUseError)(c, http.StatusBadRequest)
		return
	}
	zone, err := api.zm.NewZone(
		username,
		zoneName,
		zoneManagerKeyName,
		zoneKeyName,
		"qm..",
	)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	queueManager, err := queue.Initialize(queue.ZoneCreationQueue, api.cfg.RabbitMQ.URL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c, http.StatusBadRequest)
		return
	}
	zoneCreation := queue.ZoneCreation{
		Name:           zone.Name,
		ManagerKeyName: zoneManagerKeyName,
		ZoneKeyName:    zoneKeyName,
		UserName:       username,
	}
	if err = queueManager.PublishMessage(zoneCreation); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": zone})
}
