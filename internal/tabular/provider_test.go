package tabular

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func TestTableInputBoundaries(t *testing.T) {
	for _, body := range []string{"", "a,a\n1,2\n", "a,\n1,2\n", "a,b\n1\n", "a\n\"unclosed\n"} {
		if _, e := readTable([]byte(body), ','); e == nil {
			t.Errorf("accepted malformed table %q", body)
		}
	}
	for _, body := range []string{"{}", "[]", "{\"a\":1}\n{\"b\":2}", "{\"a\":{\"x\":1}}", "{\"a\":[]}"} {
		if e := checkJSONL([]byte(body)); e == nil {
			t.Errorf("accepted malformed records %q", body)
		}
	}
	for _, body := range []string{`{"$ref":"https://example.com/a"}`, `{"$defs":{"x":{"dynamicEnum":"file:///private"}}}`, `{"$ref":"local.json"}`} {
		if e := localSchema([]byte(body)); e == nil {
			t.Errorf("accepted external schema %s", body)
		}
	}
	if e := localSchema([]byte(`{"$ref":"#/$defs/local","$defs":{"local":{"type":"string"}}}`)); e != nil {
		t.Fatal(e)
	}
}

func TestInvalidRequestsNeverStartRuntime(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "input.csv")
	if e := os.WriteFile(input, []byte("key,value\nA,1\na,2\n"), 0600); e != nil {
		t.Fatal(e)
	}
	registry := operation.NewRegistry()
	if e := Register(registry, runtimeResolver(filepath.Join(root, "missing-qsv"))); e != nil {
		t.Fatal(e)
	}
	for _, test := range []struct {
		id   string
		opts map[string]any
	}{
		{"csv.split", map[string]any{"rows": 0}},
		{"csv.partition", map[string]any{"column": "key"}},
		{"csv.select", map[string]any{"columns": []string{"missing"}}},
		{"csv.rename", map[string]any{"names": []string{"only-one"}}},
		{"csv.sample", map[string]any{"count": 3}},
		{"csv.melt", map[string]any{"columns": []string{"key", "value"}}},
	} {
		test.opts["output"] = filepath.Join(root, test.id)
		cap, _ := registry.Get(test.id)
		_, e := cap.Runner.Run(context.Background(), operation.Request{Inputs: []string{input}, Options: test.opts})
		if e == nil || operation.AsError(e).Code != operation.CodeInvalidInput {
			t.Errorf("%s: expected validation, got %v", test.id, e)
		}
	}
}
