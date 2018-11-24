// Package dshelp provides utilities for parsing and creating
// datastore keys used by go-ipfs
package dshelp

import (
	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	"gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore"
	"gx/ipfs/QmfVj3x4D6Jkq9SEoi5n2NmoUomLwoeiwnYz2KQa15wRw6/base32"
)

// NewKeyFromBinary creates a new key from a byte slice.
func NewKeyFromBinary(rawKey []byte) datastore.Key {
	buf := make([]byte, 1+base32.RawStdEncoding.EncodedLen(len(rawKey)))
	buf[0] = '/'
	base32.RawStdEncoding.Encode(buf[1:], rawKey)
	return datastore.RawKey(string(buf))
}

// BinaryFromDsKey returns the byte slice corresponding to the given Key.
func BinaryFromDsKey(k datastore.Key) ([]byte, error) {
	return base32.RawStdEncoding.DecodeString(k.String()[1:])
}

// CidToDsKey creates a Key from the given Cid.
func CidToDsKey(k cid.Cid) datastore.Key {
	return NewKeyFromBinary(k.Bytes())
}

// DsKeyToCid converts the given Key to its corresponding Cid.
func DsKeyToCid(dsKey datastore.Key) (cid.Cid, error) {
	kb, err := BinaryFromDsKey(dsKey)
	if err != nil {
		return cid.Cid{}, err
	}
	return cid.Cast(kb)
}
