package builtin

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func testRegistry(t *testing.T) *operation.Registry {
	t.Helper()
	registry := operation.NewRegistry()
	if err := Register(registry); err != nil {
		t.Fatal(err)
	}
	return registry
}

func runTestOperation(t *testing.T, id string, request operation.Request) operation.Result {
	t.Helper()
	capability, ok := testRegistry(t).Get(id)
	if !ok {
		t.Fatalf("operation %s is not registered", id)
	}
	if request.Options == nil {
		request.Options = map[string]any{}
	}
	result, err := capability.Runner.Run(context.Background(), request)
	if err != nil {
		t.Fatalf("%s failed: %v", id, err)
	}
	return result
}

func TestJSONRejectsTrailingGarbage(t *testing.T) {
	capability, _ := testRegistry(t).Get("json.validate")
	_, err := capability.Runner.Run(context.Background(), operation.Request{Inputs: []string{`{"ok":true} trailing`}, Options: map[string]any{}})
	if err == nil {
		t.Fatal("trailing JSON content was accepted")
	}
}

func TestTimeAutoDetectsMilliseconds(t *testing.T) {
	capability, _ := testRegistry(t).Get("time.convert")
	result, err := capability.Runner.Run(context.Background(), operation.Request{Inputs: []string{"1704067200000"}, Options: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Data["value"] != "2024-01-01T00:00:00Z" {
		t.Fatalf("value = %#v", result.Data["value"])
	}
}

func TestUUIDFormatAndVersion(t *testing.T) {
	capability, _ := testRegistry(t).Get("uuid.generate")
	result, err := capability.Runner.Run(context.Background(), operation.Request{Options: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	value := result.Data["values"].([]string)[0]
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(value) {
		t.Fatalf("invalid UUID v4 %q", value)
	}
}

func TestJSONQuery(t *testing.T) {
	capability, ok := testRegistry(t).Get("json.query")
	if !ok {
		t.Fatal("json.query not registered")
	}
	result, err := capability.Runner.Run(context.Background(), operation.Request{Inputs: []string{`{"users":[{"name":"Ada"}]}`, "users.0.name"}, Options: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if got := result.Data["text"]; got != `"Ada"`+"\n" {
		t.Fatalf("text = %#v", got)
	}
}

func TestBase64RoundTrip(t *testing.T) {
	registry := testRegistry(t)
	encode, _ := registry.Get("base64.encode")
	decode, _ := registry.Get("base64.decode")
	encoded, err := encode.Runner.Run(context.Background(), operation.Request{Inputs: []string{"FinishBit"}, Options: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decode.Runner.Run(context.Background(), operation.Request{Inputs: []string{encoded.Data["text"].(string)}, Options: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Data["text"] != "FinishBit" {
		t.Fatalf("decoded = %#v", decoded.Data["text"])
	}
}

func TestCoreOperationCatalogHappyPaths(t *testing.T) {
	directory := t.TempDir()
	inputPath := filepath.Join(directory, "input.txt")
	if err := os.WriteFile(inputPath, []byte("beta\nalpha\nbeta\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		id      string
		request operation.Request
		check   func(t *testing.T, result operation.Result)
	}{
		{"hash.calculate", operation.Request{Inputs: []string{"abc"}}, func(t *testing.T, result operation.Result) {
			if result.Data["digest"] != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
				t.Fatalf("digest = %v", result.Data["digest"])
			}
		}},
		{"file.info", operation.Request{Inputs: []string{inputPath}}, func(t *testing.T, result operation.Result) {
			if result.Data["directory"] != false || result.Data["size"] != int64(16) {
				t.Fatalf("file info = %#v", result.Data)
			}
		}},
		{"file.checksum", operation.Request{Inputs: []string{inputPath}}, func(t *testing.T, result operation.Result) {
			if len(result.Data["digest"].(string)) != 64 {
				t.Fatalf("checksum = %#v", result.Data)
			}
		}},
		{"json.format", operation.Request{Inputs: []string{`{"b":2,"a":1}`}, Options: map[string]any{"indent": 2}}, func(t *testing.T, result operation.Result) {
			if !strings.Contains(result.Data["text"].(string), "  \"a\": 1") {
				t.Fatalf("formatted JSON = %q", result.Data["text"])
			}
		}},
		{"json.minify", operation.Request{Inputs: []string{"{ \"ok\": true }"}}, func(t *testing.T, result operation.Result) {
			if result.Data["text"] != `{"ok":true}` {
				t.Fatalf("minified JSON = %q", result.Data["text"])
			}
		}},
		{"json.validate", operation.Request{Inputs: []string{"[1,2]"}}, func(t *testing.T, result operation.Result) {
			if result.Data["type"] != "array" {
				t.Fatalf("JSON type = %v", result.Data["type"])
			}
		}},
		{"text.count", operation.Request{Inputs: []string{inputPath}}, func(t *testing.T, result operation.Result) {
			if result.Data["lines"] != 3 || result.Data["words"] != 3 {
				t.Fatalf("text count = %#v", result.Data)
			}
		}},
		{"text.replace", operation.Request{Inputs: []string{"banana", "a", "X"}, Options: map[string]any{"count": 2}}, func(t *testing.T, result operation.Result) {
			if result.Data["text"] != "bXnXna" {
				t.Fatalf("replace = %q", result.Data["text"])
			}
		}},
		{"text.sort", operation.Request{Inputs: []string{inputPath}, Options: map[string]any{"descending": true}}, func(t *testing.T, result operation.Result) {
			if result.Data["text"] != "beta\nbeta\nalpha" {
				t.Fatalf("sort = %q", result.Data["text"])
			}
		}},
		{"text.unique", operation.Request{Inputs: []string{inputPath}}, func(t *testing.T, result operation.Result) {
			if result.Data["text"] != "beta\nalpha" {
				t.Fatalf("unique = %q", result.Data["text"])
			}
		}},
		{"time.convert", operation.Request{Inputs: []string{"1704067200"}, Options: map[string]any{"unit": "seconds", "format": "unix"}}, func(t *testing.T, result operation.Result) {
			if result.Data["value"] != "1704067200" {
				t.Fatalf("time = %#v", result.Data)
			}
		}},
		{"url.encode", operation.Request{Inputs: []string{"a b+c"}}, func(t *testing.T, result operation.Result) {
			if result.Data["text"] != "a+b%2Bc" {
				t.Fatalf("encoded URL = %q", result.Data["text"])
			}
		}},
		{"url.decode", operation.Request{Inputs: []string{"a+b%2Bc"}}, func(t *testing.T, result operation.Result) {
			if result.Data["text"] != "a b+c" {
				t.Fatalf("decoded URL = %q", result.Data["text"])
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			test.check(t, runTestOperation(t, test.id, test.request))
		})
	}
}

func TestCoreOperationWritesRequestedOutput(t *testing.T) {
	output := filepath.Join(t.TempDir(), "nested", "encoded.txt")
	result := runTestOperation(t, "base64.encode", operation.Request{Inputs: []string{"FinishBit"}, Options: map[string]any{"output": output}})
	if len(result.Outputs) != 1 || result.Outputs[0] != output {
		t.Fatalf("outputs = %#v", result.Outputs)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "RmluaXNoQml0" {
		t.Fatalf("output = %q", data)
	}
}

func TestCoreOperationsRejectInvalidInputs(t *testing.T) {
	registry := testRegistry(t)
	tests := []struct {
		id      string
		request operation.Request
	}{
		{"base64.decode", operation.Request{Inputs: []string{"%%%"}, Options: map[string]any{}}},
		{"hash.calculate", operation.Request{Inputs: []string{"value"}, Options: map[string]any{"algorithm": "crc32"}}},
		{"file.info", operation.Request{Inputs: []string{filepath.Join(t.TempDir(), "missing")}, Options: map[string]any{}}},
		{"json.format", operation.Request{Inputs: []string{"{}"}, Options: map[string]any{"indent": 17}}},
		{"json.query", operation.Request{Inputs: []string{`{"items":[]}`, "items.2"}, Options: map[string]any{}}},
		{"time.convert", operation.Request{Inputs: []string{"123"}, Options: map[string]any{"unit": "minutes"}}},
		{"time.convert", operation.Request{Inputs: []string{"123"}, Options: map[string]any{"timezone": "Not/AZone"}}},
		{"url.decode", operation.Request{Inputs: []string{"%zz"}, Options: map[string]any{}}},
		{"uuid.generate", operation.Request{Options: map[string]any{"count": 1001}}},
	}
	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			capability, ok := registry.Get(test.id)
			if !ok {
				t.Fatalf("operation %s is not registered", test.id)
			}
			if _, err := capability.Runner.Run(context.Background(), test.request); err == nil {
				t.Fatal("invalid input succeeded")
			}
		})
	}
}
