package app

import (
	"testing"

	"github.com/RTradeLtd/config"
)

func TestNew(t *testing.T) {
	type args struct {
		cmds map[string]Cmd
		cfg  Config
	}
	tests := []struct {
		name     string
		args     args
		wantCmds int
	}{
		{"with version", args{make(map[string]Cmd), Config{Version: "1"}}, 2},
		{"without version", args{make(map[string]Cmd), Config{}}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.cmds, tt.args.cfg)
			if len(got.cmds) != tt.wantCmds {
				t.Errorf("expected %d commands, got %d", len(got.cmds), tt.wantCmds)
			}
		})
	}
}

func TestApp_PreRun(t *testing.T) {
	type fields struct {
		cfg  Config
		cmds map[string]Cmd
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantFound bool
	}{
		{
			"help command should run",
			fields{Config{}, make(map[string]Cmd)},
			args{[]string{"help"}},
			true,
		},
		{
			"invalid command should not run",
			fields{Config{}, make(map[string]Cmd)},
			args{[]string{"asdfasdf"}},
			false,
		},
		{
			"non-prerun command should not run",
			fields{Config{}, map[string]Cmd{"notme": Cmd{Action: func(config.TemporalConfig, map[string]string) {}}}},
			args{[]string{"notme"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(tt.fields.cmds, tt.fields.cfg)
			exit := a.PreRun(tt.args.args)
			if (exit == 0) != tt.wantFound {
				t.Errorf("expected command run to be %v, got %v",
					tt.wantFound, (exit == 0))
			}
		})
	}
}

func TestApp_Run(t *testing.T) {
	type fields struct {
		cfg  Config
		cmds map[string]Cmd
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantFound bool
	}{
		{
			"help command should run",
			fields{Config{}, make(map[string]Cmd)},
			args{[]string{"help"}},
			true,
		},
		{
			"invalid command should not run",
			fields{Config{}, make(map[string]Cmd)},
			args{[]string{"asdfasdf"}},
			false,
		},
		{
			"custom command should run",
			fields{Config{}, map[string]Cmd{"me": Cmd{Action: func(config.TemporalConfig, map[string]string) {}}}},
			args{[]string{"me"}},
			true,
		},
		{
			"nested commands should run",
			fields{Config{}, map[string]Cmd{"me": Cmd{
				Action: func(config.TemporalConfig, map[string]string) {},
				Children: map[string]Cmd{"too": Cmd{
					Children: map[string]Cmd{"wow": Cmd{
						Action: func(config.TemporalConfig, map[string]string) {},
					}},
				}},
			}}},
			args{[]string{"me", "too", "wow"}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(tt.fields.cmds, tt.fields.cfg)
			exit := a.Run(config.TemporalConfig{}, nil, tt.args.args)

			if (exit == 0) != tt.wantFound {
				t.Errorf("expected command run to be %v, got %v",
					tt.wantFound, (exit == 0))
			}
		})
	}
}
