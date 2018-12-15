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
		{"Dev", args{"./tmp/dev.log", true}, false},
		{"Prod", args{"./tmp/prog.log", false}, false},
		{"DevFailDir", args{"./tmp", true}, true},
		{"ProdFailDir", args{"./tmp", false}, true},
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
