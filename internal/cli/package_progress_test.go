package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/app"
	"github.com/fu9zhou/finish-bit/pkg/packagemanager"
)

func TestPackageProgressPreservesJSONOutput(t *testing.T) {
	registry, err := packagemanager.BuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	pkg, _ := registry.Find("ffmpeg")
	artifact, ok := pkg.Artifacts[packagemanager.PlatformKey()]
	if !ok {
		t.Skip("no ffmpeg for this platform")
	}
	root := t.TempDir()
	directory := filepath.Join(root, "packages", pkg.Name, pkg.Version)
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	executable := "ffmpeg"
	if packagemanager.PlatformKey() == "win32-x64" || packagemanager.PlatformKey() == "win32-arm64" {
		executable += ".exe"
	}
	_ = artifact
	data := []byte("local installed fixture")
	digest := sha256.Sum256(data)
	if err := os.WriteFile(filepath.Join(directory, executable), data, 0755); err != nil {
		t.Fatal(err)
	}
	installed := packagemanager.Installed{Name: pkg.Name, Version: pkg.Version, Platform: packagemanager.PlatformKey(), Executables: map[string]string{"ffmpeg": executable}, Files: map[string]string{executable: hex.EncodeToString(digest[:])}}
	metadata, _ := json.Marshal(installed)
	if err := os.WriteFile(filepath.Join(directory, "package.json"), metadata, 0600); err != nil {
		t.Fatal(err)
	}
	a, err := app.New(app.Config{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	for _, machine := range []bool{false, true} {
		var stdout, stderr bytes.Buffer
		args := []string{"pkg", "add", "ffmpeg"}
		if machine {
			args = append(args, "--json")
		}
		if code := New(a, &stdout, &stderr, "test").Run(context.Background(), args); code != 0 {
			t.Fatalf("code=%d stderr=%s", code, stderr.String())
		}
		if machine {
			if stderr.Len() != 0 || !json.Valid(stdout.Bytes()) {
				t.Fatalf("mixed JSON/progress: %s %s", stdout.String(), stderr.String())
			}
		} else if !bytes.Contains(stderr.Bytes(), []byte("installed files verified")) {
			t.Fatal("missing human progress")
		}
	}
}
