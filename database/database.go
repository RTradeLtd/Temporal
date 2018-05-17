package database

import (
	"log"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var uploadObj *models.Upload
var userObj *models.User

type DatabaseManager struct {
	DB     *gorm.DB
	Upload *models.UploadManager
}

func Initialize() *DatabaseManager {
	dbm := DatabaseManager{}
	db := OpenDBConnection()
	dbm.DB = db
	dbm.RunMigrations()
	return &dbm
}

func (dbm *DatabaseManager) RunMigrations() {
	dbm.DB.AutoMigrate(uploadObj)
}

// OpenDBConnection is used to create a database connection
func OpenDBConnection() *gorm.DB {
	db, err := gorm.Open("postgres", "host=127.0.0.1 port=5432 user=postgres dbname=temporal password=password123")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// CloseDBConnection is used to close a db
func CloseDBConnection(db *gorm.DB) {
	db.Close()
}
