package app

import (
	"context"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

// ExecuteCall shares validation and execution with positional CLI requests.
func (a *App) ExecuteCall(ctx context.Context, call operation.Call) (operation.Result, error) {
	return a.Execute(ctx, call.Operation, operation.Request{Inputs: call.Inputs, Options: call.Options})
}
