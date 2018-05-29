package rtfs_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfs"
)

const testPIN = "QmNZiPk974vDsPmQii3YbrMKfi12KTSNM7XMiYyiea4VYZ"

func TestInitialize(t *testing.T) {
	im := rtfs.Initialize("")
	info, err := im.Shell.ID()
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}
	fmt.Println(info)
}

func TestPin(t *testing.T) {
	im := rtfs.Initialize("")
	err := im.Pin(testPIN)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("pin test successful")
}

func TestGetObjectFileSizeInBytes(t *testing.T) {
	im := rtfs.Initialize("")
	size, err := im.GetObjectFileSizeInBytes(testPIN)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("file size in bytes ", size)
}

func TestObjectStat(t *testing.T) {
	im := rtfs.Initialize("")
	stat, err := im.ObjectStat(testPIN)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("object stat ", stat)
}

func TestParseLocalPinsForHash(t *testing.T) {
	im := rtfs.Initialize("")
	exists, err := im.ParseLocalPinsForHash(testPIN)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal(err)
	}
}

func TestPubSub(t *testing.T) {
	im := rtfs.Initialize("test topic")
	err := im.SubscribeToPubSubTopic(im.PubTopic)
	if err != nil {
		t.Fatal(err)
	}
	err = im.PublishPubSubMessage(im.PubTopic, "data")
	if err != nil {
		t.Fatal(err)
	}
}
