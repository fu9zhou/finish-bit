package extension

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestRemoveRetriesOnlyTransientWindowsErrors(t *testing.T) {
	for _, test := range []struct {
		name           string
		windows        bool
		err            error
		failures, want int
	}{
		{"short sharing lock", true, &os.PathError{Op: "remove", Path: "runner.exe", Err: syscall.Errno(32)}, 2, 3},
		{"access denied", true, os.ErrPermission, 1, 2},
		{"persistent lock", true, syscall.Errno(33), 100, 8},
		{"non-Windows permission", false, os.ErrPermission, 100, 1},
		{"unrelated error", true, os.ErrInvalid, 100, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls, sleeps := 0, 0
			err := removeWithRetry("fixture", func(string) error {
				calls++
				if calls <= test.failures {
					return test.err
				}
				return nil
			}, func(time.Duration) { sleeps++ }, test.windows)
			if calls != test.want || sleeps != calls-1 {
				t.Fatalf("calls=%d sleeps=%d", calls, sleeps)
			}
			if test.failures >= test.want {
				if !errors.Is(err, test.err) {
					t.Fatalf("lost error: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
