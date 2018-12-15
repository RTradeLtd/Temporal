package v2

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/gin-gonic/gin"
)

func TestAPI_LogInfo(t *testing.T) {
	type args struct {
		message string
	}
	tests := []struct {
		name        string
		args        args
		wantMessage string
	}{
		{"Single", args{"hello"}, "hello"},
		{"Double", args{"hello world"}, "hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observer, out := observer.New(zap.InfoLevel)
			logger := zap.New(observer).Sugar()
			api := API{l: logger, service: "test"}
			//	r := httptest.NewRecorder()
			//	c, _ := gin.CreateTestContext(r)
			api.LogInfo(tt.args.message)
			// check log message and context
			b, _ := json.Marshal(out.All()[0].ContextMap())
			entry := out.All()[0].Message + string(b)
			if !strings.Contains(entry, tt.wantMessage) {
				t.Errorf("got %s, want %s", entry, tt.wantMessage)
			}
		})
	}
}
func TestAPI_LogError(t *testing.T) {
	type args struct {
		err     error
		message string
		fields  []interface{}
	}
	tests := []struct {
		name     string
		args     args
		wantLog  string
		wantResp string
	}{
		{"with err no message", args{errors.New("hi"), "", nil}, "hi", "hi"},
		{"with err and message", args{errors.New("hi"), "bye", nil}, "hi", "bye"},
		{"message and additional fields", args{errors.New("hi"), "hi", []interface{}{"wow", "amazing"}}, "amazing", "hi"},
		{"message and odd fields should ignore fields", args{errors.New("hi"), "hi", []interface{}{"wow"}}, "hi", "hi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observer, out := observer.New(zap.InfoLevel)
			logger := zap.New(observer).Sugar()
			api := API{l: logger, service: "test"}

			// log error and execute callback
			r := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(r)
			if tt.args.fields != nil {
				api.LogError(tt.args.err, tt.args.message, tt.args.fields...)(c)
			} else {
				api.LogError(tt.args.err, tt.args.message)(c)
			}

			// check log message and context
			b, _ := json.Marshal(out.All()[0].ContextMap())
			entry := out.All()[0].Message + string(b)
			if !strings.Contains(entry, tt.wantLog) {
				t.Errorf("got %s, want %s", entry, tt.wantLog)
			}

			// check http response
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
			}
			if !strings.Contains(string(b), tt.wantResp) {
				t.Errorf("got %s, want %s", string(b), tt.wantResp)
			}
		})
	}
}
