package models

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/RTradeLtd/database/utils"
	"github.com/RTradeLtd/gorm"
)

const (
	// ErrShorterGCD is an error triggered when updating to update an upload for a user
	// with a hold time that would result in a shorter garbage collection date
	ErrShorterGCD = "upload would not extend garbage collection date so there is no need to process"
	// ErrAlreadyExistingUpload is an error triggered when attempting to insert  a new row into the database
	// for a content that already exists in the database for a user. This means you should be using the UpdateUpload
	// function to allow for updating garbage collection dates.
	ErrAlreadyExistingUpload = "the content you are inserting into the database already exists, please use the UpdateUpload function"
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
	Encrypted          bool `gorm:"type:bool"`
}

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
	_, err := um.FindUploadByHashAndUserAndNetwork(opts.Username, contentHash, opts.NetworkName)
	if err == nil {
		// this means that there is already an upload in hte database matching this content hash and network name, so we will skip
		return nil, errors.New(ErrAlreadyExistingUpload)
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
		Encrypted:          opts.Encrypted,
	}
	if check := um.DB.Create(&upload); check.Error != nil {
		return nil, check.Error
	}
	return &upload, nil
}

// UpdateUpload is used to update the garbage collection time for an already existing upload
func (um *UploadManager) UpdateUpload(holdTimeInMonths int64, username, contentHash, networkName string) (*Upload, error) {
	upload, err := um.FindUploadByHashAndUserAndNetwork(username, contentHash, networkName)
	if err != nil {
		return nil, err
	}
	oldGcd := upload.GarbageCollectDate
	newGcd := utils.CalculateGarbageCollectDate(int(holdTimeInMonths))
	if newGcd.Unix() < oldGcd.Unix() {
		return nil, errors.New(ErrShorterGCD)
	}
	upload.HoldTimeInMonths = holdTimeInMonths
	upload.GarbageCollectDate = newGcd
	if check := um.DB.Save(upload); check.Error != nil {
		return nil, err
	}
	return upload, nil
}

// FindUploadsByNetwork is used to find all uploads corresponding to a given network
func (um *UploadManager) FindUploadsByNetwork(networkName string) ([]Upload, error) {
	uploads := []Upload{}
	if check := um.DB.Where("network_name = ?", networkName).Find(&uploads); check.Error != nil {
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
func (um *UploadManager) FindUploadsByHash(hash string) ([]Upload, error) {
	uploads := []Upload{}
	if err := um.DB.Where("hash = ?", hash).Find(&uploads).Error; err != nil {
		return nil, err
	}
	return uploads, nil
}

// FindUploadByHashAndUserAndNetwork is used to look for an upload based off its hash, user, and network
func (um *UploadManager) FindUploadByHashAndUserAndNetwork(username, hash, networkName string) (*Upload, error) {
	upload := &Upload{}
	if err := um.DB.Where("user_name = ? AND hash = ? AND network_name = ?", username, hash, networkName).First(upload).Error; err != nil {
		return nil, err
	}
	return upload, nil
}

// GetUploadByHashForUser is used to retrieve the last (most recent) upload for a user
func (um *UploadManager) GetUploadByHashForUser(hash string, username string) ([]Upload, error) {
	uploads := []Upload{}
	if err := um.DB.Where("hash = ? AND user_name = ?", hash, username).Find(&uploads).Error; err != nil {
		return nil, err
	}
	return uploads, nil
}

// GetUploads is used to return all  uploads
func (um *UploadManager) GetUploads() ([]Upload, error) {
	uploads := []Upload{}
	if check := um.DB.Find(&uploads); check.Error != nil {
		return nil, check.Error
	}
	return uploads, nil
}

// GetUploadsForUser is used to retrieve all uploads by a user name
func (um *UploadManager) GetUploadsForUser(username string) ([]Upload, error) {
	uploads := []Upload{}
	if check := um.DB.Where("user_name = ?", username).Find(&uploads); check.Error != nil {
		return nil, check.Error
	}
	return uploads, nil
}

// ExtendGarbageCollectionPeriod is used to extend the garbage collection period for a particular upload
func (um *UploadManager) ExtendGarbageCollectionPeriod(username, hash, network string, holdTimeInMonths int) error {
	upload, err := um.FindUploadByHashAndUserAndNetwork(username, hash, network)
	if err != nil {
		return err
	}
	// update garbage collection period
	upload.GarbageCollectDate = upload.GarbageCollectDate.AddDate(0, holdTimeInMonths, 0)
	// save the updated model
	return um.DB.Model(upload).Update("garbage_collect_date", upload.GarbageCollectDate).Error
}
