package operation

import (
	"context"
	"testing"
)

func TestRegistryRejectsDuplicateOperation(t *testing.T) {
	registry := NewRegistry()
	capability := Capability{Definition: Definition{ID: "test.echo", Summary: "Echo input"}, Runner: Func(func(_ context.Context, _ Request) (Result, error) { return Result{}, nil })}
	if err := registry.Register(capability); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(capability); err == nil {
		t.Fatal("duplicate registration succeeded")
	}
}

func TestValidateDefinitionRejectsInvalidID(t *testing.T) {
	if err := ValidateDefinition(Definition{ID: "Bad ID", Summary: "Bad"}); err == nil {
		t.Fatal("invalid ID was accepted")
	}
}
