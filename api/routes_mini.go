package api

import (
	"fmt"
	"net/http"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/gin-gonic/gin"
)

// MakeBucket is used to create a bucket in our minio container
func (api *API) makeBucket(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access")
		return
	}
	accessKey := api.TConfig.MINIO.AccessKey
	secretKey := api.TConfig.MINIO.SecretKey
	endpoint := fmt.Sprintf("%s:%s", api.TConfig.MINIO.Connection.IP, api.TConfig.MINIO.Connection.Port)
	manager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, true)
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
