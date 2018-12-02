package clients_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/grpc-clients"
	"github.com/RTradeLtd/config"
)

const (
	testCfgPath = "../testenv/config.json"
)

func TestLensClient_Fail(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Endpoints.Lens.TLS.CertPath = "fakepath"
	if _, err = clients.NewLensClient(cfg.Endpoints); err == nil {
		t.Fatal("error expected")
	}
}
func TestLensClient_Pass(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = clients.NewLensClient(cfg.Endpoints); err != nil {
		t.Fatal(err)
	}
}

func TestSignerClient_Pass(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = clients.NewSignerClient(cfg, true); err != nil {
		t.Fatal(err)
	}
}

func TestOrchestratorClient_Fail(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Orchestrator.TLS.CertPath = "fakepath"
	if _, err = clients.NewOcrhestratorClient(cfg.Orchestrator); err == nil {
		t.Fatal("error expected")
	}
}
func TestOrchestratorClient_Pass(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = clients.NewOcrhestratorClient(cfg.Orchestrator); err != nil {
		t.Fatal(err)
	}
}
