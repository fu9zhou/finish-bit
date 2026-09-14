package app

import (
	"context"
	"errors"

	"github.com/fu9zhou/finish-bit/pkg/packagemanager"
)

type PackageProgress = packagemanager.Progress
type PackageTask = packagemanager.Task
type PackageTaskInspection = packagemanager.TaskInspection

func (a *App) InspectPackageTask(id string) (PackageTaskInspection, error) {
	return a.packages.InspectTask(id)
}

var ErrPackageTaskRunning = errors.New("package task already running")

func (a *App) PackageTasks() ([]PackageTask, error) { return a.packages.Tasks() }

func (a *App) StartPackageInstall(ctx context.Context, name string, repair bool) (PackageTask, error) {
	if _, err := a.PackageInfo(name); err != nil {
		return PackageTask{}, err
	}
	tasks, err := a.PackageTasks()
	if err != nil {
		return PackageTask{}, err
	}
	for _, task := range tasks {
		if task.Package == name && (task.Status == "running" || task.Status == "recovering" || task.Status == "unknown") {
			return PackageTask{}, ErrPackageTaskRunning
		}
	}
	return a.packages.StartInstall(ctx, name, repair)
}

// WithPackageProgress observes installation without coupling it to CLI output.
func WithPackageProgress(ctx context.Context, observer func(PackageProgress)) context.Context {
	return packagemanager.WithProgress(ctx, observer)
}
