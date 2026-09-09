// Package app is the transport-independent FinishBit application service.
// CLI and future HTTP adapters must call this package instead of duplicating behavior.
package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	archiveprovider "github.com/fu9zhou/finish-bit/internal/archive"
	"github.com/fu9zhou/finish-bit/internal/builtin"
	"github.com/fu9zhou/finish-bit/internal/desktop"
	"github.com/fu9zhou/finish-bit/internal/document"
	ffmpegprovider "github.com/fu9zhou/finish-bit/internal/ffmpeg"
	"github.com/fu9zhou/finish-bit/internal/localtools"
	"github.com/fu9zhou/finish-bit/internal/ocr"
	pdfprovider "github.com/fu9zhou/finish-bit/internal/pdf"
	"github.com/fu9zhou/finish-bit/internal/raster"
	"github.com/fu9zhou/finish-bit/internal/tabular"
	"github.com/fu9zhou/finish-bit/pkg/extension"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"github.com/fu9zhou/finish-bit/pkg/packagemanager"
	searchpkg "github.com/fu9zhou/finish-bit/pkg/search"
)

type App struct {
	mu              sync.RWMutex
	extensionIssues map[string]error
	registry        *operation.Registry
	packages        *packagemanager.Manager
	extensions      *extension.Manager
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
	if err := builtin.Register(registry, packages); err != nil {
		return nil, err
	}
	if err := localtools.Register(registry); err != nil {
		return nil, err
	}
	if err := ocr.Register(registry, packages); err != nil {
		return nil, err
	}
	if err := desktop.Register(registry, packages); err != nil {
		return nil, err
	}
	if err := ffmpegprovider.Register(registry, packages); err != nil {
		return nil, err
	}
	if err := pdfprovider.Register(registry, packages); err != nil {
		return nil, err
	}
	if err := raster.Register(registry, packages); err != nil {
		return nil, err
	}
	if err := tabular.Register(registry, packages); err != nil {
		return nil, err
	}
	if err := document.Register(registry, packages); err != nil {
		return nil, err
	}
	if err := archiveprovider.Register(registry, packages); err != nil {
		return nil, err
	}
	if err := registerWorkflows(registry); err != nil {
		return nil, err
	}
	issues := extensions.RegisterAvailable(registry)
	return &App{registry: registry, packages: packages, extensions: extensions, extensionIssues: issues}, nil
}

func (a *App) snapshot() *operation.Registry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.registry
}

// refresh is called with mu held after an installation mutation.
func (a *App) refresh() error {
	registry := operation.NewRegistry()
	if err := builtin.Register(registry, a.packages); err != nil {
		return err
	}
	if err := localtools.Register(registry); err != nil {
		return err
	}
	if err := ocr.Register(registry, a.packages); err != nil {
		return err
	}
	if err := desktop.Register(registry, a.packages); err != nil {
		return err
	}
	if err := ffmpegprovider.Register(registry, a.packages); err != nil {
		return err
	}
	if err := pdfprovider.Register(registry, a.packages); err != nil {
		return err
	}
	if err := raster.Register(registry, a.packages); err != nil {
		return err
	}
	if err := tabular.Register(registry, a.packages); err != nil {
		return err
	}
	if err := document.Register(registry, a.packages); err != nil {
		return err
	}
	if err := archiveprovider.Register(registry, a.packages); err != nil {
		return err
	}
	if err := registerWorkflows(registry); err != nil {
		return err
	}
	a.extensionIssues = a.extensions.RegisterAvailable(registry)
	a.registry = registry
	return nil
}

func (a *App) Search(query string, limit int) []searchpkg.Match {
	return searchpkg.New(a.snapshot().Definitions()).Search(query, limit)
}

func (a *App) Describe(id string) (operation.Definition, error) {
	capability, ok := a.snapshot().Get(id)
	if !ok {
		return operation.Definition{}, &operation.Error{Code: operation.CodeNotFound, Message: fmt.Sprintf("operation %q was not found", id), Suggestion: fmt.Sprintf("fnsh search %q", id)}
	}
	return capability.Definition, nil
}

func (a *App) Capabilities() []operation.Definition { return a.snapshot().Definitions() }

func (a *App) Execute(ctx context.Context, id string, request operation.Request) (operation.Result, error) {
	capability, ok := a.snapshot().Get(id)
	if !ok {
		return operation.Result{}, &operation.Error{Code: operation.CodeNotFound, Message: fmt.Sprintf("operation %q was not found", id), Suggestion: fmt.Sprintf("fnsh search %q", id)}
	}
	if err := operation.ValidateRequest(capability.Definition, request); err != nil {
		typed := operation.AsError(err)
		typed.Operation = id
		return operation.Result{}, typed
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
	a.mu.Lock()
	defer a.mu.Unlock()
	manifest, err := a.extensions.InstallValidated(source, func(manifest extension.Manifest) error {
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
	if err != nil {
		return extension.Manifest{}, err
	}
	return manifest, a.refresh()
}
func (a *App) RemoveExtension(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.extensions.Remove(name); err != nil {
		return err
	}
	return a.refresh()
}
func (a *App) ExtensionInfo(name string) (extension.Manifest, error) { return a.extensions.Info(name) }
func (a *App) Extensions() []extension.Manifest                      { return a.extensions.List() }

type Check struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (a *App) Doctor() []Check {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.refresh(); err != nil {
		return []Check{{Name: "runtime", OK: false, Message: err.Error()}}
	}
	checks := []Check{{Name: "runtime", OK: true, Message: "operation registry loaded"}}
	for name, err := range a.extensionIssues {
		checks = append(checks, Check{Name: "extension:" + name, OK: false, Message: err.Error()})
	}
	root := a.packages.Root()
	err := os.MkdirAll(root, 0o755)
	checks = append(checks, Check{Name: "data-directory", OK: err == nil, Message: root})
	for _, health := range a.packages.Health() {
		message := statusMessage(health.Err)
		if !health.Installed && health.Err == nil {
			message = "not installed (optional until an operation requires it)"
		}
		checks = append(checks, Check{Name: "package:" + health.Name, OK: health.Err == nil, Message: message})
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
