package api

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func TestAPI_LogError(t *testing.T) {
	api := API{l: log.New(), service: "test"}
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
		{"with message and no err", args{nil, "bye", nil}, "bye", "bye"},
		{"no message and no err", args{nil, "", nil}, "", ""},
		{"message and additional fields", args{nil, "hi", []interface{}{"wow", "amazing"}}, "amazing", "hi"},
		{"message and odd fields should ignore fields", args{nil, "hi", []interface{}{"wow"}}, "hi", "hi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			api.l.Out = &buf
			r := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(r)

			// log error and execute callback
			if tt.args.fields != nil {
				api.LogError(tt.args.err, tt.args.message, tt.args.fields...)(c)
			} else {
				api.LogError(tt.args.err, tt.args.message)(c)
			}

			// check log output
			b, err := ioutil.ReadAll(&buf)
			if err != nil {
				t.Error(err)
			}
			if !strings.Contains(string(b), tt.wantLog) {
				t.Errorf("got %s, want %s", string(b), tt.wantLog)
			}

			// check http response
			b, err = ioutil.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
			}
			if !strings.Contains(string(b), tt.wantResp) {
				t.Errorf("got %s, want %s", string(b), tt.wantResp)
			}
		})
	}
}
