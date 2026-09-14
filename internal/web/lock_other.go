//go:build !windows && !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package web

import "fmt"

func lockServiceFile(path string) (func(), error) {
	return nil, fmt.Errorf("package locking is unsupported on this platform")
}
