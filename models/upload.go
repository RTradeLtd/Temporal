package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type Upload struct {
	gorm.Model
	Hash               string `gorm:"type:varchar(255);not null;"`
	Type               string `gorm:"type:varchar(255);not null;"` //  file, pin
	NetworkName        string `gorm:"type:varchar(255)"`
	HoldTimeInMonths   int64  `gorm:"type:integer;not null;"`
	UploadAddress      string `gorm:"type:varchar(255);not null;"`
	GarbageCollectDate time.Time
	UploaderAddresses  pq.StringArray `gorm:"type:text[];not null;"`
}

const dev = true

type UploadManager struct {
	DB *gorm.DB
}

// NewUploadManager is used to generate an upload manager interface
func NewUploadManager(db *gorm.DB) *UploadManager {
	return &UploadManager{DB: db}
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

func (um *UploadManager) FindUploadsByNetwork(networkName string) ([]*Upload, error) {
	uploads := []*Upload{}
	if check := um.DB.Where("network_name = ?", networkName).Find(uploads); check.Error != nil {
		return nil, check.Error
	}
	return uploads, nil
}
func (um *UploadManager) FindUploadByHashAndNetwork(hash, networkName string) (*Upload, error) {
	upload := &Upload{}
	if check := um.DB.Where("hash = ? AND network_name = ?", hash, networkName).First(upload); check.Error != nil {
		return nil, check.Error
	}
	return upload, nil
}

// FindUploadsByHash is used to return all instances of uploads matching the
// given hash
func (um *UploadManager) FindUploadsByHash(hash string) []*Upload {

	uploads := []*Upload{}

	um.DB.Find(&uploads).Where("hash = ?", hash)

	return uploads
}

// GetUploadByHashForUploader is used to retrieve the last (most recent) upload for a user
func (um *UploadManager) GetUploadByHashForUploader(hash string, uploaderAddress string) []*Upload {
	var uploads []*Upload
	um.DB.Find(&uploads).Where("hash = ? AND uploader_address = ?", hash, uploaderAddress)
	return uploads
}

// GetUploads is used to return all  uploads
func (um *UploadManager) GetUploads() *[]Upload {
	var uploads []Upload
	um.DB.Find(&uploads)
	return &uploads
}

// GetUploadsForAddress is used to retrieve all uploads by an address
func (um *UploadManager) GetUploadsForAddress(address string) *[]Upload {
	var uploads []Upload
	um.DB.Where("upload_address = ?", address).Find(&uploads)
	return &uploads
}

// AddPinHash is used to upload a pin hash to our database
func (um *UploadManager) AddPinHash(hash string, uploaderAddress string, networkName string, holdTimeInMonths int64) {
	var upload Upload
	upload.HoldTimeInMonths = holdTimeInMonths
	upload.UploadAddress = uploaderAddress
	upload.Hash = hash
	upload.NetworkName = networkName
	upload.Type = "pin"
	um.DB.Create(upload)
}

// AddFileHash is used to add the hash of a file to our database
func (um *UploadManager) AddFileHash(hash string, uploaderAddress string, networkName string, holdTimeInMonths int64) {
	var upload Upload
	upload.HoldTimeInMonths = holdTimeInMonths
	upload.UploadAddress = uploaderAddress
	upload.Hash = hash
	upload.NetworkName = networkName
	upload.Type = "file"
	um.DB.Create(&upload)
}
