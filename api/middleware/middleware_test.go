package middleware

import (
	"os"
	"reflect"
	"testing"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const (
	testRealm = "test-realm"
)

func TestAPIMiddleware(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	api := APIRestrictionMiddleware(db)
	if reflect.TypeOf(api).String() != "gin.HandlerFunc" {
		t.Fatal("failed to reflect correct middleware type")
	}
}

func TestJwtMiddleware(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	logger := log.New()
	logger.Out = os.Stdout
	jwt := JwtConfigGenerate(cfg.API.JwtKey, testRealm, db, logger)
	if reflect.TypeOf(jwt).String() != "*jwt.GinJWTMiddleware" {
		t.Fatal("failed to reflect correct middleware type")
	}
	if jwt.Realm != testRealm {
		t.Fatal("failed to set correct realm name")
	}
}

func TestCORSMiddleware(t *testing.T) {
	cors := CORSMiddleware()
	if reflect.TypeOf(cors).String() != "gin.HandlerFunc" {
		t.Fatal("failed to reflect correct middleware type")
	}
}

func loadDatabase(cfg *config.TemporalConfig) (*gorm.DB, error) {
	return database.OpenDBConnection(database.DBOptions{
		User:           cfg.Database.Username,
		Password:       cfg.Database.Password,
		Address:        cfg.Database.URL,
		Port:           cfg.Database.Port,
		SSLModeDisable: true,
	})
}
