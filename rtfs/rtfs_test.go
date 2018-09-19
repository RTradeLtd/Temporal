package rtfs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfs"
)

const (
	testPIN        = "QmNZiPk974vDsPmQii3YbrMKfi12KTSNM7XMiYyiea4VYZ"
	nodeOneAPIAddr = "192.168.1.101:5001"
	nodeTwoAPIAddr = "192.168.2.101:5001"
)

func TestInitialize(t *testing.T) {
	im, err := rtfs.Initialize("", nodeOneAPIAddr)
	if err != nil {
		t.Fatal(err)
	}
	info, err := im.Shell.ID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(info)
}

func TestDHTFindProvs(t *testing.T) {
	im, err := rtfs.Initialize("", nodeOneAPIAddr)
	if err != nil {
		t.Fatal(err)
	}
	err = im.DHTFindProvs("QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv", "10")
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildCustomRequest(t *testing.T) {
	im, err := rtfs.Initialize("", nodeOneAPIAddr)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := im.BuildCustomRequest(context.Background(),
		nodeOneAPIAddr, "dht/findprovs", nil, "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", resp)
}

func TestPin(t *testing.T) {
	im, err := rtfs.Initialize("", nodeOneAPIAddr)
	if err != nil {
		t.Fatal(err)
	}

	// create pin
	if err = im.Pin(testPIN); err != nil {
		t.Fatal(err)
	}

	// check if pin was created
	exists, err := im.ParseLocalPinsForHash(testPIN)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("pin not found")
	}
}

func TestGetObjectFileSizeInBytes(t *testing.T) {
	im, err := rtfs.Initialize("", nodeOneAPIAddr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = im.GetObjectFileSizeInBytes(testPIN)
	if err != nil {
		t.Fatal(err)
	}
}

func TestObjectStat(t *testing.T) {
	im, err := rtfs.Initialize("", nodeOneAPIAddr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = im.ObjectStat(testPIN)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPubSub(t *testing.T) {
	im, err := rtfs.Initialize("", nodeOneAPIAddr)
	if err != nil {
		t.Fatal(err)
	}
	err = im.PublishPubSubMessage(im.PubTopic, "data")
	if err != nil {
		t.Fatal(err)
	}
}
