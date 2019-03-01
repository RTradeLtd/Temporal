// +build !windows

package iptbutil

import (
	"os/exec"
	"syscall"
)

func SetupOpt(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
