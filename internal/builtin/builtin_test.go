package builtin

import (
	"context"
	"regexp"
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
