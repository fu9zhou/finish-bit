// Package packagemanager installs pinned, integrity-checked runtime dependencies.
package packagemanager

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

//go:embed registry.json
var builtinRegistry []byte

type Registry struct {
	Schema   int       `json:"schema"`
	Packages []Package `json:"packages"`
}

type Package struct {
	Name        string              `json:"name"`
	Version     string              `json:"version"`
	Description string              `json:"description"`
	License     string              `json:"license"`
	Source      string              `json:"source"`
	Artifacts   map[string]Artifact `json:"artifacts"`
}

type Artifact struct {
	URL         string             `json:"url"`
	Mirrors     []string           `json:"mirrors,omitempty"`
	SHA256      string             `json:"sha256"`
	Format      string             `json:"format"`
	Executables map[string]string  `json:"executables"`
	Resources   []string           `json:"resources,omitempty"`
	Extractor   string             `json:"extractor,omitempty"`
	Downloads   []ResourceDownload `json:"downloads,omitempty"`
}

// ResourceDownload is a pinned supplementary data file, never an executable installer.
type ResourceDownload struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Path   string `json:"path"`
}

func BuiltinRegistry() (Registry, error) {
	var registry Registry
	if err := json.Unmarshal(builtinRegistry, &registry); err != nil {
		return Registry{}, fmt.Errorf("decode built-in package registry: %w", err)
	}
	if registry.Schema != 1 {
		return Registry{}, fmt.Errorf("unsupported package registry schema %d", registry.Schema)
	}
	seen := map[string]bool{}
	for _, pkg := range registry.Packages {
		if pkg.Name == "" || strings.Trim(pkg.Name, "abcdefghijklmnopqrstuvwxyz0123456789-") != "" || pkg.Version == "" || strings.Trim(pkg.Version, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-") != "" || seen[pkg.Name] {
			return Registry{}, fmt.Errorf("package registry contains an invalid or duplicate package %q", pkg.Name)
		}
		seen[pkg.Name] = true
		if parsed, err := url.Parse(pkg.Source); err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return Registry{}, fmt.Errorf("package %q has an invalid source URL", pkg.Name)
		}
		if len(pkg.Artifacts) == 0 {
			return Registry{}, fmt.Errorf("package %q has no artifacts", pkg.Name)
		}
		for platform, artifact := range pkg.Artifacts {
			if platform == "" {
				return Registry{}, fmt.Errorf("package %q has an empty platform", pkg.Name)
			}
			if err := validateArtifact(artifact); err != nil {
				return Registry{}, fmt.Errorf("package %q platform %q: %w", pkg.Name, platform, err)
			}
		}
	}
	return registry, nil
}

func (r Registry) Find(name string) (Package, bool) {
	for _, pkg := range r.Packages {
		if pkg.Name == name {
			return pkg, true
		}
	}
	return Package{}, false
}
