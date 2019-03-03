//+build windows

package cli

import (
	"syscall"
)

const invalid_file_handle syscall.Errno = 0x6

func isErrnoNotSupported(err error) bool {
	switch err {
	case syscall.EINVAL, syscall.ENOTSUP, syscall.ENOTTY, invalid_file_handle:
		return true
	}
	return false
}
