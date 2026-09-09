package extension

import (
	"errors"
	"os"
	"runtime"
	"syscall"
	"time"
)

func removeDirectory(path string) error {
	return removeWithRetry(path, os.RemoveAll, time.Sleep, runtime.GOOS == "windows")
}

// A recently exited process or a scanner can briefly hold a Windows file open.
// Retry only sharing/permission failures, and retain the final OS error.
func removeWithRetry(path string, remove func(string) error, sleep func(time.Duration), windows bool) error {
	var err error
	for attempt := 0; attempt < 8; attempt++ {
		err = remove(path)
		retryable := os.IsPermission(err) || errors.Is(err, syscall.Errno(32)) || errors.Is(err, syscall.Errno(33))
		if err == nil || !windows || !retryable || attempt == 7 {
			return err
		}
		sleep(time.Duration(min(attempt+1, 4)) * 100 * time.Millisecond)
	}
	return err
}
