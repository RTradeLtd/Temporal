package middleware

import (
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"

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
	apiMiddleware := APIRestrictionMiddleware(db)
	if reflect.TypeOf(apiMiddleware).String() != "gin.HandlerFunc" {
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
	testRecorder := httptest.NewRecorder()
	testCtx, _ := gin.CreateTestContext(testRecorder)
	if token, valid := jwt.Authenticator("testuser", "admin", testCtx); !valid {
		t.Fatal("failed to authenticate user")
	} else if token != "testuser" {
		t.Fatal("failed to authenticate")
	}
	if valid := jwt.Authorizator("testuser", testCtx); !valid {
		t.Fatal("failed to authorize user")
	}
	if valid := jwt.Authorizator("testuser2", testCtx); valid {
		t.Fatal("failed to authorize user")
	}
	jwt.Unauthorized(testCtx, 401, "unauthorized access")
	if testRecorder.Code != 401 {
		t.Fatal("failed to validate http status code")
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
