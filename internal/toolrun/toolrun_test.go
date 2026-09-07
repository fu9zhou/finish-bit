package toolrun

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestArtifactPreservesDestinationUntilSuccess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.dat")
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("tool failed")
	if err := Artifact(context.Background(), path, true, func(temporary string) error {
		if err := os.WriteFile(temporary, []byte("partial"), 0600); err != nil {
			return err
		}
		return failure
	}); !errors.Is(err, failure) {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "original" {
		t.Fatal("failed generator modified destination")
	}
	if err := Artifact(context.Background(), path, false, func(string) error { t.Fatal("must not invoke generator for existing destination"); return nil }); err == nil {
		t.Fatal("overwrite should require opt-in")
	}
	if err := Artifact(context.Background(), path, true, func(temporary string) error { return os.WriteFile(temporary, []byte("complete"), 0600) }); err != nil {
		t.Fatal(err)
	}
	got, _ = os.ReadFile(path)
	if string(got) != "complete" {
		t.Fatal("successful generator did not publish")
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("temporary files leaked: %v", entries)
	}
}

func TestDirectoryRejectsExistingAndCleansFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "frames")
	_, err := Directory(context.Background(), path, func(temp string) error {
		_ = os.WriteFile(filepath.Join(temp, "frame.png"), []byte("partial"), 0600)
		return errors.New("decode failed")
	})
	if err == nil {
		t.Fatal("expected failure")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("failed output was published")
	}
	outputs, err := Directory(context.Background(), path, func(temp string) error {
		return os.WriteFile(filepath.Join(temp, "frame.png"), []byte("complete"), 0600)
	})
	if err != nil || len(outputs) != 1 {
		t.Fatalf("outputs=%v error=%v", outputs, err)
	}
	_, err = Directory(context.Background(), path, func(string) error { t.Fatal("must not generate into existing directory"); return nil })
	if err == nil {
		t.Fatal("expected collision")
	}
}

func TestLocalFileRejectsProtocolsAndDirectories(t *testing.T) {
	for _, path := range []string{"https://example.com/video.mp4", t.TempDir(), "not-present-file"} {
		if _, err := LocalFile(path); err == nil {
			t.Fatalf("accepted %q", path)
		}
	}
}
