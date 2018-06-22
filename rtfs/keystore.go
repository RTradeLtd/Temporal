package rtfs

/*
Utilities for manipulating the IPFS fs-keystore

consider adding a mutex keystore instead
*/

import (
	ksm "github.com/ipfs/go-ipfs-keystore"
	ci "github.com/libp2p/go-libp2p-crypto"
)

type KeystoreManager struct {
	FSKeystore *ksm.FSKeystore
}

func GenerateKeystoreManager() (*KeystoreManager, error) {
	var km KeystoreManager
	fsk, err := ksm.NewFSKeystore(DefaultFSKeystorePath)
	if err != nil {
		return nil, err
	}
	km.FSKeystore = fsk
	return &km, nil
}

func (km *KeystoreManager) CheckIfKeyIsPresent(keyName string) (bool, error) {
	present, err := km.FSKeystore.Has(keyName)
	if err != nil {
		return false, err
	}
	if !present {
		return false, nil
	}
	return true, nil
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
