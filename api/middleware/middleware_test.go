package middleware

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/RTradeLtd/Temporal/log"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/jinzhu/gorm"
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
	logger, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
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
	// test a failed auth
	if _, valid := jwt.Authenticator("testuser22", "admin22", testCtx); valid {
		t.Fatal("user authenticated when auth should've failed")
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
