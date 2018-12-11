package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/jinzhu/gorm"
)

// TestQueueIPFS is used to test IPFS queues
func TestQueuesIPFS(t *testing.T) {
	logFilePath = "../../templogs.log"
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err = loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		parentCmd string
		childCmd  string
	}
	tests := []struct {
		name string
		args args
	}{
		{"IPNSEntry", args{"ipfs", "ipns-entry"}},
		{"IPFSPin", args{"ipfs", "pin"}},
		{"IPFSFile", args{"ipfs", "file"}},
		{"IPFSKey", args{"ipfs", "key-creation"}},
		{"IPFSCluster", args{"ipfs", "cluster"}},
	}
	queueCmds := commands["queue"]

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
			queueCmds.Children[tt.args.parentCmd].Children[tt.args.childCmd].Action(*cfg, nil)
			cancel()
		})
	}
}

func TestQueuesDFA(t *testing.T) {
	logFilePath = "../../templogs.log"
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	commands["queue"].Children["dfa"].Action(*cfg, nil)
}

func TestQueuesEmailSend(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	commands["queue"].Children["email-send"].Action(*cfg, nil)
}

func TestMigrations(t *testing.T) {
	logFilePath = "../../templogs.log"
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	// this wont work with our test environment as the psql server doesn't have ssl
	//commands["migrate"].Action(*cfg, nil)
	commands["migrate-insecure"].Action(*cfg, nil)
}
func TestInit(t *testing.T) {
	logFilePath = "../../templogs.log"
	if err := os.Setenv("CONFIG_DAG", "../../testenv/new_config.json"); err != nil {
		t.Fatal(err)
	}
	commands["init"].Action(config.TemporalConfig{}, nil)
}

func TestAdmin(t *testing.T) {
	logFilePath = "../../templogs.log"
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
	logFilePath = "../../templogs.log"
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

func loadDatabase(cfg *config.TemporalConfig) (*gorm.DB, error) {
	return database.OpenDBConnection(database.DBOptions{
		User:           cfg.Database.Username,
		Password:       cfg.Database.Password,
		Address:        cfg.Database.URL,
		Port:           cfg.Database.Port,
		SSLModeDisable: true,
	})
}
