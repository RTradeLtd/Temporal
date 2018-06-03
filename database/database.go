package database

import (
	"fmt"
	"log"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var UploadObj *models.Upload
var UserObj *models.User
var PaymentObj *models.Payment

type DatabaseManager struct {
	DB     *gorm.DB
	Upload *models.UploadManager
}

func Initialize(dbPass, dbURL string) *DatabaseManager {
	dbm := DatabaseManager{}
	db := OpenDBConnection(dbPass, dbURL)
	dbm.DB = db
	dbm.RunMigrations()
	return &dbm
}

func (dbm *DatabaseManager) RunMigrations() {
	dbm.DB.AutoMigrate(UploadObj)
	dbm.DB.AutoMigrate(UserObj)
	dbm.DB.AutoMigrate(PaymentObj)
	//dbm.DB.Model(userObj).Related(uploadObj.Users)
}

// OpenDBConnection is used to create a database connection
func OpenDBConnection(dbPass, dbURL string) *gorm.DB {
	// look into whether or not we wil disable sslmode
	dbConnURL := fmt.Sprintf("host=%s port=5432 user=temporal dbname=temporal password=%s", dbURL, dbPass)
	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func OpenTestDBConnection(dbPass string) (*gorm.DB, error) {
	dbConnURL := fmt.Sprintf("host=127.0.0.1 port=5432 user=postgres dbname=temporal password=%s", dbPass)
	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// CloseDBConnection is used to close a db
func CloseDBConnection(db *gorm.DB) {
	db.Close()
}
