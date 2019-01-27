package main

import (
	"context"
	"testing"
	"time"

	"github.com/RTradeLtd/config"
)

func init() {
	var t = true
	dbNoSSL = &t
	dbMigrate = &t
	devMode = &t
	grpcNoSSL = &t
	var blank string
	configPath = &blank
}

func TestAPI(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		port          string
		listenAddress string
		certFilePath  string
		keyFilePath   string
		logDir        string
	}
	tests := []struct {
		name string
		args args
	}{
		{"NoTLS-NoPort-NoLogDir", args{"", "127.0.0.1", "", "", ""}},
		{"NoTLS-WithPort-WithLogDir", args{"6768", "127.0.0.1", "", "", "./tmp/"}},
		{"TLS-WithPort-WithLogDir", args{"6769", "127.0.0.1", "../../testenv/certs/api.cert", "../../testenv/certs/api.key", "./tmp/"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up call args
			apiPort = &tt.args.port
			flags := map[string]string{
				"listenAddress": tt.args.listenAddress,
				"certFilePath":  tt.args.certFilePath,
				"keyFilePath":   tt.args.keyFilePath,
			}
			cfg.LogDir = tt.args.logDir
			// setup global context
			ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
			commands["api"].Action(*cfg, flags)
			cancel()
		})
	}
}

// TestQueueIPFS is used to test IPFS queues
func TestQueuesIPFS(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		parentCmd string
		childCmd  string
		logDir    string
	}
	tests := []struct {
		name string
		args args
	}{
		{"IPNSEntry-NoLogDir", args{"ipfs", "ipns-entry", ""}},
		{"IPNSEntry-LogDir", args{"ipfs", "ipns-entry", "./tmp/"}},
		{"IPFSPin-NoLogDir", args{"ipfs", "pin", ""}},
		{"IPFSPin-LogDir", args{"ipfs", "pin", "./tmp/"}},
		{"IPFSKey-NoLogDir", args{"ipfs", "key-creation", ""}},
		{"IPFSKey-LogDir", args{"ipfs", "key-creation", "./tmp/"}},
		{"IPFSCluster-NoLogDir", args{"ipfs", "cluster", ""}},
		{"IPFSCluster-LogDir", args{"ipfs", "cluster", "./tmp/"}},
	}
	queueCmds := commands["queue"]
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg.LogDir = tt.args.logDir
			ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			queueCmds.Children[tt.args.parentCmd].Children[tt.args.childCmd].Action(*cfg, nil)
		})
	}
}

func TestQueuesDFA(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		logDir string
	}
	tests := []struct {
		name string
		args args
	}{
		{"NoLogDir", args{""}},
		{"LogDir", args{"./tmp"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg.LogDir = tt.args.logDir
			ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			commands["queue"].Children["dfa"].Action(*cfg, nil)
		})
	}
}

func TestQueuesEmailSend(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		logDir string
	}
	tests := []struct {
		name string
		args args
	}{
		{"NoLogDir", args{""}},
		{"LogDir", args{"./tmp"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg.LogDir = tt.args.logDir
			ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			commands["queue"].Children["email-send"].Action(*cfg, nil)
		})
	}
}

func TestMigrations(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	// this wont work with our test environment as the psql server doesn't have ssl
	//commands["migrate"].Action(*cfg, nil)
	commands["migrate"].Action(*cfg, nil)
}

func TestInit(t *testing.T) {
	*configPath = "tmp/new_config.json"
	commands["init"].Action(config.TemporalConfig{}, nil)
}

func TestAdmin(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	flags := map[string]string{
		"dbAdmin": "testuser",
	}
	commands["admin"].Action(*cfg, flags)
}

func TestUser(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	flags := map[string]string{
		"user":  "myuser",
		"pass":  "mypass",
		"email": "myuser+test@example.org",
	}
	commands["user"].Action(*cfg, flags)
}
