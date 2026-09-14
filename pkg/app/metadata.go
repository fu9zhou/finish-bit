package app

import (
	_ "embed"
	"encoding/json"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

// Legacy built-ins predate path/choice discovery metadata. Keep their explicit
// compatibility records here, keyed by operation and parameter (never parse prose
// in an adapter). New providers/extensions can declare Kind and Choices directly.
//
//go:embed catalog-metadata.json
var metadataJSON []byte

type parameterMetadata struct {
	Kind    string   `json:"kind,omitempty"`
	Choices []string `json:"choices,omitempty"`
}

var catalogMetadata = func() map[string]map[string]parameterMetadata {
	var metadata map[string]map[string]parameterMetadata
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		panic(err)
	}
	return metadata
}()

func describeMetadata(def operation.Definition) operation.Definition {
	metadata := catalogMetadata[def.ID]
	decorate := func(params []operation.Parameter) []operation.Parameter {
		if params == nil {
			return nil
		}
		copy := append([]operation.Parameter{}, params...)
		for i := range copy {
			p := &copy[i]
			m := metadata[p.Name]
			if p.Kind == "" {
				p.Kind = m.Kind
			}
			if len(p.Choices) == 0 {
				p.Choices = append([]string(nil), m.Choices...)
			}
		}
		return copy
	}
	def.Inputs = decorate(def.Inputs)
	def.Options = decorate(def.Options)
	return def
}
