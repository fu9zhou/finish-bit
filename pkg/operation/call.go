package operation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Call is a transport-independent invocation of one Operation.
type Call struct {
	Operation string         `json:"operation"`
	Inputs    []string       `json:"inputs,omitempty"`
	Options   map[string]any `json:"options,omitempty"`
}

const MaxCallBytes int64 = 1 << 20

// DecodeCall accepts exactly one UTF-8 JSON object, preserving integer precision.
func DecodeCall(reader io.Reader) (Call, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxCallBytes+1))
	if err != nil {
		return Call{}, &Error{Code: CodeInvalidInput, Message: "cannot read structured request", Err: err}
	}
	if int64(len(data)) > MaxCallBytes {
		return Call{}, &Error{Code: CodeInvalidInput, Message: "structured request exceeds 1 MiB"}
	}
	if !utf8.Valid(data) {
		return Call{}, &Error{Code: CodeInvalidInput, Message: "structured request must be UTF-8"}
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err == nil {
		for _, name := range []string{"inputs", "options"} {
			if raw, ok := fields[name]; ok && string(bytes.TrimSpace(raw)) == "null" {
				return Call{}, &Error{Code: CodeInvalidInput, Message: name + " cannot be null"}
			}
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var call Call
	if err := decoder.Decode(&call); err != nil {
		return Call{}, &Error{Code: CodeInvalidInput, Message: fmt.Sprintf("invalid structured request: %v", err), Err: err}
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Call{}, &Error{Code: CodeInvalidInput, Message: "structured request must contain exactly one JSON object"}
	}
	if strings.TrimSpace(call.Operation) == "" {
		return Call{}, &Error{Code: CodeInvalidInput, Message: "structured request requires operation"}
	}
	return call, nil
}
