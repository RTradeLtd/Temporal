package rtfs

import (
	"errors"
	"fmt"

	"github.com/RTradeLtd/rtfs/krab"

	mnemonics "github.com/RTradeLtd/entropy-mnemonics"
	ci "github.com/libp2p/go-libp2p-crypto"
)

// KeystoreManager is howe we manipulat keys
type KeystoreManager struct {
	store *krab.Krab
}

// NewKeystoreManager instantiates a new keystore manager. Takes an optional
// filepath for the store.
func NewKeystoreManager(store *krab.Krab) (*KeystoreManager, error) {
	return &KeystoreManager{
		store: store,
	}, nil
}

// CheckIfKeyExists is used to check if a key exists
func (km *KeystoreManager) CheckIfKeyExists(keyName string) (bool, error) {
	return km.store.Has(keyName)
}

// GetPrivateKeyByName is used to get a private key by its name
func (km *KeystoreManager) GetPrivateKeyByName(keyName string) (ci.PrivKey, error) {
	return km.store.Get(keyName)
}

// ListKeyIdentifiers will list out all key IDs (aka, public hashes)
func (km *KeystoreManager) ListKeyIdentifiers() ([]string, error) {
	return km.store.List()
}

// SavePrivateKey is used to save a private key under the specified name
func (km *KeystoreManager) SavePrivateKey(keyName string, pk ci.PrivKey) error {
	return km.store.Put(keyName, pk)
}

// CreateAndSaveKey is used to create a key of the given type and size
func (km *KeystoreManager) CreateAndSaveKey(keyName string, keyType, bits int) (ci.PrivKey, error) {
	if present, err := km.store.Has(keyName); err == nil {
		return nil, fmt.Errorf("failed to check for key '%s': %s", keyName, err.Error())
	} else if present {
		return nil, errors.New("key name already exists")
	}
	var pk ci.PrivKey
	var err error
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
	if err = km.SavePrivateKey(keyName, pk); err != nil {
		return nil, err
	}

	return pk, nil
}

// ExportKeyAsMnemonic is used to take an IPFS key, and return a human-readable friendly version.
// The idea is to allow users to easily export the keys they create, allowing them to take control of their records (ipns, tns, etc..)
func (km *KeystoreManager) ExportKeyAsMnemonic(keyName string) (string, error) {
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
	return phrase.String(), nil
}

// MnemonicToKey takes an exported mnemonic phrase, and converts it to a private key
func MnemonicToKey(phrase string) (ci.PrivKey, error) {
	mnemonicBytes, err := mnemonics.FromString(phrase, mnemonics.English)
	if err != nil {
		return nil, err
	}
	return ci.UnmarshalPrivateKey(mnemonicBytes)
}
