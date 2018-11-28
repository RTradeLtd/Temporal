package krab

import (
	"bytes"
	"errors"
	"strings"

	"github.com/RTradeLtd/crypto"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	badger "github.com/ipfs/go-ds-badger"
	ci "github.com/libp2p/go-libp2p-crypto"
)

// Krab is used to manage an encrypted IPFS keystore
type Krab struct {
	em *crypto.EncryptManager
	ds *badger.Datastore
}

// Opts is used to configure a Krab keystore
type Opts struct {
	Passphrase string
	DSPath     string
	ReadOnly   bool
}

// NewKrab is used to create a new krab ipfs keystore manager
func NewKrab(opts Opts) (*Krab, error) {
	badgerOpts := &badger.DefaultOptions
	badgerOpts.ReadOnly = opts.ReadOnly
	ds, err := badger.NewDatastore(opts.DSPath, badgerOpts)
	if err != nil {
		return nil, err
	}
	return &Krab{
		em: crypto.NewEncryptManager(opts.Passphrase),
		ds: ds,
	}, nil
}

// Has is used to check whether or not the given key name exists
func (km *Krab) Has(name string) (bool, error) {
	if err := validateName(name); err != nil {
		return false, err
	}
	if has, err := km.ds.Has(ds.NewKey(name)); err != nil {
		return false, err
	} else if !has {
		return false, errors.New(ErrNoSuchKey)
	}
	return true, nil
}

// Put is used to store a key in our keystore
func (km *Krab) Put(name string, privKey ci.PrivKey) error {
	if err := validateName(name); err != nil {
		return err
	}
	if has, err := km.Has(name); err == nil {
		return errors.New(ErrKeyExists)
	} else if has {
		return errors.New(ErrKeyExists)
	}
	pkBytes, err := privKey.Bytes()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(pkBytes)
	// encrypt the private key
	encryptedPK, err := km.em.Encrypt(reader)
	if err != nil {
		return err
	}
	return km.ds.Put(ds.NewKey(name), encryptedPK)
}

// Get is used to retrieve a key from our keystore
func (km *Krab) Get(name string) (ci.PrivKey, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if has, err := km.Has(name); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New(ErrNoSuchKey)
	}
	encryptedPKBytes, err := km.ds.Get(ds.NewKey(name))
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(encryptedPKBytes)
	pkBytes, err := km.em.Decrypt(reader)
	if err != nil {
		return nil, err
	}
	return ci.UnmarshalPrivateKey(pkBytes)
}

// Delete is used to remove a key from our keystore
func (km *Krab) Delete(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	return km.ds.Delete(ds.NewKey(name))
}

// List is used to list all key identifiers in our keystore
func (km *Krab) List() ([]string, error) {
	entries, err := km.ds.Query(query.Query{})
	if err != nil {
		return nil, err
	}
	keys, err := entries.Rest()
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, v := range keys {
		ids = append(ids, strings.Split(v.Key, "/")[1])
	}
	return ids, nil
}

// Close is used to close our badger connection
func (km *Krab) Close() error {
	return km.ds.Close()
}
