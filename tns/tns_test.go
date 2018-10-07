package tns_test

import (
	"errors"
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
	if _, err := tns.GenerateTNSManager(testZoneName); err != nil {
		t.Fatal(err)
	}
}

func TestTNS_MakeHostNoOpts(t *testing.T) {
	manager, err := tns.GenerateTNSManager(testZoneName)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(nil); err != nil {
		t.Fatal(err)
	}
}

func TestTNS_MakeHostWithOpts(t *testing.T) {
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
	if err = manager.MakeHost(opts); err != nil {
		t.Fatal(err)
	}
}

func TestTNS_HostMultiAddress(t *testing.T) {
	manager, err := tns.GenerateTNSManager(testZoneName)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(nil); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.HostMultiAddress(); err != nil {
		t.Fatal(err)
	}
}

func TestTNS_ReachableAddress(t *testing.T) {
	manager, err := tns.GenerateTNSManager(testZoneName)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(nil); err != nil {
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
		fmt.Println(addr)
		count++
	}
}

func TestTNS_RunTNSFail(t *testing.T) {
	manager, err := tns.GenerateTNSManager(testZoneName)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.MakeHost(nil); err != nil {
		t.Fatal(err)
	}
	defer manager.Host.Close()
	go manager.RunTNS()
	if err = manager.QueryTNS(manager.Host.ID()); err == nil {
		t.Fatal(errors.New("no error encountered when one should've been"))
	}
}
