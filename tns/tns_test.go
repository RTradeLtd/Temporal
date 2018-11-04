package tns_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/tns"
	"github.com/RTradeLtd/config"
)

// Issue with libp2p and being unable to run multiple tests one after another
// need to debug to figure out how we can avoid this

const (
	testZoneName              = "example.org"
	testIPAddress             = "0.0.0.0"
	testPort                  = "9999"
	testIPVersion             = "ip4"
	testProtocol              = "tcp"
	testCfgPath               = "../test/config.json"
	defaultZoneName           = "myzone"
	defaultZoneManagerKeyName = "postables-3072"
	defaultZoneKeyName        = "postables-testkeydemo"
	defaultZoneUserName       = "postables"
	defaultRecordName         = "myrecord"
	defaultRecordKeyName      = "postables-testkeydemo2"
	defaultRecordUserName     = "postables"
	testPIN                   = "QmNZiPk974vDsPmQii3YbrMKfi12KTSNM7XMiYyiea4VYZ"
)

func TestTNS_Echo(t *testing.T) {
	manager, err := tns.GenerateTNSManager(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(manager.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	defer manager.Host.Close()
	go manager.RunTNSDaemon()
	client, err := tns.GenerateTNSClient(true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.MakeHost(client.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	defer client.Host.Close()
	addr, err := manager.ReachableAddress(0)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := client.AddPeerToPeerStore(addr)
	if err != nil {
		t.Fatal(err)
	}
	intf, err := client.QueryTNS(pid, "echo", nil)
	if err != nil {
		t.Fatal(err)
	}
	if intf != nil {
		t.Fatal("intf is not nil when it should be")
	}
}

func TestTNS_ZoneRequestFail(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	dbm, err := database.Initialize(cfg, database.DatabaseOptions{SSLModeDisable: true})
	if err != nil {
		t.Fatal(err)
	}
	manager, err := tns.GenerateTNSManager(nil, dbm.DB)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ZM.NewZone(
		defaultZoneUserName,
		defaultZoneName,
		defaultZoneManagerKeyName,
		defaultZoneKeyName,
		testPIN,
	); err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(manager.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	defer manager.Host.Close()
	go manager.RunTNSDaemon()
	client, err := tns.GenerateTNSClient(true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.MakeHost(client.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	defer client.Host.Close()
	addr, err := manager.ReachableAddress(0)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := client.AddPeerToPeerStore(addr)
	if err != nil {
		t.Fatal(err)
	}
	intf, err := client.QueryTNS(pid, "zone-request", tns.ZoneRequest{
		ZoneName:           defaultZoneName,
		ZoneManagerKeyName: defaultZoneManagerKeyName,
		UserName:           defaultZoneUserName,
	})
	if err != nil {
		t.Fatal(err)
	}
	if intf == nil {
		t.Fatal("intf is nil when it shouldnt be")
	}
}

func TestTNS_GenerateTNSClient(t *testing.T) {
	if _, err := tns.GenerateTNSClient(true, nil); err != nil {
		t.Fatal(err)
	}
}
func TestTNS_ClientMakeHost(t *testing.T) {
	c, err := tns.GenerateTNSClient(true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = c.MakeHost(c.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	c.Host.Close()
}

func TestTNS_GenerateTNSManager(t *testing.T) {
	if _, err := tns.GenerateTNSManager(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestTNS_ManagerMakeHost(t *testing.T) {
	m, err := tns.GenerateTNSManager(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.MakeHost(m.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	m.Host.Close()
}

func TestTNS_HostMultiAddress(t *testing.T) {
	manager, err := tns.GenerateTNSManager(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(manager.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	defer manager.Host.Close()
	if _, err = manager.HostMultiAddress(); err != nil {
		t.Fatal(err)
	}
}

func TestTNS_ReachableAddress(t *testing.T) {
	manager, err := tns.GenerateTNSManager(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(manager.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	defer manager.Host.Close()
	count := 0
	max := len(manager.Host.Addrs())
	for count < max {
		addr, err := manager.ReachableAddress(count)
		if err != nil {
			t.Fatal(err)
		}
		if addr == "" {
			t.Fatal("bad address constructed but no error")
		}
		count++
	}
}

func TestTNSClient_AddPeerToPeerStore(t *testing.T) {
	manager, err := tns.GenerateTNSManager(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(manager.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	defer manager.Host.Close()
	client, err := tns.GenerateTNSClient(true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.MakeHost(client.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	defer client.Host.Close()
	addr, err := manager.ReachableAddress(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.AddPeerToPeerStore(addr); err != nil {
		t.Fatal(err)
	}
}
