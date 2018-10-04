package api

import (
	"fmt"
	"net/http"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/gin-gonic/gin"
)

// MakeBucket is used to create a bucket in our minio container
func (api *API) makeBucket(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	bucketName, exists := c.GetPostForm("bucket_name")
	if !exists {
		FailWithBadRequest(c, "bucket_name")
		return
	}

	var (
		accessKey = api.cfg.MINIO.AccessKey
		secretKey = api.cfg.MINIO.SecretKey
		endpoint  = fmt.Sprintf("%s:%s", api.cfg.MINIO.Connection.IP, api.cfg.MINIO.Connection.Port)
	)
	manager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, true)
	if err != nil {
		api.LogError(err, MinioConnectionError)(c)
		return
	}

	args := make(map[string]string)
	args["name"] = bucketName
	if err = manager.MakeBucket(args); err != nil {
		api.LogError(err, MinioBucketCreationError)(c)
		return
	}

	api.LogWithUser(username).Info("minio bucket created")

	Respond(c, http.StatusOK, gin.H{"response": "bucket created"})
}
