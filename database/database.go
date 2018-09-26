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
	PaymentObj       *models.Payment
	IpnsObj          *models.IPNS
	HostedIpfsNetObj *models.HostedIPFSPrivateNetwork
)

type DatabaseManager struct {
	DB     *gorm.DB
	Upload *models.UploadManager
}

type DatabaseOptions struct {
	RunMigrations  bool
	SSLModeDisable bool
	LogMode        bool
}

func Initialize(cfg *config.TemporalConfig, opts DatabaseOptions) (*DatabaseManager, error) {
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

	dbm := DatabaseManager{DB: db}
	if opts.RunMigrations {
		dbm.RunMigrations()
	}
	return &dbm, nil
}

// RunMigrations runs all migrations
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

// Close shuts down database connection
func (dbm *DatabaseManager) Close() error { return dbm.DB.Close() }

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
