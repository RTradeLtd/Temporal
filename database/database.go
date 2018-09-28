package database

import (
	"errors"
	"fmt"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	UploadObj        *models.Upload
	UserObj          *models.User
	PaymentObj       *models.Payments
	IpnsObj          *models.IPNS
	HostedIpfsNetObj *models.HostedIPFSPrivateNetwork
)

type DatabaseManager struct {
	DB     *gorm.DB
	Upload *models.UploadManager
}

func Initialize(cfg *config.TemporalConfig, runMigrations bool) (*DatabaseManager, error) {
	if cfg == nil {
		return nil, errors.New("invalid configuration provided")
	}

	db, err := OpenDBConnection(DBOptions{
		User:     cfg.Database.Name,
		Password: cfg.Database.Password,
		Address:  cfg.Database.URL,
	})
	if err != nil {
		return nil, err
	}

	dbm := DatabaseManager{DB: db}
	if runMigrations {
		dbm.RunMigrations()
	}
	return &dbm, nil
}

func (dbm *DatabaseManager) RunMigrations() {
	dbm.DB.AutoMigrate(UploadObj)
	dbm.DB.AutoMigrate(UserObj)
	dbm.DB.AutoMigrate(PaymentObj)
	// gorm will default table to name of ip_ns
	// so we will override with ipns
	dbm.DB.AutoMigrate(IpnsObj)
	dbm.DB.AutoMigrate(HostedIpfsNetObj)
	//dbm.DB.Model(userObj).Related(uploadObj.Users)
}

// DBOptions declares options for opening a database connection
type DBOptions struct {
	User           string
	Password       string
	Address        string
	SSLModeDisable bool
}

// OpenDBConnection is used to create a database connection
func OpenDBConnection(opts DBOptions) (*gorm.DB, error) {
	if opts.User == "" {
		opts.User = "postgres"
	}
	// look into whether or not we wil disable sslmode
	dbConnURL := fmt.Sprintf("host=%s port=5432 user=%s dbname=temporal password=%s",
		opts.Address, opts.User, opts.Password)
	if opts.SSLModeDisable {
		dbConnURL = "sslmode=disable " + dbConnURL
	}
	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func OpenTestDBConnection(dbPass string) (*gorm.DB, error) {
	dbConnURL := fmt.Sprintf("host=127.0.0.1 port=5433 user=postgres dbname=temporal password=%s", dbPass)
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
