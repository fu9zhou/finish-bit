package app

import "github.com/fu9zhou/finish-bit/pkg/packagemanager"

// PackageCatalog includes optional packages that have not been installed yet.
func (a *App) PackageCatalog() ([]packagemanager.Status, error) {
	registry, err := packagemanager.BuiltinRegistry()
	if err != nil {
		return nil, err
	}
	result := make([]packagemanager.Status, 0, len(registry.Packages))
	for _, pkg := range registry.Packages {
		status, err := a.PackageInfo(pkg.Name)
		if err != nil {
			return nil, err
		}
		result = append(result, status)
	}
	return result, nil
}
