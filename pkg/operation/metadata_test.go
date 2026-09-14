package operation

import "testing"

func TestParameterDiscoveryMetadata(t *testing.T) {
	base := Definition{ID: "test.example", Summary: "Test", Source: "test"}
	for _, p := range []Parameter{
		{Name: "file", Type: TypeString, Kind: "file"},
		{Name: "files", Type: TypeStrings, Kind: "path"},
		{Name: "format", Type: TypeString, Choices: []string{"zip", "7z"}},
	} {
		base.Inputs = []Parameter{p}
		if err := ValidateDefinition(base); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range []Parameter{
		{Name: "file", Type: TypeString, Kind: "unknown"},
		{Name: "file", Type: TypeInteger, Kind: "file"},
		{Name: "mode", Type: TypeString, Choices: []string{"zip", "zip"}},
		{Name: "mode", Type: TypeBoolean, Choices: []string{"yes"}},
	} {
		base.Inputs = []Parameter{p}
		if err := ValidateDefinition(base); err == nil {
			t.Fatalf("accepted invalid metadata: %+v", p)
		}
	}
}
