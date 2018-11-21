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
	forms := api.extractPostForms(c, "user_name", "zone_name")
	if len(forms) == 0 {
		return
	}
	zone, err := api.zm.FindZoneByNameAndUser(forms["zone_name"], forms["user_name"])
	if err != nil {
		api.LogError(err, eh.ZoneSearchError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": zone})
}

// PerformRecordRequest is used to perform a request request lookup
func (api *API) performRecordRequest(c *gin.Context) {
	forms := api.extractPostForms(c, "user_name", "record_name")
	if len(forms) == 0 {
		return
	}
	record, err := api.rm.FindRecordByNameAndUser(forms["user_name"], forms["record_name"])
	if err != nil {
		api.LogError(err, eh.RecordSearchError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": record})
}

// AddRecordToZone is used to an a record to a TNS zone
func (api *API) addRecordToZone(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	forms := api.extractPostForms(c, "zone_name", "record_name", "record_key_name")
	if len(forms) == 0 {
		return
	}
	// slightly more complex form unmarshaling so this will be separated
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
		ZoneName:      forms["zone_name"],
		RecordName:    forms["record_name"],
		RecordKeyName: forms["record_key_name"],
		UserName:      username,
		MetaData:      intf,
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
	forms := api.extractPostForms(c, "zone_name", "zone_manager_key_name", "zone_key_name")
	if len(forms) == 0 {
		return
	}
	valid, err := api.um.CheckIfKeyOwnedByUser(username, forms["zone_mananger_key_name"])
	if err != nil {
		api.LogError(err, eh.KeySearchError)(c, http.StatusBadRequest)
		return
	}
	if !valid {
		api.LogError(err, eh.KeyUseError)(c, http.StatusBadRequest)
		return
	}
	valid, err = api.um.CheckIfKeyOwnedByUser(username, forms["zone_key_name"])
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
		forms["zone_name"],
		forms["zone_manager_key_name"],
		forms["zone_key_name"],
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
		ManagerKeyName: forms["zone_manager_key_name"],
		ZoneKeyName:    forms["zone_key_name"],
		UserName:       username,
	}
	if err = queueManager.PublishMessage(zoneCreation); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": zone})
}
