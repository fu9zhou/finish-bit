package app

import (
	"context"
	"github.com/fu9zhou/finish-bit/pkg/packagemanager"
)

type PackageProgress = packagemanager.Progress

// WithPackageProgress observes installation without coupling it to CLI output.
func WithPackageProgress(ctx context.Context, observer func(PackageProgress)) context.Context {
	return packagemanager.WithProgress(ctx, observer)
}
