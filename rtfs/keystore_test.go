package rtfs_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/rtfs"
	ci "github.com/libp2p/go-libp2p-crypto"
)

/*
	if we dont disable gocache, the key generate will always be the same
	env GOCACHE=off go test -v ....
*/
func TestKeystoreManager(t *testing.T) {
	km, err := rtfs.GenerateKeystoreManager()
	if err != nil {
		t.Fatal(err)
	}

	keys, err := km.ListKeyIdentifiers()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range keys {
		pk, err := km.GetPrivateKeyByName(v)
		if err != nil {
			t.Error(err)
			continue
		}
		present, err := km.CheckIfKeyIsPresent(v)
		if err != nil {
			t.Error(err)
			continue
		}
		if !present {
			t.Error("key not present when it should be")
			continue
		}
		fmt.Println(pk)
	}

	// DO NOT USE 1024 IN PRODUCTION, >= 2048
	pk, _, err := ci.GenerateKeyPair(ci.RSA, 1024)
	if err != nil {
		t.Fatal(err)
	}

	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		t.Fatal(err)
	}

	hexed := hex.EncodeToString(b)
	fmt.Println(hexed)
	err = km.SavePrivateKey(hexed, pk)
	if err != nil {
		t.Fatal(err)
	}
}
