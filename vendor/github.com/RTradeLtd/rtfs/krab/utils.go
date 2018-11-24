package krab

import (
	"fmt"
	"strings"
)

// all of this is taken from https://github.com/ipfs/go-ipfs-keystore

var (
	// ErrNoSuchKey is returned if a key of the given name is not found in the store
	ErrNoSuchKey = "no key by the given name was found"
	// ErrKeyExists is returned when writing a key would overwrite an existing key
	ErrKeyExists = "key by that name already exists, refusing to overwrite"
	// ErrKeyFmt is returned when the key's format is invalid
	ErrKeyFmt = "key has invalid format"
)

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("%s: key names must be at least one character", ErrKeyFmt)
	}

	if strings.Contains(name, "/") {
		return fmt.Errorf("%s: key names may not contain slashes", ErrKeyFmt)
	}

	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("%s: key names may not begin with a period", ErrKeyFmt)
	}

	return nil
}
