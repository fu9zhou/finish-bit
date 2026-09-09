package packagemanager

import "context"

// Progress reports package lifecycle events independently of an output adapter.
// Download events are throttled to at most one per second, plus start/completion.
type Progress struct {
	Package string
	Stage   string
	Source  string
	Bytes   int64
	Total   int64
}
type progressKey struct{}
type packageKey struct{}

// WithProgress attaches a synchronous observer to this call and nested installs.
func WithProgress(ctx context.Context, observer func(Progress)) context.Context {
	return context.WithValue(ctx, progressKey{}, observer)
}
func reportProgress(ctx context.Context, event Progress) {
	event.Package, _ = ctx.Value(packageKey{}).(string)
	if observer, ok := ctx.Value(progressKey{}).(func(Progress)); ok && observer != nil {
		observer(event)
	}
}
