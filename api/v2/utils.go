package v2

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/gorm"
	"github.com/c2h5oh/datasize"
	"github.com/gin-gonic/gin"
	jwt "gopkg.in/dgrijalva/jwt-go.v3"
)

var nilTime time.Time

const (
	// FilesUploadBucket is the bucket files are stored into before being processed
	FilesUploadBucket = "filesuploadbucket"
	// RtcCostUsd is the price of a single RTC in USD
	RtcCostUsd = 0.125
)

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

// GetIPFSEndpoint is used to construct the api url to connect to
// for private ipfs networks. in the case of dev mode it returns
// an default, non nexus based ipfs api address
func (api *API) GetIPFSEndpoint(networkName string) string {
	if dev {
		return api.cfg.IPFS.APIConnection.Host + ":" + api.cfg.IPFS.APIConnection.Port
	}
	return api.cfg.Nexus.Host + ":" + api.cfg.Nexus.Delegator.Port + "/network/" + networkName
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
		return errors.New(eh.FileTooBigError)
	}
	return nil
}

func (api *API) validateBlockchain(blockchain string) bool {
	switch blockchain {
	case "ethereum", "bitcoin", "litecoin", "monero", "dash":
		return true
	}
	return false
}

// generateEmailJWTToken is used to generate a jwt token used to validate emails
func (api *API) generateEmailJWTToken(username, verificationString string) (string, error) {
	// generate a jwt with claims to verify email
	verificationJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"user":                    username,
		"emailVerificationString": verificationString,
		"expire":                  time.Now().Add(time.Hour * 24).UTC().String(),
	})
	// return a signed version of the jwt
	return verificationJWT.SignedString([]byte(api.cfg.API.JWT.Key))
}

func (api *API) verifyEmailJWTToken(jwtString, username string) error {
	// parse the jwt for a token
	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unable to validate signing method: %v", token.Header["alg"])
		} else if method != jwt.SigningMethodHS512 {
			return nil, errors.New("expect hs512 signing method")
		}
		// return byte version of signing key
		return []byte(api.cfg.JWT.Key), nil
	})
	// verify jwt was parsed properly
	if err != nil {
		return err
	}
	// verify that the token is valid
	if !token.Valid {
		return errors.New("failed to validate token")
	}
	// extract claims from token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("failed to parse claims")
	}
	// verify the username matches what we are expected
	if claims["user"] != username {
		return fmt.Errorf("username from claim does not match expected user of %s", username)
	}
	// get user model so we can validate the email verification string
	user, err := api.um.FindByUserName(username)
	if err != nil {
		return errors.New(eh.UserSearchError)
	}
	emailVerificationString, ok := claims["emailVerificationString"].(string)
	if !ok {
		return errors.New("failed to convert verification token to string")
	}
	// validate email verification string
	if claims["emailVerificationString"] != user.EmailVerificationToken {
		return errors.New("failed to validate email verification token")
	}
	// ensure we can cast claims["expire"] to string type
	expireString, ok := claims["expire"].(string)
	if !ok {
		return errors.New("failed to convert expire value to string")
	}
	// parse expire string into time.Time
	expireTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", expireString)
	if err != nil {
		return err
	}
	// validate that the token hasn't expired
	if time.Now().UTC().Unix() > expireTime.Unix() {
		return errors.New("token is expired")
	}
	// enable email activity
	if _, err := api.um.ValidateEmailVerificationToken(username, emailVerificationString); err != nil {
		return err
	}
	return nil
}

// validateUserCredits is used to validate whether or not a user has enough credits to pay for an action
// and if they do, it is deducted from their account
func (api *API) validateUserCredits(username string, cost float64) error {
	availableCredits, err := api.um.GetCreditsForUser(username)
	if err != nil {
		return err
	}
	if availableCredits < cost {
		return errors.New(eh.InvalidBalanceError)
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
		api.l.With("user", username, "call_type", callType, "error", err.Error()).Error(eh.CreditRefundError)
	}
}

// validateAdminRequest is used to validate whether or not the requesting user is an administrator
func (api *API) validateAdminRequest(username string) error {
	isAdmin, err := api.um.CheckIfAdmin(username)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New(eh.UnAuthorizedAdminAccess)
	}
	return nil
}

func (api *API) formatUploadErrorMessage(file string, currentDataUsedBytes, maxDataAllowedBytes uint64) string {
	currentDataUsedGB := float64(currentDataUsedBytes) / float64(datasize.GB.Bytes())
	maxDataAllowedGB := float64(maxDataAllowedBytes) / float64(datasize.GB.Bytes())
	return fmt.Sprintf(
		"uploading object %s would breach your current data limit of %vGB as you are currently using %vGB, please upload a smaller object",
		file, maxDataAllowedGB, currentDataUsedGB,
	)
}

func (api *API) extractPostForms(c *gin.Context, formNames ...string) map[string]string {
	forms := make(map[string]string)
	for _, name := range formNames {
		value, exists := c.GetPostForm(name)
		if !exists {
			FailWithMissingField(c, name)
			return nil
		}
		forms[name] = value
	}
	return forms
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
// from https://golangcode.com/unzip-files-in-go/
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filenames, err
			}

		}
	}
	return filenames, nil
}
