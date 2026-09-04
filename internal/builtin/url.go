package builtin

import (
	"context"
	"net/url"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func registerURL(registry *operation.Registry) error {
	input := []operation.Parameter{param("input", "Text to encode or decode", true)}
	return registerAll(registry,
		operation.Capability{Definition: operation.Definition{ID: "url.encode", Summary: "Percent-encode URL data", Description: "Encode text as a URL query component.", Aliases: []string{"percent encode", "url 编码"}, Tags: []string{"url", "encoding"}, Inputs: input, Source: "core"}, Runner: operation.Func(func(_ context.Context, request operation.Request) (operation.Result, error) {
			if err := operation.RequireInputs(request, 1); err != nil {
				return operation.Result{}, err
			}
			return finishText("url.encode", url.QueryEscape(request.Inputs[0]), "")
		})},
		operation.Capability{Definition: operation.Definition{ID: "url.decode", Summary: "Decode percent-encoded URL data", Description: "Decode a URL query component.", Aliases: []string{"percent decode", "url 解码"}, Tags: []string{"url", "encoding"}, Inputs: input, Source: "core"}, Runner: operation.Func(runURLDecode)},
	)
}

func runURLDecode(_ context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return operation.Result{}, err
	}
	decoded, err := url.QueryUnescape(request.Inputs[0])
	if err != nil {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: "input is not valid percent-encoded data", Err: err}
	}
	return finishText("url.decode", decoded, "")
}
