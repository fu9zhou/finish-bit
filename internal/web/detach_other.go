//go:build !windows && !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package web

import "os/exec"

func detachService(command *exec.Cmd) {}
