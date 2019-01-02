package main

import "path/filepath"

func logPath(base, file string) (logPath string) {
	if base == "" {
		logPath = filepath.Join(base, file)
	} else {
		logPath = filepath.Join(base, file)
	}
	return
}
