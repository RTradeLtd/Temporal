package database

import (
	"fmt"
	"log"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var uploadObj *models.Upload
var userObj *models.User
var paymentObj *models.Payment

type DatabaseManager struct {
	DB     *gorm.DB
	Upload *models.UploadManager
}

func Initialize(dbPass string) *DatabaseManager {
	dbm := DatabaseManager{}
	db := OpenDBConnection(dbPass)
	dbm.DB = db
	dbm.RunMigrations()
	return &dbm
}

func (dbm *DatabaseManager) RunMigrations() {
	dbm.DB.AutoMigrate(uploadObj)
	dbm.DB.AutoMigrate(userObj)
	dbm.DB.AutoMigrate(paymentObj)
	//dbm.DB.Model(userObj).Related(uploadObj.Users)
}

// OpenDBConnection is used to create a database connection
func OpenDBConnection(dbPass string) *gorm.DB {
	// look into whether or not we wil disable sslmode
	dbConnURL := fmt.Sprintf("host=127.0.0.1 port=5432 user=postgres dbname=temporal password=%s sslmode=disable", dbPass)
	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// CloseDBConnection is used to close a db
func CloseDBConnection(db *gorm.DB) {
	db.Close()
}
