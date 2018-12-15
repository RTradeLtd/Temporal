package log

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	type args struct {
		logPath string
		dev     bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Dev", args{"../testenv/dev.log", true}, false},
		{"NoDev", args{"../testenv/nodev.log", false}, false},
		{"DevFail", args{"../testenv", true}, true},
		{"NoDevFail", args{"../testenv", false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewLogger(tt.args.logPath, tt.args.dev); (err != nil) != tt.wantErr {
				t.Fatalf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
