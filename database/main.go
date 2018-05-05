package database

import (
	"log"

	"github.com/RTradeLtd/RTC-IPFS/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

// OpenDBConnection is used to create a database connection
func OpenDBConnection() *gorm.DB {
	db, err := gorm.Open("sqlite3", "./ipfs_database.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// CloseDBConnection is used to close a db
func CloseDBConnection(db *gorm.DB) {
	db.Close()
}

// AddHash his used to add a hash to our database
func AddHash(c *gin.Context) {
	var upload models.Upload
	hash := c.Param("hash")
	upload.Hash = hash
	db := OpenDBConnection()
	db.Create(&upload)
	db.Close()
}

// AddFileHash is used to add the hash of a file to our database
func AddFileHash(hash string) {
	var upload models.Upload
	upload.Hash = hash
	db := OpenDBConnection()
	db.AutoMigrate(&upload)
	db.Create(&upload)
	db.Close()
}
