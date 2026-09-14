package packagemanager

import "context"

// Progress reports package lifecycle events independently of an output adapter.
// Download events are throttled to at most one per second, plus start/completion.
type Progress struct {
	Package     string
	Stage       string
	Source      string
	Bytes       int64
	Total       int64
	SourceIndex int // One-based position in the current artifact/resource's sources.
	SourceCount int
	Reason      string
	NextSource  string
}
type downloadSourceKey struct{}
type progressKey struct{}
type packageKey struct{}

// WithProgress attaches a synchronous observer to this call and nested installs.
func WithProgress(ctx context.Context, observer func(Progress)) context.Context {
	return context.WithValue(ctx, progressKey{}, observer)
}
func reportProgress(ctx context.Context, event Progress) {
	if source, ok := ctx.Value(downloadSourceKey{}).(Progress); ok {
		event.SourceIndex, event.SourceCount = source.SourceIndex, source.SourceCount
	}
	event.Package, _ = ctx.Value(packageKey{}).(string)
	if observer, ok := ctx.Value(progressKey{}).(func(Progress)); ok && observer != nil {
		observer(event)
	}
}
