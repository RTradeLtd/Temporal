package rtfs_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfs"
)

const testPIN = "QmNZiPk974vDsPmQii3YbrMKfi12KTSNM7XMiYyiea4VYZ"

func TestInitialize(t *testing.T) {
	im, err := rtfs.Initialize("", "")
	if err != nil {
		t.Fatal(err)
	}
	info, err := im.Shell.ID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(info)
}

func TestPin(t *testing.T) {
	im, err := rtfs.Initialize("", "")
	if err != nil {
		t.Fatal(err)
	}
	err = im.Pin(testPIN)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetObjectFileSizeInBytes(t *testing.T) {
	im, err := rtfs.Initialize("", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = im.GetObjectFileSizeInBytes(testPIN)
	if err != nil {
		t.Fatal(err)
	}
}

func TestObjectStat(t *testing.T) {
	im, err := rtfs.Initialize("", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = im.ObjectStat(testPIN)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseLocalPinsForHash(t *testing.T) {
	im, err := rtfs.Initialize("", "")
	if err != nil {
		t.Fatal(err)
	}
	exists, err := im.ParseLocalPinsForHash(testPIN)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal(err)
	}
}

func TestPubSub(t *testing.T) {
	im, err := rtfs.Initialize("", "")
	if err != nil {
		t.Fatal(err)
	}
	err = im.SubscribeToPubSubTopic(im.PubTopic)
	if err != nil {
		t.Fatal(err)
	}
	err = im.PublishPubSubMessage(im.PubTopic, "data")
	if err != nil {
		t.Fatal(err)
	}
}
