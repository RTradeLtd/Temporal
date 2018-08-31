package keystore

import (
	errors "github.com/pkg/errors"
	"strings"
)

func validateName(name string) error {
	if name == "" {
		return errors.Wrap(ErrKeyFmt, "key names must be at least one character")
	}

	if strings.Contains(name, "/") {
		return errors.Wrap(ErrKeyFmt, "key names may not contain slashes")
	}

	if strings.HasPrefix(name, ".") {
		return errors.Wrap(ErrKeyFmt, "key names may not begin with a period")
	}

	return nil
}
