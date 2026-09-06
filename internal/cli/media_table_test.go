package cli

import (
	"context"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCapabilitiesThroughStructuredRequests(t *testing.T) {
	directory := t.TempDir()
	input := filepath.Join(directory, "image.png")
	file, err := os.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	png.Encode(file, image.NewRGBA(image.Rect(0, 0, 4, 2)))
	file.Close()
	output := filepath.Join(directory, "resized.png")
	request := map[string]any{"operation": "image.resize", "inputs": []string{input}, "options": map[string]any{"width": 2, "output": output}}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "request.json")
	os.WriteFile(path, data, 0600)
	c, out, errout := newTestCLI(t)
	if code := c.Run(context.Background(), []string{"run", "--request", path, "--json"}); code != 0 {
		t.Fatalf("code=%d err=%s", code, errout)
	}
	var result struct {
		Data struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Data.Width != 2 || result.Data.Height != 1 {
		t.Fatalf("output=%s error=%v", out, err)
	}
	out.Reset()
	errout.Reset()
	if code := c.Run(context.Background(), []string{"json", "to-csv", `[{"b":"2","a":"001"}]`, "--columns", "b", "--columns", "a"}); code != 0 || out.String() != "b,a\n2,001\n" {
		t.Fatalf("code=%d out=%s err=%s", code, out, errout)
	}
	out.Reset()
	errout.Reset()
	if code := c.Run(context.Background(), []string{"search", "图片缩放", "--json"}); code != 0 || !strings.Contains(out.String(), "image.resize") {
		t.Fatalf("code=%d out=%s", code, out)
	}
}
