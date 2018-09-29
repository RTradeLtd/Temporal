package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/utils"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/c2h5oh/datasize"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var nilTime time.Time

// FilesUploadBucket is the bucket files are stored into before being processed
const FilesUploadBucket = "filesuploadbucket"

// CalculateFileSize helper route used to calculate the size of a file
func CalculateFileSize(c *gin.Context) {
	fileHandler, err := c.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}
	size := utils.CalculateFileSizeInGigaBytes(fileHandler.Size)
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"file_size_gb": size, "file_size_bytes": fileHandler.Size}})
}

// FailNoExistPostForm is a failure used when a post form does not exist
func FailNoExistPostForm(c *gin.Context, formName string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":     http.StatusBadRequest,
		"response": fmt.Sprintf("%s post form not present", formName),
	})
}

// FailNotAuthorized is a failure used when a user is unauthorized for an action
func FailNotAuthorized(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"code":     http.StatusForbidden,
		"response": message,
	})
}

// FailOnError is a failure used when an error occurs
func FailOnError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":     http.StatusBadRequest,
		"response": err.Error(),
	})
}

// FailOnServerError is an error handler used when a server error occurs
func FailOnServerError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":     http.StatusInternalServerError,
		"response": err.Error(),
	})
}

// CheckAccessForPrivateNetwork checks if a user has access to a private network
func CheckAccessForPrivateNetwork(ethAddress, networkName string, db *gorm.DB) error {
	um := models.NewUserManager(db)
	canUpload, err := um.CheckIfUserHasAccessToNetwork(ethAddress, networkName)
	if err != nil {
		return err
	}

	if !canUpload {
		return errors.New("unauthorized access to private network")
	}
	return nil
}

// GetAuthenticatedUserFromContext is used to pull the eth address of hte user
func GetAuthenticatedUserFromContext(c *gin.Context) string {
	claims := jwt.ExtractClaims(c)
	// this is their eth address
	return claims["id"].(string)
}

// Respond is a wrapper used to handle API responses
func Respond(c *gin.Context, status int, body gin.H) {
	body["code"] = status
	c.JSON(status, body)
}

// LogError is a wrapper used by the API to handle logging of errors
func (api *API) LogError(err error, message string) {
	api.Logger.WithFields(log.Fields{
		"service": api.Service,
		"error":   err.Error(),
	}).Error(message)
}

// FileSizeCheck is used to check and validate the size of the uploaded file
func (api *API) FileSizeCheck(size int64) error {
	sizeInt, err := strconv.ParseInt(
		api.TConfig.API.SizeLimitInGigaBytes,
		10,
		64,
	)
	if err != nil {
		return err
	}
	gbInt := int64(datasize.GB.Bytes()) * sizeInt
	if size > gbInt {
		return errors.New(FileTooBigError)
	}
	return nil
}

func (api *API) getDepositAddress(paymentType string) (string, error) {
	switch paymentType {
	case "eth", "rtc":
		return "0xc7459562777DDf3A1A7afefBE515E8479Bd3FDBD", nil
	case "btc":
		return "0", nil
	case "ltc":
		return "0", nil
	case "xmr":
		return "0", nil
	}
	return "", errors.New(InvalidPaymentTypeError)
}

func (api *API) validateBlockchain(blockchain string) bool {
	switch blockchain {
	case "ethereum", "bitcoin", "litecoin", "monero":
		return true
	}
	return false
}

// validateUserCredits is used to validate whether or not a user has enough credits to pay for an action
func (api *API) validateUserCredits(username string, cost float64) error {
	um := models.NewUserManager(api.DBM.DB)
	availableCredits, err := um.GetCreditsForUser(username)
	if err != nil {
		return err
	}
	if availableCredits < cost {
		return errors.New(InvalidBalanceError)
	}
	return nil
}
