package v2

import (
	"fmt"
	"net/http"
	"os"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/mini"
	"github.com/gin-gonic/gin"
)

// MakeBucket is used to create a bucket in our minio container
func (api *API) makeBucket(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, eh.UnAuthorizedAdminAccess)
		return
	}
	forms := api.extractPostForms(c, "bucket_name")
	if len(forms) == 0 {
		return
	}
	var (
		accessKey = api.cfg.MINIO.AccessKey
		secretKey = api.cfg.MINIO.SecretKey
		endpoint  = fmt.Sprintf("%s:%s", api.cfg.MINIO.Connection.IP, api.cfg.MINIO.Connection.Port)
	)
	manager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, os.Getenv("MINI_SSL_ENABLE") == "true")
	if err != nil {
		api.LogError(err, eh.MinioConnectionError)(c)
		return
	}

	args := make(map[string]string)
	args["name"] = forms["bucket_name"]
	if err = manager.MakeBucket(args); err != nil {
		api.LogError(err, eh.MinioBucketCreationError)(c)
		return
	}
	api.l.Infow("minio bucket created", "user", username)

	Respond(c, http.StatusOK, gin.H{"response": "bucket created"})
}
