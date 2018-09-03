package database_test

import (
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/jinzhu/gorm"
)

var (
	travis = os.Getenv("TRAVIS") != ""
	dbPass string
)

func TestDatabase(t *testing.T) {
	db, err := gorm.Open(
		"postgres", "host=127.0.0.1 port=5433 user=postgres dbname=temporal password=password123 sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestDatabaseMigrations(t *testing.T) {
	db, err := database.OpenDBConnection(database.DBOptions{
		User:           "postgres",
		Password:       "password123",
		Address:        "127.0.0.1",
		SSLModeDisable: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(database.UploadObj)
	db.AutoMigrate(database.UserObj)
	db.AutoMigrate(database.PaymentObj)
	db.Close()
}
