package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func registerJSON(registry *operation.Registry) error {
	input := []operation.Parameter{param("input", "JSON text, file path, or - for stdin", true)}
	output := []operation.Parameter{option("output", operation.TypeString, "Write the result to this file", "")}
	return registerAll(registry,
		operation.Capability{Definition: operation.Definition{ID: "json.format", Summary: "Format JSON with stable indentation", Description: "Parse and pretty-print JSON.", Aliases: []string{"pretty json", "格式化 json"}, Tags: []string{"json", "format"}, Inputs: input, Options: append([]operation.Parameter{option("indent", operation.TypeInteger, "Spaces per indentation level", 2)}, output...), Source: "core"}, Runner: operation.Func(runJSONFormat)},
		operation.Capability{Definition: operation.Definition{ID: "json.minify", Summary: "Minify JSON", Description: "Parse JSON and remove insignificant whitespace.", Aliases: []string{"compress json", "压缩 json"}, Tags: []string{"json", "format"}, Inputs: input, Options: output, Source: "core"}, Runner: operation.Func(runJSONMinify)},
		operation.Capability{Definition: operation.Definition{ID: "json.validate", Summary: "Validate JSON syntax", Description: "Validate JSON and report its root value type.", Aliases: []string{"check json", "校验 json"}, Tags: []string{"json", "validate"}, Inputs: input, Source: "core"}, Runner: operation.Func(runJSONValidate)},
		operation.Capability{Definition: operation.Definition{ID: "json.query", Summary: "Read a value from JSON", Description: "Resolve a dotted object path with numeric array indexes.", Aliases: []string{"get json path", "查询 json"}, Tags: []string{"json", "query"}, Inputs: append(input, param("path", "Dotted path such as users.0.name", true)), Options: output, Source: "core"}, Runner: operation.Func(runJSONQuery)},
	)
}

func parseJSON(request operation.Request) (any, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return nil, err
	}
	data, err := readInput(request.Inputs[0])
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, &operation.Error{Code: operation.CodeInvalidInput, Message: "input is not valid JSON", Err: err}
	}
	if trailingErr := decoder.Decode(&struct{}{}); trailingErr != io.EOF {
		return nil, &operation.Error{Code: operation.CodeInvalidInput, Message: "input contains multiple JSON values"}
	}
	return value, nil
}

func marshalJSON(value any, prefix, indent string) (string, error) {
	data, err := json.MarshalIndent(value, prefix, indent)
	if err != nil {
		return "", fmt.Errorf("encode JSON: %w", err)
	}
	return string(data) + "\n", nil
}

func runJSONFormat(_ context.Context, request operation.Request) (operation.Result, error) {
	value, err := parseJSON(request)
	if err != nil {
		return operation.Result{}, err
	}
	indent, err := operation.IntOption(request, "indent", 2)
	if err != nil {
		return operation.Result{}, err
	}
	if indent < 0 || indent > 16 {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: "indent must be between 0 and 16"}
	}
	text, err := marshalJSON(value, "", strings.Repeat(" ", indent))
	if err != nil {
		return operation.Result{}, err
	}
	output, err := outputOption(request)
	if err != nil {
		return operation.Result{}, err
	}
	return finishText("json.format", text, output)
}

func runJSONMinify(_ context.Context, request operation.Request) (operation.Result, error) {
	value, err := parseJSON(request)
	if err != nil {
		return operation.Result{}, err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return operation.Result{}, fmt.Errorf("encode JSON: %w", err)
	}
	output, err := outputOption(request)
	if err != nil {
		return operation.Result{}, err
	}
	return finishText("json.minify", string(data), output)
}

func runJSONValidate(_ context.Context, request operation.Request) (operation.Result, error) {
	value, err := parseJSON(request)
	if err != nil {
		return operation.Result{}, err
	}
	kind := fmt.Sprintf("%T", value)
	switch value.(type) {
	case map[string]any:
		kind = "object"
	case []any:
		kind = "array"
	case string:
		kind = "string"
	case json.Number:
		kind = "number"
	case bool:
		kind = "boolean"
	case nil:
		kind = "null"
	}
	return operation.Result{Operation: "json.validate", Data: map[string]any{"valid": true, "type": kind}}, nil
}

func runJSONQuery(_ context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 2); err != nil {
		return operation.Result{}, err
	}
	value, err := parseJSON(request)
	if err != nil {
		return operation.Result{}, err
	}
	for _, part := range strings.Split(request.Inputs[1], ".") {
		switch current := value.(type) {
		case map[string]any:
			var found bool
			value, found = current[part]
			if !found {
				return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("JSON path segment %q does not exist", part)}
			}
		case []any:
			index, parseErr := strconv.Atoi(part)
			if parseErr != nil || index < 0 || index >= len(current) {
				return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("JSON array index %q is invalid", part)}
			}
			value = current[index]
		default:
			return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("cannot traverse path segment %q", part)}
		}
	}
	text, err := marshalJSON(value, "", "  ")
	if err != nil {
		return operation.Result{}, err
	}
	output, err := outputOption(request)
	if err != nil {
		return operation.Result{}, err
	}
	return finishText("json.query", text, output)
}
