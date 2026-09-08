package localtools

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func TestAdditionalBoundaries(t *testing.T) {
	call(t, "useragent.parse", []string{"Chrome/"}, nil)
	r := call(t, "cron.next", []string{"0 9 * * 1-5"}, map[string]any{"after": "2024-01-05T10:00:00Z", "count": 1})
	if r.Data["times"].([]string)[0] != "2024-01-08T09:00:00Z" {
		t.Fatal(r)
	}
	r = call(t, "toml.to-json", []string{"count = 9223372036854775807"}, nil)
	if !strings.Contains(r.Data["text"].(string), "9223372036854775807") {
		t.Fatal(r)
	}
	r = call(t, "text.hide", []string{"cover", "secret"}, nil)
	x := r.Data["text"].(string)
	pos := strings.LastIndex(x, "\u200b")
	x = x[:pos] + "\u200c" + x[pos+3:]
	reg := operation.NewRegistry()
	_ = Register(reg)
	c, _ := reg.Get("text.reveal")
	if _, err := c.Runner.Run(context.Background(), operation.Request{Inputs: []string{x}}); err == nil {
		t.Fatal("corrupt hidden payload accepted")
	}
}

func TestDocumentCompressionPreservesPartsAndRejectsUnsafePackages(t *testing.T) {
	for _, bad := range []string{"", "../escape", "_xmlsignatures/sig1.xml"} {
		t.Run(bad, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "input.docx")
			output := filepath.Join(dir, "output.docx")
			f, err := os.Create(input)
			if err != nil {
				t.Fatal(err)
			}
			z := zip.NewWriter(f)
			parts := map[string]string{"[Content_Types].xml": "<Types/>", "word/document.xml": strings.Repeat("<p>keep me</p>", 100)}
			if bad != "" {
				parts[bad] = "untrusted"
			}
			for name, body := range parts {
				w, err := z.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
				if err != nil {
					t.Fatal(err)
				}
				_, _ = w.Write([]byte(body))
			}
			if err := z.Close(); err != nil {
				t.Fatal(err)
			}
			_ = f.Close()
			reg := operation.NewRegistry()
			_ = Register(reg)
			c, _ := reg.Get("document.compress")
			_, err = c.Runner.Run(context.Background(), operation.Request{Inputs: []string{input}, Options: map[string]any{"output": output}})
			if bad != "" {
				if err == nil {
					t.Fatal("unsafe package accepted")
				}
				if _, err := os.Stat(output); !os.IsNotExist(err) {
					t.Fatal("partial output published")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			zr, err := zip.OpenReader(output)
			if err != nil {
				t.Fatal(err)
			}
			defer zr.Close()
			if len(zr.File) != len(parts) {
				t.Fatal("package members lost")
			}
			for _, member := range zr.File {
				rc, err := member.Open()
				if err != nil {
					t.Fatal(err)
				}
				var b strings.Builder
				_, err = io.Copy(&b, rc)
				_ = rc.Close()
				if err != nil || b.String() != parts[member.Name] {
					t.Fatal("part changed", member.Name)
				}
			}
		})
	}
}
