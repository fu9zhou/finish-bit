package builtin

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func registerTime(registry *operation.Registry) error {
	return registry.Register(operation.Capability{Definition: operation.Definition{
		ID: "time.convert", Summary: "Convert timestamps and RFC 3339 times", Description: "Convert Unix seconds, Unix milliseconds, or RFC 3339 into a requested timezone and format.",
		Aliases: []string{"timestamp convert", "时间戳转换"}, Tags: []string{"time", "timestamp"}, Inputs: []operation.Parameter{param("input", "Unix timestamp or RFC 3339 value", true)},
		Options: []operation.Parameter{option("unit", operation.TypeString, "auto, seconds, milliseconds, or rfc3339", "auto"), option("timezone", operation.TypeString, "IANA timezone name", "UTC"), option("format", operation.TypeString, "Go time layout or unix", time.RFC3339)}, Source: "core",
	}, Runner: operation.Func(runTimeConvert)})
}

func runTimeConvert(_ context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return operation.Result{}, err
	}
	unit, err := operation.StringOption(request, "unit", "auto")
	if err != nil {
		return operation.Result{}, err
	}
	var parsed time.Time
	if unit == "auto" {
		var timestamp int64
		timestamp, err = strconv.ParseInt(request.Inputs[0], 10, 64)
		if err == nil {
			if timestamp > 100000000000 {
				parsed = time.UnixMilli(timestamp)
			} else {
				parsed = time.Unix(timestamp, 0)
			}
		} else {
			parsed, err = time.Parse(time.RFC3339, request.Inputs[0])
		}
	} else if unit == "rfc3339" {
		parsed, err = time.Parse(time.RFC3339, request.Inputs[0])
	} else {
		var timestamp int64
		timestamp, err = strconv.ParseInt(request.Inputs[0], 10, 64)
		if err == nil {
			if unit == "milliseconds" {
				parsed = time.UnixMilli(timestamp)
			} else if unit == "seconds" {
				parsed = time.Unix(timestamp, 0)
			} else {
				return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("unsupported time unit %q", unit)}
			}
		}
	}
	if err != nil {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: "input is not a valid timestamp or RFC 3339 time", Err: err}
	}
	timezone, err := operation.StringOption(request, "timezone", "UTC")
	if err != nil {
		return operation.Result{}, err
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("unknown timezone %q", timezone), Err: err}
	}
	format, err := operation.StringOption(request, "format", time.RFC3339)
	if err != nil {
		return operation.Result{}, err
	}
	parsed = parsed.In(location)
	value := parsed.Format(format)
	if format == "unix" {
		value = strconv.FormatInt(parsed.Unix(), 10)
	}
	return operation.Result{Operation: "time.convert", Data: map[string]any{"value": value, "unix": parsed.Unix(), "timezone": timezone}}, nil
}
