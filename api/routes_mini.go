package api

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/gin-gonic/gin"
)

// MakeBucket is used to create a bucket in our minio container
func MakeBucket(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access")
		return
	}
	credentials, ok := c.MustGet("minio_credentials").(map[string]string)
	if !ok {
		FailedToLoadMiddleware(c, "minio credentials")
		return
	}
	secure, ok := c.MustGet("minio_secure").(bool)
	if !ok {
		FailedToLoadMiddleware(c, "minio secure")
		return
	}
	endpoint, ok := c.MustGet("minio_endpoint").(string)
	if !ok {
		FailedToLoadMiddleware(c, "minio endpoint")
		return
	}
	manager, err := mini.NewMinioManager(endpoint, credentials["access_key"], credentials["secret_key"], secure)
	if err != nil {
		FailOnError(c, err)
		return
	}
	bucketName, exists := c.GetPostForm("bucket_name")
	if !exists {
		FailNoExistPostForm(c, "bucket_name")
		return
	}
	args := make(map[string]string)
	args["name"] = bucketName
	err = manager.MakeBucket(args)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status": "bucket created",
	})

}
