package rtfs

/*
Utilities for manipulating the IPFS fs-keystore

consider adding a mutex keystore instead

*/

import (
	"errors"

	"github.com/NebulousLabs/entropy-mnemonics"
	keystore "github.com/ipfs/go-ipfs-keystore"
	ci "github.com/libp2p/go-libp2p-crypto"
)

// KeystoreManager is used to manage ipfs keys
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

// CheckIfKeyExists is used to check if the specified key exists
func (km *KeystoreManager) CheckIfKeyExists(keyName string) (bool, error) {
	present, err := km.FSKeystore.Has(keyName)
	if err != nil {
		return false, err
	}
	return present, nil
}

// GetPrivateKeyByName is used search for a key by its name
func (km *KeystoreManager) GetPrivateKeyByName(keyName string) (ci.PrivKey, error) {
	pk, err := km.FSKeystore.Get(keyName)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

// ListKeyIdentifiers is used to list all known key IDs
func (km *KeystoreManager) ListKeyIdentifiers() ([]string, error) {
	keys, err := km.FSKeystore.List()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// SavePrivateKey is used to store a private key
func (km *KeystoreManager) SavePrivateKey(keyName string, pk ci.PrivKey) error {
	err := km.FSKeystore.Put(keyName, pk)
	if err != nil {
		return err
	}
	return nil
}

// CreateAndSaveKey is used to create, and save a key
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

// ExportKeyToMnemonic is used to take an IPFS key, and return a human-readable friendly version.
// The idea is to allow users to easily export the keys they create, allowing them to take control of their records (ipns, tns, etc..)
func (km *KeystoreManager) ExportKeyToMnemonic(keyName string) (string, error) {
	pk, err := km.GetPrivateKeyByName(keyName)
	if err != nil {
		return "", err
	}
	pkBytes, err := pk.Bytes()
	if err != nil {
		return "", err
	}
	phrase, err := mnemonics.ToPhrase(pkBytes, mnemonics.English)
	if err != nil {
		return "", err
	}
	pkBytes, err = mnemonics.FromPhrase(phrase, mnemonics.English)
	if err != nil {
		return "", err
	}
	pk2, err := ci.UnmarshalPrivateKey(pkBytes)
	if err != nil {
		return "", err
	}
	valid := pk.Equals(pk2)
	if !valid {
		return "", errors.New("failed to validate key")
	}
	return phrase.String(), nil
}

// MnemonicToKey takes an exported mnemonic phrase, and converts it to a private key
func (km *KeystoreManager) MnemonicToKey(stringPhrase string) (ci.PrivKey, error) {
	mnemonicBytes, err := mnemonics.FromString(stringPhrase, mnemonics.English)
	if err != nil {
		return nil, err
	}
	return ci.UnmarshalPrivateKey(mnemonicBytes)
}
