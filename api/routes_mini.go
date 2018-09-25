package api

import (
	"fmt"
	"net/http"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// MakeBucket is used to create a bucket in our minio container
func (api *API) makeBucket(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access")
		return
	}
	bucketName, exists := c.GetPostForm("bucket_name")
	if !exists {
		FailNoExistPostForm(c, "bucket_name")
		return
	}
	accessKey := api.cfg.MINIO.AccessKey
	secretKey := api.cfg.MINIO.SecretKey
	endpoint := fmt.Sprintf("%s:%s", api.cfg.MINIO.Connection.IP, api.cfg.MINIO.Connection.Port)
	manager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, true)
	if err != nil {
		api.LogError(err, MinioConnectionError)
		FailOnError(c, err)
		return
	}

	args := make(map[string]string)
	args["name"] = bucketName
	if err = manager.MakeBucket(args); err != nil {
		api.LogError(err, MinioBucketCreationError)
		FailOnError(c, err)
		return
	}

	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    ethAddress,
	}).Info("minio bucket created")

	Respond(c, http.StatusOK, gin.H{"response": "bucket created"})

}
