//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package web

import (
	"os/exec"
	"syscall"
)

func detachService(command *exec.Cmd) { command.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }
