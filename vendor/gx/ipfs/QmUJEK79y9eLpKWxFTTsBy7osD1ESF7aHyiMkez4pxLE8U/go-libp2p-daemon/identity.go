package p2pd

import (
	"io/ioutil"

	crypto "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
)

func ReadIdentity(path string) (crypto.PrivKey, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(bytes)
}

func WriteIdentity(k crypto.PrivKey, path string) error {
	bytes, err := crypto.MarshalPrivateKey(k)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bytes, 0400)
}
