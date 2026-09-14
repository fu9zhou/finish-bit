//go:build !windows && !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package packagemanager

import "fmt"

func packageLockBusy(err error) bool { return false }

func lockPackageFile(path string) (func(), error) {
	return nil, fmt.Errorf("package locking is unsupported on this platform")
}
