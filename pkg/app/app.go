// Package app is the transport-independent FinishBit application service.
// CLI and future HTTP adapters must call this package instead of duplicating behavior.
package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fu9zhou/finish-bit/internal/builtin"
	ffmpegprovider "github.com/fu9zhou/finish-bit/internal/ffmpeg"
	"github.com/fu9zhou/finish-bit/pkg/extension"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"github.com/fu9zhou/finish-bit/pkg/packagemanager"
	searchpkg "github.com/fu9zhou/finish-bit/pkg/search"
)

type App struct {
	registry   *operation.Registry
	packages   *packagemanager.Manager
	extensions *extension.Manager
}

type Config struct{ Root string }

func New(config Config) (*App, error) {
	root := config.Root
	if root == "" {
		var err error
		root, err = packagemanager.DefaultRoot()
		if err != nil {
			return nil, err
		}
	}
	packageRegistry, err := packagemanager.BuiltinRegistry()
	if err != nil {
		return nil, err
	}
	packages := packagemanager.New(root, packageRegistry)
	extensions := extension.New(filepath.Join(root, "extensions"))
	registry := operation.NewRegistry()
	if err := builtin.Register(registry); err != nil {
		return nil, err
	}
	if err := ffmpegprovider.Register(registry, packages); err != nil {
		return nil, err
	}
	if err := extensions.Register(registry); err != nil {
		return nil, err
	}
	return &App{registry: registry, packages: packages, extensions: extensions}, nil
}

func (a *App) Search(query string, limit int) []searchpkg.Match {
	return searchpkg.New(a.registry.Definitions()).Search(query, limit)
}

func (a *App) Describe(id string) (operation.Definition, error) {
	capability, ok := a.registry.Get(id)
	if !ok {
		return operation.Definition{}, &operation.Error{Code: operation.CodeNotFound, Message: fmt.Sprintf("operation %q was not found", id), Suggestion: fmt.Sprintf("fnsh search %q", id)}
	}
	return capability.Definition, nil
}

func (a *App) Capabilities() []operation.Definition { return a.registry.Definitions() }

func (a *App) Execute(ctx context.Context, id string, request operation.Request) (operation.Result, error) {
	capability, ok := a.registry.Get(id)
	if !ok {
		return operation.Result{}, &operation.Error{Code: operation.CodeNotFound, Message: fmt.Sprintf("operation %q was not found", id), Suggestion: fmt.Sprintf("fnsh search %q", id)}
	}
	result, err := capability.Runner.Run(ctx, request)
	if err != nil {
		typed := operation.AsError(err)
		if typed.Operation == "" {
			typed.Operation = id
		}
		return operation.Result{}, typed
	}
	result.Operation = id
	return result, nil
}

func (a *App) InstallPackage(ctx context.Context, name string) (packagemanager.Installed, error) {
	return a.packages.Install(ctx, name)
}
func (a *App) RepairPackage(ctx context.Context, name string) (packagemanager.Installed, error) {
	return a.packages.Repair(ctx, name)
}
func (a *App) RemovePackage(name string) error { return a.packages.Remove(name) }
func (a *App) PackageInfo(name string) (packagemanager.Status, error) {
	return a.packages.Status(name)
}
func (a *App) Packages() []packagemanager.Installed { return a.packages.List() }

func (a *App) InstallExtension(source string) (extension.Manifest, error) {
	return a.extensions.InstallValidated(source, func(manifest extension.Manifest) error {
		seen := make(map[string]bool, len(manifest.Operations))
		for _, definition := range manifest.Operations {
			if seen[definition.ID] {
				return &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("operation %q already registered by extension %q", definition.ID, manifest.Name)}
			}
			seen[definition.ID] = true
			if _, exists := a.registry.Get(definition.ID); exists {
				return &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("operation %q already registered", definition.ID)}
			}
		}
		return nil
	})
}
func (a *App) RemoveExtension(name string) error                     { return a.extensions.Remove(name) }
func (a *App) ExtensionInfo(name string) (extension.Manifest, error) { return a.extensions.Info(name) }
func (a *App) Extensions() []extension.Manifest                      { return a.extensions.List() }

type Check struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (a *App) Doctor() []Check {
	checks := []Check{{Name: "runtime", OK: true, Message: "operation registry loaded"}}
	root := a.packages.Root()
	err := os.MkdirAll(root, 0o755)
	checks = append(checks, Check{Name: "data-directory", OK: err == nil, Message: root})
	installedPackages := map[string]bool{}
	for _, installed := range a.packages.List() {
		installedPackages[installed.Name] = true
	}
	for _, definition := range a.registry.Definitions() {
		for _, requirement := range definition.Requirements {
			if !installedPackages[requirement.Package] {
				checks = append(checks, Check{Name: "package:" + requirement.Package, OK: true, Message: "not installed (optional until an operation requires it)"})
				continue
			}
			_, dependencyErr := a.packages.Executable(requirement.Package, requirement.Package)
			checks = append(checks, Check{Name: "package:" + requirement.Package, OK: dependencyErr == nil, Message: statusMessage(dependencyErr)})
		}
	}
	unique := map[string]Check{}
	for _, check := range checks {
		unique[check.Name] = check
	}
	checks = checks[:0]
	for _, check := range unique {
		checks = append(checks, check)
	}
	sort.Slice(checks, func(i, j int) bool { return checks[i].Name < checks[j].Name })
	return checks
}

func statusMessage(err error) string {
	if err == nil {
		return "ready"
	}
	return err.Error()
}
