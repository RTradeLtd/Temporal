package rtfs

/*
Utilities for manipulating the IPFS fs-keystore

consider adding a mutex keystore instead

*/

import (
	"errors"

	keystore "github.com/ipfs/go-ipfs-keystore"
	ci "github.com/libp2p/go-libp2p-crypto"
)

type KeystoreManager struct {
	FSKeystore *keystore.FSKeystore
}

// GenerateKeystoreManager instantiates a new keystore manager. Takes an optional
// filepath for the store.
func GenerateKeystoreManager(keystorePath ...string) (*KeystoreManager, error) {
	var (
		storePath = DefaultFSKeystorePath
		km        KeystoreManager
	)
	if keystorePath != nil && len(keystorePath) > 0 {
		storePath = keystorePath[0]
	}
	fsk, err := keystore.NewFSKeystore(storePath)
	if err != nil {
		return nil, err
	}
	km.FSKeystore = fsk
	return &km, nil
}

func (km *KeystoreManager) CheckIfKeyExists(keyName string) (bool, error) {
	present, err := km.FSKeystore.Has(keyName)
	if err != nil {
		return false, err
	}
	return present, nil
}

func (km *KeystoreManager) GetPrivateKeyByName(keyName string) (ci.PrivKey, error) {
	pk, err := km.FSKeystore.Get(keyName)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

func (km *KeystoreManager) ListKeyIdentifiers() ([]string, error) {
	keys, err := km.FSKeystore.List()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (km *KeystoreManager) SavePrivateKey(keyName string, pk ci.PrivKey) error {
	err := km.FSKeystore.Put(keyName, pk)
	if err != nil {
		return err
	}
	return nil
}

func (km *KeystoreManager) CreateAndSaveKey(keyName string, keyType, bits int) (ci.PrivKey, error) {
	var pk ci.PrivKey
	var err error

	present, err := km.FSKeystore.Has(keyName)
	if err != nil {
		return nil, err
	}
	if present {
		return nil, errors.New("key name already exists")
	}

	switch keyType {
	case ci.Ed25519:
		pk, _, err = ci.GenerateKeyPair(keyType, 256)
		if err != nil {
			return nil, err
		}
	case ci.RSA:
		pk, _, err = ci.GenerateKeyPair(keyType, bits)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("key type provided not a valid key type")
	}

	err = km.SavePrivateKey(keyName, pk)
	if err != nil {
		return nil, err
	}

	return pk, nil
}
