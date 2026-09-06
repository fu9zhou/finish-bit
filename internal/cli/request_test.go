package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStructuredRequestFileAndStdin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "request with spaces.json")
	body := `{"operation":"text.replace","inputs":["a\nb","a","--json"],"options":{"count":1}}`
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{path, "-"} {
		t.Run(source, func(t *testing.T) {
			if source == "-" {
				file, err := os.Open(path)
				if err != nil {
					t.Fatal(err)
				}
				original := os.Stdin
				os.Stdin = file
				defer func() { os.Stdin = original; file.Close() }()
			}
			c, out, errout := newTestCLI(t)
			if code := c.Run(context.Background(), []string{"run", "--request", source, "--json"}); code != 0 || !strings.Contains(out.String(), `--json\nb`) {
				t.Fatalf("code=%d out=%s err=%s", code, out, errout)
			}
		})
	}
}

func TestStructuredRequestErrors(t *testing.T) {
	for _, body := range []string{`{"operation":"text.count","inputs":["x"],"options":{"unknown":1}}`, `{"operation":"uuid.generate","options":{"count":1.5}}`, `{"operation":"missing.operation"}`, `{"operation":"text.count"}`, `{"operation":"text.count","inputs":["x"]} {}`} {
		path := filepath.Join(t.TempDir(), "request.json")
		os.WriteFile(path, []byte(body), 0600)
		c, _, errout := newTestCLI(t)
		if code := c.Run(context.Background(), []string{"run", "--request=" + path, "--json"}); code != 2 || !strings.Contains(errout.String(), `"error"`) {
			t.Fatalf("code=%d error=%s", code, errout)
		}
	}
	c, _, _ := newTestCLI(t)
	if code := c.Run(context.Background(), []string{"run", "--request", "missing", "extra"}); code != 2 {
		t.Fatal(code)
	}
}
