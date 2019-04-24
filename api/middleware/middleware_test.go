package middleware

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/gin-gonic/gin"

	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2"
)

func TestRequestIDMiddleware(t *testing.T) {
	testRecorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(testRecorder)
	engine.Use(RequestID())
	req, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	engine.GET("/foo", func(c *gin.Context) {
		c.String(200, "hello")
	})
	engine.ServeHTTP(testRecorder, req)
	if testRecorder.HeaderMap.Get("X-Request-ID") == "" {
		t.Fatal("failed to set a request header")
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
	logger := zaptest.NewLogger(t).Sugar()
	jwt := JwtConfigGenerate(cfg.JWT.Key, cfg.JWT.Realm, db.DB, logger)
	if reflect.TypeOf(jwt).String() != "*jwt.GinJWTMiddleware" {
		t.Fatal("failed to reflect correct middleware type")
	}
	if jwt.Realm != cfg.JWT.Realm {
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
	cors := CORSMiddleware(true, DefaultAllowedOrigins)
	if reflect.TypeOf(cors).String() != "gin.HandlerFunc" {
		t.Fatal("failed to reflect correct middleware type")
	}
	cors = CORSMiddleware(false, DefaultAllowedOrigins)
	if reflect.TypeOf(cors).String() != "gin.HandlerFunc" {
		t.Fatal("failed to reflect correct middleware type")
	}
}

func TestSecMiddleware(t *testing.T) {
	sec := NewSecWare(false)
	if reflect.TypeOf(sec).String() != "gin.HandlerFunc" {
		t.Fatal("failed to reflect correct middleware type")
	}
	sec = NewSecWare(true)
	if reflect.TypeOf(sec).String() != "gin.HandlerFunc" {
		t.Fatal("failed to reflect correct middleware type")
	}
}

func loadDatabase(cfg *config.TemporalConfig) (*database.Manager, error) {
	return database.New(cfg, database.Options{
		SSLModeDisable: true,
	})
}
