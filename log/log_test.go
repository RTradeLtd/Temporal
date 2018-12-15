package log

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	type args struct {
		logpath string
		dev     bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"dev-no-path", args{"", true}, false},
		{"prod-no-path", args{"", false}, false},
		{"dev-with-path", args{"./tmp/log", true}, false},
		{"prod-with-path", args{"./tmp/log", false}, false},
		{"bad-dir-dev", args{"/root/toor", true}, true},
		{"bad-dir-prod", args{"/root/toor", false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSugar, err := NewLogger(tt.args.logpath, tt.args.dev)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotSugar == nil {
				t.Error("got unexpected nil logger")
			}
		})
	}
}

func TestNewProcessLogger(t *testing.T) {
	l, out := NewTestLogger()
	logger := NewProcessLogger(l, "network_up", "id", "1234")
	logger.Info("hi")
	if out.All()[0].ContextMap()["network_up.id"].(string) != "1234" {
		t.Error("bad logger")
	}
}
