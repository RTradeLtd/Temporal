package cli

import (
	"os"
)

func isSyncNotSupportedErr(err error) bool {
	perr, ok := err.(*os.PathError)
	if !ok {
		return false
	}
	return perr.Op == "sync" && isErrnoNotSupported(perr.Err)
}
