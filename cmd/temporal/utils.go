package main

import (
	"path/filepath"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/jinzhu/gorm"
)

func logPath(base, file string) (logPath string) {
	if base == "" {
		logPath = filepath.Join(base, file)
	} else {
		logPath = filepath.Join(base, file)
	}
	return
}

func newDB(cfg config.TemporalConfig, noSSL bool) (*gorm.DB, error) {
	return database.OpenDBConnection(database.DBOptions{
		User:           tCfg.Database.Username,
		Password:       tCfg.Database.Password,
		Address:        tCfg.Database.URL,
		Port:           tCfg.Database.Port,
		SSLModeDisable: noSSL,
	})
}
