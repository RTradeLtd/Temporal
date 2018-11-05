package database

import (
	"errors"
	"fmt"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/config"
	"github.com/jinzhu/gorm"

	// import our postgres dialect used to talk with a postgres databse
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	uploadObj          *models.Upload
	encryptedUploadObj *models.EncryptedUpload
	userObj            *models.User
	paymentObj         *models.Payments
	ipnsObj            *models.IPNS
	hostedIpfsNetObj   *models.HostedIPFSPrivateNetwork
	dropObj            *models.Drop
	tnsZoneObj         *models.Zone
	tnsRecordObj       *models.Record
)

// Manager is used to manage databases
type Manager struct {
	DB     *gorm.DB
	Upload *models.UploadManager
}

// Options is used to configure a connection to the database
type Options struct {
	RunMigrations  bool
	SSLModeDisable bool
	LogMode        bool
}

// Initialize is used to init our connection to a database, and return a manager struct
func Initialize(cfg *config.TemporalConfig, opts Options) (*Manager, error) {
	if cfg == nil {
		return nil, errors.New("invalid configuration provided")
	}

	db, err := OpenDBConnection(DBOptions{
		User:           cfg.Database.Username,
		Password:       cfg.Database.Password,
		Address:        cfg.Database.URL,
		Port:           cfg.Database.Port,
		SSLModeDisable: opts.SSLModeDisable,
	})
	if err != nil {
		return nil, err
	}

	db.LogMode(opts.LogMode)

	dbm := Manager{DB: db}
	if opts.RunMigrations {
		dbm.RunMigrations()
	}
	return &dbm, nil
}

// RunMigrations runs all migrations
func (dbm *Manager) RunMigrations() {
	dbm.DB.AutoMigrate(uploadObj)
	dbm.DB.AutoMigrate(userObj)
	dbm.DB.AutoMigrate(paymentObj)
	// gorm will default table to name of ip_ns
	// so we will override with ipns
	dbm.DB.AutoMigrate(ipnsObj)
	dbm.DB.AutoMigrate(hostedIpfsNetObj)
	dbm.DB.AutoMigrate(dropObj)
	dbm.DB.AutoMigrate(encryptedUploadObj)
	dbm.DB.AutoMigrate(tnsZoneObj)
	dbm.DB.AutoMigrate(tnsRecordObj)
}

// Close shuts down database connection
func (dbm *Manager) Close() error { return dbm.DB.Close() }

// DBOptions declares options for opening a database connection
type DBOptions struct {
	User           string
	Password       string
	Address        string
	Port           string
	SSLModeDisable bool
}

// OpenDBConnection is used to create a database connection
func OpenDBConnection(opts DBOptions) (*gorm.DB, error) {
	if opts.User == "" {
		opts.User = "postgres"
	}
	// look into whether or not we wil disable sslmode
	dbConnURL := fmt.Sprintf("host=%s port=%s user=%s dbname=temporal password=%s",
		opts.Address, opts.Port, opts.User, opts.Password)
	if opts.SSLModeDisable {
		dbConnURL = "sslmode=disable " + dbConnURL
	}
	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		return nil, err
	}
	return db, nil
}
