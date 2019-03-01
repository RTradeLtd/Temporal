//+build !windows

package cli

import (
	"syscall"
)

func isErrnoNotSupported(err error) bool {
	switch err {
	case
		// Operation not supported
		syscall.EINVAL, syscall.EROFS, syscall.ENOTSUP,
		// File descriptor doesn't support syncing (found on MacOS).
		syscall.ENOTTY,
		// MacOS is weird. It returns EBADF when calling fsync on stdout
		// when piped.
		//
		// This is never returned for, e.g., filesystem errors so
		// there's nothing we can do but ignore it and continue.
		syscall.EBADF:
		return true
	}
	return false
}
