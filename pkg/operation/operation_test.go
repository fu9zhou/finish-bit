package operation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestRegistryRejectsDuplicateOperation(t *testing.T) {
	registry := NewRegistry()
	capability := Capability{Definition: Definition{ID: "test.echo", Summary: "Echo input", Source: "test"}, Runner: Func(func(_ context.Context, _ Request) (Result, error) { return Result{}, nil })}
	if err := registry.Register(capability); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(capability); err == nil {
		t.Fatal("duplicate registration succeeded")
	}
}

func TestRegistryReturnsDefinitionsSortedAndRunsCapability(t *testing.T) {
	registry := NewRegistry()
	for _, id := range []string{"test.zeta", "test.alpha"} {
		id := id
		capability := Capability{Definition: Definition{ID: id, Summary: "Test", Source: "test"}, Runner: Func(func(_ context.Context, _ Request) (Result, error) {
			return Result{Operation: id}, nil
		})}
		if err := registry.Register(capability); err != nil {
			t.Fatal(err)
		}
	}
	definitions := registry.Definitions()
	if len(definitions) != 2 || definitions[0].ID != "test.alpha" || definitions[1].ID != "test.zeta" {
		t.Fatalf("definitions = %#v", definitions)
	}
	capability, ok := registry.Get("test.alpha")
	if !ok {
		t.Fatal("registered capability not found")
	}
	result, err := capability.Runner.Run(context.Background(), Request{})
	if err != nil || result.Operation != "test.alpha" {
		t.Fatalf("Run() = %#v, %v", result, err)
	}
}

func TestRegistryRejectsMissingRunner(t *testing.T) {
	err := NewRegistry().Register(Capability{Definition: Definition{ID: "test.run", Summary: "Run", Source: "test"}})
	if err == nil {
		t.Fatal("capability without runner was accepted")
	}
}

func TestValidateDefinitionRejectsInvalidID(t *testing.T) {
	if err := ValidateDefinition(Definition{ID: "Bad ID", Summary: "Bad"}); err == nil {
		t.Fatal("invalid ID was accepted")
	}
}

func TestValidateDefinitionRejectsUnsupportedParameterType(t *testing.T) {
	definition := Definition{ID: "test.run", Summary: "Run test", Source: "test", Options: []Parameter{{Name: "value", Type: "number", Description: "Value"}}}
	if err := ValidateDefinition(definition); err == nil {
		t.Fatal("unsupported parameter type was accepted")
	}
}

func TestValidateDefinitionRequiresSource(t *testing.T) {
	if err := ValidateDefinition(Definition{ID: "test.run", Summary: "Run test"}); err == nil {
		t.Fatal("missing source was accepted")
	}
}

func TestValidateDefinitionRejectsDuplicateMetadataAndEmptyRequirement(t *testing.T) {
	tests := []Definition{
		{ID: "test.run", Summary: "Run", Source: "test", Aliases: []string{"run", "run"}},
		{ID: "test.run", Summary: "Run", Source: "test", Tags: []string{"test", "test"}},
		{ID: "test.run", Summary: "Run", Source: "test", Requirements: []Requirement{{Package: ""}}},
	}
	for _, definition := range tests {
		if err := ValidateDefinition(definition); err == nil {
			t.Fatalf("invalid definition was accepted: %#v", definition)
		}
	}
}

func TestIntOptionRejectsFractionalJSONNumber(t *testing.T) {
	request := Request{Options: map[string]any{"count": 1.9}}
	if _, err := IntOption(request, "count", 0); err == nil {
		t.Fatal("fractional integer option was accepted")
	}
}

func TestOptionReadersAcceptSupportedRepresentations(t *testing.T) {
	request := Request{Options: map[string]any{"text": "value", "bool": "true", "int": "42", "float-int": float64(7)}}
	if value, err := StringOption(request, "text", ""); err != nil || value != "value" {
		t.Fatalf("StringOption() = %q, %v", value, err)
	}
	if value, err := BoolOption(request, "bool", false); err != nil || !value {
		t.Fatalf("BoolOption() = %v, %v", value, err)
	}
	if value, err := IntOption(request, "int", 0); err != nil || value != 42 {
		t.Fatalf("IntOption(string) = %d, %v", value, err)
	}
	if value, err := IntOption(request, "float-int", 0); err != nil || value != 7 {
		t.Fatalf("IntOption(float64) = %d, %v", value, err)
	}
	if value, err := StringOption(request, "missing", "fallback"); err != nil || value != "fallback" {
		t.Fatalf("StringOption(fallback) = %q, %v", value, err)
	}
	if err := RequireInputs(Request{Inputs: []string{"one"}}, 2); err == nil {
		t.Fatal("missing inputs were accepted")
	}
}

func TestOptionReadersRejectWrongTypes(t *testing.T) {
	tests := []struct {
		name string
		read func() error
	}{
		{"string", func() error { _, err := StringOption(Request{Options: map[string]any{"x": true}}, "x", ""); return err }},
		{"boolean", func() error {
			_, err := BoolOption(Request{Options: map[string]any{"x": "not-bool"}}, "x", false)
			return err
		}},
		{"integer", func() error {
			_, err := IntOption(Request{Options: map[string]any{"x": "not-int"}}, "x", 0)
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.read(); err == nil {
				t.Fatal("invalid option was accepted")
			}
		})
	}
}

func TestStructuredErrorsAndOptionDecoding(t *testing.T) {
	cause := errors.New("cause")
	typed := &Error{Code: CodeInvalidInput, Message: "bad input", Err: cause}
	if typed.Error() != "bad input" || !errors.Is(typed, cause) || AsError(typed) != typed {
		t.Fatalf("typed error behavior is incorrect: %#v", typed)
	}
	wrapped := AsError(errors.New("boom"))
	if wrapped.Code != CodeExecutionFailed || wrapped.Message != "boom" {
		t.Fatalf("AsError() = %#v", wrapped)
	}
	options, err := DecodeOptions(json.RawMessage(`{"count":2}`))
	if err != nil || options["count"] != float64(2) {
		t.Fatalf("DecodeOptions() = %#v, %v", options, err)
	}
	if empty, err := DecodeOptions(nil); err != nil || len(empty) != 0 {
		t.Fatalf("DecodeOptions(nil) = %#v, %v", empty, err)
	}
	if _, err := DecodeOptions(json.RawMessage(`[]`)); err == nil || !strings.Contains(err.Error(), "JSON object") {
		t.Fatalf("invalid DecodeOptions() error = %v", err)
	}
}
