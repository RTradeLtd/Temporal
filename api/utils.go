package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/c2h5oh/datasize"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var nilTime time.Time

const (
	// FilesUploadBucket is the bucket files are stored into before being processed
	FilesUploadBucket = "filesuploadbucket"
	// RtcCostUsd is the price of a single RTC in USD
	RtcCostUsd = 0.125
)

// CalculateFileSize helper route used to calculate the size of a file
func CalculateFileSize(c *gin.Context) {
	fileHandler, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}
	size := utils.BytesToGigaBytes(fileHandler.Size)
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"file_size_gb": size, "file_size_bytes": fileHandler.Size}})
}

// CheckAccessForPrivateNetwork checks if a user has access to a private network
func CheckAccessForPrivateNetwork(username, networkName string, db *gorm.DB) error {
	um := models.NewUserManager(db)
	canUpload, err := um.CheckIfUserHasAccessToNetwork(username, networkName)
	if err != nil {
		return err
	}

	if !canUpload {
		return errors.New("unauthorized access to private network")
	}
	return nil
}

// FileSizeCheck is used to check and validate the size of the uploaded file
func (api *API) FileSizeCheck(size int64) error {
	sizeInt, err := strconv.ParseInt(
		api.cfg.API.SizeLimitInGigaBytes,
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
		return "", nil
	case "ltc":
		return "", nil
	case "xmr":
		return "", nil
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
	availableCredits, err := api.um.GetCreditsForUser(username)
	if err != nil {
		return err
	}
	if availableCredits < cost {
		return errors.New(InvalidBalanceError)
	}
	if _, err := api.um.RemoveCredits(username, cost); err != nil {
		return err
	}
	return nil
}

// refundUserCredits is used to trigger a credit refund for a user, in the event of an API level processing failure.
// Note that we do not do any error handling here, instead we will log the information so that we may manually
// remediate the situation
func (api *API) refundUserCredits(username, callType string, cost float64) {
	if _, err := api.um.AddCredits(username, cost); err != nil {
		api.l.WithFields(log.Fields{
			"service":   api.service,
			"user":      username,
			"call_type": callType,
			"error":     err.Error(),
		}).Error(CreditRefundError)
	}
}

// validateAdminRequest is used to validate whether or not the requesting user is an administrator
func (api *API) validateAdminRequest(username string) error {
	isAdmin, err := api.um.CheckIfAdmin(username)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New(UnAuthorizedAdminAccess)
	}
	return nil
}
