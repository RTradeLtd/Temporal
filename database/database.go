package database

import (
	"fmt"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var UploadObj *models.Upload
var UserObj *models.User
var PinPaymentObj *models.PinPayment
var FilePaymentObj *models.FilePayment
var IpnsObj *models.IPNS
var HostedIpfsNetObj *models.HostedIPFSPrivateNetwork

type DatabaseManager struct {
	DB     *gorm.DB
	Upload *models.UploadManager
}

func Initialize(cfg *config.TemporalConfig, runMigrations bool) (*DatabaseManager, error) {
	dbm := DatabaseManager{}
	db, err := OpenDBConnection(
		cfg.Database.Password,
		cfg.Database.URL,
		cfg.Database.Username)
	if err != nil {
		return nil, err
	}
	dbm.DB = db
	if runMigrations {
		dbm.RunMigrations()
	}
	return &dbm, nil
}

func (dbm *DatabaseManager) RunMigrations() {
	dbm.DB.AutoMigrate(UploadObj)
	dbm.DB.AutoMigrate(UserObj)
	dbm.DB.AutoMigrate(PinPaymentObj)
	dbm.DB.AutoMigrate(FilePaymentObj)
	// gorm will default table to name of ip_ns
	// so we will override with ipns
	dbm.DB.AutoMigrate(IpnsObj)
	dbm.DB.AutoMigrate(HostedIpfsNetObj)
	//dbm.DB.Model(userObj).Related(uploadObj.Users)
}

// OpenDBConnection is used to create a database connection
func OpenDBConnection(dbPass, dbURL, dbUser string) (*gorm.DB, error) {
	if dbUser == "" {
		dbUser = "postgres"
	}
	// look into whether or not we wil disable sslmode
	dbConnURL := fmt.Sprintf("host=%s port=5432 user=%s dbname=temporal password=%s", dbURL, dbUser, dbPass)
	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		return nil, err
	}
	return db, nil
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
func CloseDBConnection(db *gorm.DB) error {
	err := db.Close()
	if err != nil {
		return err
	}
	return nil
}
