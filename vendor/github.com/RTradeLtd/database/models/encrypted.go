package models

import (
	"strings"

	"github.com/RTradeLtd/gorm"
)

// EncryptedUpload is an uploaded that has been encrypted by Temporal
type EncryptedUpload struct {
	gorm.Model
	UserName      string `gorm:"type:varchar(255)"`
	FileName      string `gorm:"type:varchar(255)"`
	FileNameUpper string `gorm:"type:varchar(255)"`
	FileNameLower string `gorm:"type:varchar(255)"`
	NetworkName   string `gorm:"type:varchar(255)"`
	IPFSHash      string `gorm:"type:varchar(255)"`
}

// EncryptedUploadManager is used to manipulate encrypted uplaods
type EncryptedUploadManager struct {
	DB *gorm.DB
}

// NewEncryptedUploadManager is used to generate our db helper
func NewEncryptedUploadManager(db *gorm.DB) *EncryptedUploadManager {
	return &EncryptedUploadManager{DB: db}
}

// NewUpload is used to store a new encrypted upload in the database
func (ecm *EncryptedUploadManager) NewUpload(username, filename, networname, ipfsHash string) (*EncryptedUpload, error) {
	eu := &EncryptedUpload{
		UserName:      username,
		FileName:      filename,
		FileNameLower: strings.ToLower(filename),
		FileNameUpper: strings.ToUpper(filename),
		NetworkName:   networname,
		IPFSHash:      ipfsHash,
	}

	if err := ecm.DB.Create(eu).Error; err != nil {
		return nil, err
	}
	return eu, nil
}

// FindUploadsByUser is used to find all uploads for a given user
func (ecm *EncryptedUploadManager) FindUploadsByUser(username string) (*[]EncryptedUpload, error) {
	uploads := []EncryptedUpload{}
	if err := ecm.DB.Where("user_name = ?", username).Find(&uploads).Error; err != nil {
		return nil, err
	}
	return &uploads, nil
}
