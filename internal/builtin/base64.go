package builtin

import (
	"context"
	"encoding/base64"
	"strings"
	"unicode/utf8"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func registerBase64(registry *operation.Registry) error {
	commonOptions := []operation.Parameter{
		option("url-safe", operation.TypeBoolean, "Use URL-safe Base64 alphabet", false),
		option("output", operation.TypeString, "Write the result to this file", ""),
	}
	return registerAll(registry,
		operation.Capability{Definition: operation.Definition{
			ID: "base64.encode", Summary: "Encode data as Base64", Description: "Encode literal text, a file, or stdin (-) as Base64.",
			Aliases: []string{"base64 编码", "encode base64"}, Tags: []string{"encoding", "text", "binary"},
			Inputs: []operation.Parameter{param("input", "Literal text, file path, or - for stdin", true)}, Options: commonOptions, Source: "core",
		}, Runner: operation.Func(runBase64Encode)},
		operation.Capability{Definition: operation.Definition{
			ID: "base64.decode", Summary: "Decode Base64 data", Description: "Decode Base64 literal text, a file, or stdin (-).",
			Aliases: []string{"base64 解码", "decode base64"}, Tags: []string{"encoding", "text", "binary"},
			Inputs: []operation.Parameter{param("input", "Base64 text, file path, or - for stdin", true)}, Options: commonOptions, Source: "core",
		}, Runner: operation.Func(runBase64Decode)},
	)
}

func runBase64Encode(_ context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return operation.Result{}, err
	}
	data, err := readInput(request.Inputs[0])
	if err != nil {
		return operation.Result{}, err
	}
	urlSafe, err := operation.BoolOption(request, "url-safe", false)
	if err != nil {
		return operation.Result{}, err
	}
	encoding := base64.StdEncoding
	if urlSafe {
		encoding = base64.URLEncoding
	}
	output, err := outputOption(request)
	if err != nil {
		return operation.Result{}, err
	}
	return finishText("base64.encode", encoding.EncodeToString(data), output)
}

func runBase64Decode(_ context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return operation.Result{}, err
	}
	data, err := readInput(request.Inputs[0])
	if err != nil {
		return operation.Result{}, err
	}
	urlSafe, err := operation.BoolOption(request, "url-safe", false)
	if err != nil {
		return operation.Result{}, err
	}
	encoding := base64.StdEncoding
	if urlSafe {
		encoding = base64.URLEncoding
	}
	decoded, err := encoding.DecodeString(strings.TrimSpace(string(data)))
	if err != nil {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: "input is not valid Base64", Err: err}
	}
	output, err := outputOption(request)
	if err != nil {
		return operation.Result{}, err
	}
	if output == "" && !utf8.Valid(decoded) {
		return operation.Result{Operation: "base64.decode", Data: map[string]any{"base64": base64.StdEncoding.EncodeToString(decoded), "encoding": "base64"}}, nil
	}
	return finishText("base64.decode", string(decoded), output)
}
