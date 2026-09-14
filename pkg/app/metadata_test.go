package app

import (
	"fmt"
	"slices"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func TestCatalogMetadataMatchesDeclaredParameters(t *testing.T) {
	a, err := New(Config{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for id, metadata := range catalogMetadata {
		def, err := a.Describe(id)
		if err != nil {
			t.Fatal(err)
		}
		parameters := map[string]operation.Parameter{}
		for _, p := range append(def.Inputs, def.Options...) {
			parameters[p.Name] = p
		}
		for name, m := range metadata {
			p, ok := parameters[name]
			if !ok {
				t.Errorf("stale metadata %s.%s", id, name)
				continue
			}
			if m.Kind != "" && !slices.Contains([]string{"file", "directory", "path", "output-file", "output-directory", "text-or-file"}, m.Kind) {
				t.Errorf("invalid kind %s.%s", id, name)
			}
			if len(m.Choices) > 0 && p.Default != nil && !slices.Contains(m.Choices, fmt.Sprint(p.Default)) {
				t.Errorf("default absent from choices %s.%s", id, name)
			}
		}
	}
	extract, _ := a.Describe("archive.extract")
	if extract.Inputs[0].Kind != "file" {
		t.Fatal("archive must be a file picker")
	}
	for _, p := range extract.Options {
		if p.Name == "entries" && p.Kind != "" {
			t.Fatal("archive members are not local files")
		}
		if p.Name == "output" && p.Kind != "output-directory" {
			t.Fatal("extraction creates a directory")
		}
		if p.Name == "format" && len(p.Choices) == 0 {
			t.Fatal("format choices missing")
		}
	}
	def, _ := a.Describe("json.query")
	if def.Inputs[1].Kind != "" {
		t.Fatal("JSON path is not a filesystem path")
	}
	// An adapter must not mutate a subsequent caller's discovery data.
	extract.Inputs[0].Kind = "changed"
	again, _ := a.Describe("archive.extract")
	if again.Inputs[0].Kind != "file" {
		t.Fatal("metadata leaked mutable state")
	}
}
