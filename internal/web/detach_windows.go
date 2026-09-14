//go:build windows

package web

import (
	"os/exec"
	"syscall"
)

func detachService(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x00000008 | 0x00000200}
}
