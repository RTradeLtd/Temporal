package queue

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestQueue_LogError(t *testing.T) {
	qm := Manager{
		logger: log.New(),
	}
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
			qm.logger.Out = &buf
			// log error and execute callback
			if tt.args.fields != nil {
				qm.LogError(tt.args.err, tt.args.message, tt.args.fields...)
			} else {
				qm.LogError(tt.args.err, tt.args.message)
			}

			// check log output
			b, err := ioutil.ReadAll(&buf)
			if err != nil {
				t.Error(err)
			}
			if !strings.Contains(string(b), tt.wantLog) {
				t.Errorf("got %s, want %s", string(b), tt.wantLog)
			}
		})
	}
}
