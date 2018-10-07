package tns_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/tns"
)

const (
	testZoneName  = "example.org"
	testIPAddress = "0.0.0.0"
	testPort      = "9999"
	testIPVersion = "ip4"
	testProtocol  = "tcp"
)

func TestTNS_NewTNSManager(t *testing.T) {
	t.Skip()
	if _, err := tns.GenerateTNSManager(testZoneName); err != nil {
		t.Fatal(err)
	}
}

func TestTNS_MakeHostNoOpts(t *testing.T) {
	t.Skip()
	manager, err := tns.GenerateTNSManager(testZoneName)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(manager.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
}

func TestTNS_MakeHostWithOpts(t *testing.T) {
	t.Skip()
	manager, err := tns.GenerateTNSManager(testZoneName)
	if err != nil {
		t.Fatal(err)
	}

	opts := &tns.HostOpts{
		IPAddress: testIPAddress,
		Port:      testPort,
		IPVersion: testIPVersion,
		Protocol:  testProtocol,
	}
	if err = manager.MakeHost(manager.PrivateKey, opts); err != nil {
		t.Fatal(err)
	}
}

func TestTNS_HostMultiAddress(t *testing.T) {
	t.Skip()
	manager, err := tns.GenerateTNSManager(testZoneName)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(manager.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.HostMultiAddress(); err != nil {
		t.Fatal(err)
	}
}

func TestTNS_ReachableAddress(t *testing.T) {
	t.Skip()
	manager, err := tns.GenerateTNSManager(testZoneName)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(manager.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
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
	t.Skip()
	manager, err := tns.GenerateTNSManager(testZoneName)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(manager.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	client, err := tns.GenerateTNSClient(true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.MakeHost(client.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	addr, err := manager.ReachableAddress(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.AddPeerToPeerStore(addr); err != nil {
		t.Fatal(err)
	}
}

func TestTNSClient_QueryTNS(t *testing.T) {
	manager, err := tns.GenerateTNSManager(testZoneName)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(manager.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	go manager.RunTNSDaemon()
	client, err := tns.GenerateTNSClient(true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.MakeHost(client.PrivateKey, nil); err != nil {
		t.Fatal(err)
	}
	addr, err := manager.ReachableAddress(0)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := client.AddPeerToPeerStore(addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("querying tns...")
	if err = client.QueryTNS(pid); err != nil {
		t.Fatal(err)
	}
}
