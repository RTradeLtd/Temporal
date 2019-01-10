package models

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/RTradeLtd/database/utils"
	"github.com/RTradeLtd/gorm"
	"github.com/lib/pq"
)

// Upload is a file or pin based upload to temporal
type Upload struct {
	gorm.Model
	Hash               string `gorm:"type:varchar(255);not null;"`
	Type               string `gorm:"type:varchar(255);not null;"` //  file, pin
	NetworkName        string `gorm:"type:varchar(255)"`
	HoldTimeInMonths   int64  `gorm:"type:integer;not null;"`
	UserName           string `gorm:"type:varchar(255);not null;"`
	GarbageCollectDate time.Time
	UserNames          pq.StringArray `gorm:"type:text[];not null;"`
	Encrypted          bool           `gorm:"type:bool"`
}

const dev = true

// UploadManager is used to manipulate upload objects in the database
type UploadManager struct {
	DB *gorm.DB
}

// NewUploadManager is used to generate an upload manager interface
func NewUploadManager(db *gorm.DB) *UploadManager {
	return &UploadManager{DB: db}
}

// UploadOptions is used to configure an upload
type UploadOptions struct {
	NetworkName      string
	Username         string
	HoldTimeInMonths int64
	Encrypted        bool
}

// NewUpload is used to create a new upload in the database
func (um *UploadManager) NewUpload(contentHash, uploadType string, opts UploadOptions) (*Upload, error) {
	_, err := um.FindUploadByHashAndNetwork(contentHash, opts.NetworkName)
	if err == nil {
		// this means that there is already an upload in hte database matching this content hash and network name, so we will skip
		return nil, errors.New("attempting to create new upload entry when one already exists in database")
	}
	holdInt, err := strconv.Atoi(fmt.Sprintf("%+v", opts.HoldTimeInMonths))
	if err != nil {
		return nil, err
	}
	upload := Upload{
		Hash:               contentHash,
		Type:               uploadType,
		NetworkName:        opts.NetworkName,
		HoldTimeInMonths:   opts.HoldTimeInMonths,
		UserName:           opts.Username,
		GarbageCollectDate: utils.CalculateGarbageCollectDate(holdInt),
		UserNames:          []string{opts.Username},
		Encrypted:          opts.Encrypted,
	}
	if check := um.DB.Create(&upload); check.Error != nil {
		return nil, check.Error
	}
	return &upload, nil
}

// UpdateUpload is used to upadte an already existing upload
func (um *UploadManager) UpdateUpload(holdTimeInMonths int64, username, contentHash, networkName string) (*Upload, error) {
	upload, err := um.FindUploadByHashAndNetwork(contentHash, networkName)
	if err != nil {
		return nil, err
	}
	isUploader := false
	upload.UserName = username
	for _, v := range upload.UserNames {
		if username == v {
			isUploader = true
			break
		}
	}
	if !isUploader {
		upload.UserNames = append(upload.UserNames, username)
	}
	holdInt, err := strconv.Atoi(fmt.Sprintf("%v", holdTimeInMonths))
	if err != nil {
		return nil, err
	}
	oldGcd := upload.GarbageCollectDate
	newGcd := utils.CalculateGarbageCollectDate(holdInt)
	if newGcd.Unix() > oldGcd.Unix() {
		upload.HoldTimeInMonths = holdTimeInMonths
		upload.GarbageCollectDate = oldGcd
	}
	if check := um.DB.Save(upload); check.Error != nil {
		return nil, err
	}
	return upload, nil
}

// RunDatabaseGarbageCollection is used to parse through the database
// and delete all objects whose GCD has passed
// TODO: Maybe move this to the database file?
func (um *UploadManager) RunDatabaseGarbageCollection() (*[]Upload, error) {
	var uploads []Upload
	var deletedUploads []Upload

	if check := um.DB.Find(&uploads); check.Error != nil {
		return nil, check.Error
	}
	for _, v := range uploads {
		if time.Now().Unix() > v.GarbageCollectDate.Unix() {
			if check := um.DB.Delete(&v); check.Error != nil {
				return nil, check.Error
			}
			deletedUploads = append(deletedUploads, v)
		}
	}
	return &deletedUploads, nil
}

// RunTestDatabaseGarbageCollection is used to run a test garbage collection run.
// NOTE that this will delete literally every single object it detects.
func (um *UploadManager) RunTestDatabaseGarbageCollection() (*[]Upload, error) {
	var foundUploads []Upload
	var deletedUploads []Upload
	if !dev {
		return nil, errors.New("not in dev mode")
	}
	// get all uploads
	if check := um.DB.Find(&foundUploads); check.Error != nil {
		return nil, check.Error
	}
	for _, v := range foundUploads {
		if check := um.DB.Delete(v); check.Error != nil {
			return nil, check.Error
		}
		deletedUploads = append(deletedUploads, v)
	}
	return &deletedUploads, nil
}

// FindUploadsByNetwork is used to find all uploads corresponding to a given network
func (um *UploadManager) FindUploadsByNetwork(networkName string) (*[]Upload, error) {
	uploads := &[]Upload{}
	if check := um.DB.Where("network_name = ?", networkName).Find(uploads); check.Error != nil {
		return nil, check.Error
	}
	return uploads, nil
}

// FindUploadByHashAndNetwork is used to search for an upload by its hash, and the network it was stored on
func (um *UploadManager) FindUploadByHashAndNetwork(hash, networkName string) (*Upload, error) {
	upload := &Upload{}
	if check := um.DB.Where("hash = ? AND network_name = ?", hash, networkName).First(upload); check.Error != nil {
		return nil, check.Error
	}
	return upload, nil
}

// FindUploadsByHash is used to return all instances of uploads matching the given hash
func (um *UploadManager) FindUploadsByHash(hash string) *[]Upload {

	uploads := []Upload{}

	um.DB.Find(&uploads).Where("hash = ?", hash)

	return &uploads
}

// GetUploadByHashForUser is used to retrieve the last (most recent) upload for a user
func (um *UploadManager) GetUploadByHashForUser(hash string, username string) []*Upload {
	var uploads []*Upload
	um.DB.Find(&uploads).Where("hash = ? AND user_name = ?", hash, username)
	return uploads
}

// GetUploads is used to return all  uploads
func (um *UploadManager) GetUploads() (*[]Upload, error) {
	uploads := []Upload{}
	if check := um.DB.Find(&uploads); check.Error != nil {
		return nil, check.Error
	}
	return &uploads, nil
}

// GetUploadsForUser is used to retrieve all uploads by a user name
func (um *UploadManager) GetUploadsForUser(username string) (*[]Upload, error) {
	uploads := []Upload{}
	if check := um.DB.Where("user_name = ?", username).Find(&uploads); check.Error != nil {
		return nil, check.Error
	}
	return &uploads, nil
}
