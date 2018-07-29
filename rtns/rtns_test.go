package rtns_test

import (
	"testing"
	"time"

	"github.com/RTradeLtd/Temporal/rtns"
	lci "github.com/libp2p/go-libp2p-crypto"
)

var testPath = "QmNZiPk974vDsPmQii3YbrMKfi12KTSNM7XMiYyiea4VYZ"

func TestRTNS(t *testing.T) {
	im, err := rtns.InitializeWithNewKey()
	if err != nil {
		t.Fatal(err)
	}

	err = im.GenerateEDKeyPair(1024)
	if err != nil {
		t.Error(err)
	}

	err = im.GenerateKeyPair(lci.RSA, 1024)
	if err != nil {
		t.Fatal(err)
	}

	timeT := time.Now().Add(time.Hour * 24)
	_, err = im.CreateEntryWithEmbed(testPath, timeT)
	if err != nil {
		t.Fatal(err)
	}

}
