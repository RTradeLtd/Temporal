package database

import (
	"errors"
	"fmt"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/gorm"

	// import our postgres dialect used to talk with a postgres databse
	_ "github.com/RTradeLtd/gorm/dialects/postgres"
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
	Logger         Logger
}

// New is used to init our connection to a database, and return a manager struct
func New(cfg *config.TemporalConfig, opts Options) (*Manager, error) {
	if cfg == nil {
		return nil, errors.New("invalid configuration provided")
	}

	db, err := openDBConnection(dbOptions{
		User:           cfg.Database.Username,
		Password:       cfg.Database.Password,
		Address:        cfg.Database.URL,
		Port:           cfg.Database.Port,
		SSLModeDisable: opts.SSLModeDisable,
	})
	if err != nil {
		return nil, err
	}

	if opts.Logger != nil {
		db.SetLogger(opts.Logger)
	}
	db.LogMode(opts.LogMode)

	var dbm = Manager{DB: db}
	if opts.RunMigrations {
		dbm.RunMigrations()
	}
	return &dbm, nil
}

// RunMigrations runs all migrations
func (dbm *Manager) RunMigrations() {
	for _, t := range []interface{}{
		&models.Upload{},
		&models.EncryptedUpload{},
		&models.User{},
		&models.Payments{},
		&models.IPNS{},
		&models.HostedNetwork{},
		&models.Zone{},
		&models.Record{},
		&models.Usage{},
	} {
		dbm.DB.AutoMigrate(t)
	}
}

// Close shuts down database connection
func (dbm *Manager) Close() error { return dbm.DB.Close() }

// dbOptions declares options for opening a database connection
type dbOptions struct {
	User           string
	Password       string
	Address        string
	Port           string
	SSLModeDisable bool
}

// openDBConnection is used to create a database connection
func openDBConnection(opts dbOptions) (*gorm.DB, error) {
	if opts.User == "" {
		opts.User = "postgres"
	}

	dbConnURL := fmt.Sprintf("host=%s port=%s user=%s dbname=temporal password=%s",
		opts.Address, opts.Port, opts.User, opts.Password)
	if opts.SSLModeDisable {
		dbConnURL = "sslmode=disable " + dbConnURL
	}

	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection with database: %s", err.Error())
	}
	return db, nil
}
