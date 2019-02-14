package clients_test

import (
	"testing"

	clients "github.com/RTradeLtd/Temporal/grpc-clients"
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
	cfg.Services.Lens.TLS.CertPath = "fakepath"
	if _, err = clients.NewLensClient(cfg.Services); err == nil {
		t.Fatal("error expected")
	}
}
func TestLensClient_Pass(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = clients.NewLensClient(cfg.Services); err != nil {
		t.Fatal(err)
	}
}

func TestSignerClient_Pass(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = clients.NewSignerClient(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestOrchestratorClient_Fail(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Nexus.TLS.CertPath = "fakepath"
	if _, err = clients.NewOcrhestratorClient(cfg.Nexus); err == nil {
		t.Fatal("error expected")
	}
}
func TestOrchestratorClient_Pass(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = clients.NewOcrhestratorClient(cfg.Nexus); err != nil {
		t.Fatal(err)
	}
}

func TestKaasClient_Pass(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	// test non fallback
	if _, err := clients.NewKaasClient(cfg.Services, false); err != nil {
		t.Fatal(err)
	}
	if _, err := clients.NewKaasClient(cfg.Services, true); err != nil {
		t.Fatal(err)
	}
}

func TestKaasClient_Fail(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Services.Krab.URL = "notarealurls"
	cfg.Services.Krab.TLS.CertPath = "/root/toor"
	cfg.Services.Krab.Fallback.URL = "notarealurls"
	cfg.Services.Krab.Fallback.TLS.CertPath = "/root/toor"
	if _, err := clients.NewKaasClient(cfg.Services, false); err == nil {
		t.Fatal("expected error")
	}
	if _, err := clients.NewKaasClient(cfg.Services, true); err == nil {
		t.Fatal("expected error")
	}
}
