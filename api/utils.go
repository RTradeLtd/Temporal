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
)

var nilTime time.Time

// FilesUploadBucket is the bucket files are stored into before being processed
const FilesUploadBucket = "filesuploadbucket"

// CalculateFileSize helper route used to calculate the size of a file
func CalculateFileSize(c *gin.Context) {
	fileHandler, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}
	size := utils.CalculateFileSizeInGigaBytes(fileHandler.Size)
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"file_size_gb": size, "file_size_bytes": fileHandler.Size}})
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
	}
	return "", errors.New(InvalidPaymentTypeError)

}
