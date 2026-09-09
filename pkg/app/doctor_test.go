package app

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/packagemanager"
)

func TestDoctorDetectsDamagedRuntime(t *testing.T) {
	for _, damaged := range []string{"engine.exe", "support.dll", "tessdata/eng.traineddata", "package.json"} {
		t.Run(damaged, func(t *testing.T) {
			root := t.TempDir()
			var payload bytes.Buffer
			archive := zip.NewWriter(&payload)
			for _, name := range []string{"engine.exe", "support.dll", "tessdata/eng.traineddata"} {
				w, _ := archive.Create(name)
				_, _ = w.Write([]byte("original " + name))
			}
			if err := archive.Close(); err != nil {
				t.Fatal(err)
			}
			hash := sha256.Sum256(payload.Bytes())
			digest := hex.EncodeToString(hash[:])
			manager := packagemanager.New(root, packagemanager.Registry{Schema: 1, Packages: []packagemanager.Package{{Name: "tesseract", Version: "1", Artifacts: map[string]packagemanager.Artifact{packagemanager.PlatformKey(): {URL: "https://example.com/runtime.zip", SHA256: digest, Format: "zip", Executables: map[string]string{"tesseract": "engine.exe"}}}}}})
			if err := os.MkdirAll(filepath.Join(root, "cache"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "cache", digest), payload.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := manager.Install(context.Background(), "tesseract"); err != nil {
				t.Fatal(err)
			}
			a, err := New(Config{Root: root})
			if err != nil {
				t.Fatal(err)
			}
			a.packages = manager
			check := func() bool {
				t.Helper()
				for _, c := range a.Doctor() {
					if c.Name == "package:tesseract" {
						return c.OK
					}
				}
				t.Fatal("missing tesseract check")
				return false
			}
			if !check() {
				t.Fatal("healthy package rejected")
			}
			target := filepath.Join(root, "packages", "tesseract", "1", "payload", filepath.FromSlash(damaged))
			if damaged == "package.json" {
				target = filepath.Join(root, "packages", "tesseract", "1", damaged)
			}
			if damaged == "tessdata/eng.traineddata" {
				err = os.Remove(target)
			} else {
				err = os.WriteFile(target, []byte("corrupt"), 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			if check() {
				t.Fatal("doctor reported damaged runtime ready")
			}
		})
	}
}
