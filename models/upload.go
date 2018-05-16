package models

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Upload struct {
	gorm.Model
	Hash             string `gorm:not null;`
	Type             string `gorm:not null;` //  file, pin
	HoldTimeInMonths int64  `gorm:not null;`
	UploadAddress    string `gorm:not null;`
}

type UploadManager struct {
	DB *gorm.DB
}

// NewUploadManager is used to generate an upload manager interface
func NewUploadManager(db *gorm.DB) *UploadManager {
	return &UploadManager{DB: db}
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
func (um *UploadManager) GetUploads() []*Upload {
	var uploads []*Upload
	um.DB.Find(uploads)
	return uploads
}

// GetUploadsForAddress is used to retrieve all uploads by an address
func (um *UploadManager) GetUploadsForAddress(address string) []*Upload {
	var uploads []*Upload
	um.DB.Where("upload_address = ?", address).Find(&uploads)
	return uploads
}

// AddHash his used to add a hash to our database
func AddHash(c *gin.Context) error {
	var upload Upload
	hash := c.Param("hash")
	address, exists := c.GetPostForm("uploadAddress")
	if !exists {
		c.AbortWithError(http.StatusBadRequest, errors.New("uploadAddress param des not exist"))
		return errors.New("holdTime param does not exist")
	}
	holdTime, exists := c.GetPostForm("holdTime")
	if !exists {
		c.AbortWithError(http.StatusBadRequest, errors.New("holdTime param does not exist"))
		return errors.New("holdTime param does not exist")
	}
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return err
	}
	upload.Hash = fmt.Sprintf("%s", hash)
	upload.Type = "pin"
	upload.HoldTimeInMonths = holdTimeInt
	upload.UploadAddress = address
	db := OpenDBConnection()
	db.Create(&upload)
	db.Close()
	return nil
}

// AddFileHash is used to add the hash of a file to our database
func AddFileHash(c *gin.Context, hash string) {
	var upload Upload
	address := c.PostForm("uploadAddress")
	holdTimeInt, err := strconv.ParseInt(c.PostForm("holdTime"), 10, 64)
	if err != nil {
		c.Error(err)
	}
	upload.HoldTimeInMonths = holdTimeInt
	upload.UploadAddress = address
	upload.Hash = hash
	upload.Type = "file"
	db := OpenDBConnection()
	db.AutoMigrate(&upload)
	db.Create(&upload)
	db.Close()
}

// OpenDBConnection is used to open a database connection
func OpenDBConnection() *gorm.DB {
	db, err := gorm.Open("sqlite3", "./ipfs_database.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}
