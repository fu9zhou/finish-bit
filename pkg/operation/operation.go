// Package operation defines FinishBit's stable capability contract.
package operation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type ValueType string

const (
	TypeString  ValueType = "string"
	TypeInteger ValueType = "integer"
	TypeBoolean ValueType = "boolean"
	TypeStrings ValueType = "strings"
)

type Parameter struct {
	Name        string    `json:"name"`
	Type        ValueType `json:"type"`
	Description string    `json:"description"`
	Required    bool      `json:"required,omitempty"`
	Default     any       `json:"default,omitempty"`
}

type Requirement struct {
	Package string `json:"package"`
}

type Definition struct {
	ID           string        `json:"id"`
	Summary      string        `json:"summary"`
	Description  string        `json:"description"`
	Aliases      []string      `json:"aliases,omitempty"`
	Tags         []string      `json:"tags,omitempty"`
	Inputs       []Parameter   `json:"inputs,omitempty"`
	Options      []Parameter   `json:"options,omitempty"`
	Requirements []Requirement `json:"requirements,omitempty"`
	Source       string        `json:"source"`
}

type Request struct {
	Inputs  []string       `json:"inputs,omitempty"`
	Options map[string]any `json:"options,omitempty"`
}

type Result struct {
	Operation string         `json:"operation"`
	Outputs   []string       `json:"outputs,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
}

type Runner interface {
	Run(context.Context, Request) (Result, error)
}

type Func func(context.Context, Request) (Result, error)

func (f Func) Run(ctx context.Context, request Request) (Result, error) { return f(ctx, request) }

type Capability struct {
	Definition Definition
	Runner     Runner
}

type Registry struct {
	mu   sync.RWMutex
	caps map[string]Capability
}

func NewRegistry() *Registry { return &Registry{caps: make(map[string]Capability)} }

func (r *Registry) Register(cap Capability) error {
	if err := ValidateDefinition(cap.Definition); err != nil {
		return err
	}
	if cap.Runner == nil {
		return fmt.Errorf("operation %q has no runner", cap.Definition.ID)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.caps[cap.Definition.ID]; exists {
		return fmt.Errorf("operation %q already registered", cap.Definition.ID)
	}
	r.caps[cap.Definition.ID] = cap
	return nil
}

func (r *Registry) Get(id string) (Capability, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cap, ok := r.caps[id]
	return cap, ok
}

func (r *Registry) Definitions() []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	definitions := make([]Definition, 0, len(r.caps))
	for _, cap := range r.caps {
		definitions = append(definitions, cap.Definition)
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].ID < definitions[j].ID })
	return definitions
}

func ValidateDefinition(def Definition) error {
	parts := strings.Split(def.ID, ".")
	if len(parts) < 2 {
		return fmt.Errorf("operation id %q must contain a domain and action", def.ID)
	}
	for _, part := range parts {
		if part == "" || strings.Trim(part, "abcdefghijklmnopqrstuvwxyz0123456789-") != "" {
			return fmt.Errorf("operation id %q contains an invalid segment", def.ID)
		}
	}
	if strings.TrimSpace(def.Summary) == "" {
		return fmt.Errorf("operation %q has no summary", def.ID)
	}
	if strings.TrimSpace(def.Source) == "" {
		return fmt.Errorf("operation %q has no source", def.ID)
	}
	metadata := []struct {
		label  string
		values []string
	}{{label: "alias", values: def.Aliases}, {label: "tag", values: def.Tags}}
	for _, group := range metadata {
		seenValues := map[string]bool{}
		for _, value := range group.values {
			if seenValues[value] {
				return fmt.Errorf("operation %q has duplicate %s %q", def.ID, group.label, value)
			}
			seenValues[value] = true
		}
	}
	for _, requirement := range def.Requirements {
		if strings.TrimSpace(requirement.Package) == "" {
			return fmt.Errorf("operation %q has an empty package requirement", def.ID)
		}
	}
	seen := map[string]bool{}
	for _, parameter := range append(append([]Parameter{}, def.Inputs...), def.Options...) {
		if parameter.Name == "" || seen[parameter.Name] {
			return fmt.Errorf("operation %q has an invalid or duplicate parameter %q", def.ID, parameter.Name)
		}
		switch parameter.Type {
		case TypeString, TypeInteger, TypeBoolean, TypeStrings:
		default:
			return fmt.Errorf("operation %q parameter %q has unsupported type %q", def.ID, parameter.Name, parameter.Type)
		}
		seen[parameter.Name] = true
	}
	return nil
}

type ErrorCode string

const (
	CodeInvalidInput        ErrorCode = "invalid_input"
	CodeNotFound            ErrorCode = "operation_not_found"
	CodeDependencyMissing   ErrorCode = "dependency_missing"
	CodeExecutionFailed     ErrorCode = "execution_failed"
	CodeIntegrityFailure    ErrorCode = "integrity_failure"
	CodeUnsupportedPlatform ErrorCode = "unsupported_platform"
)

type Error struct {
	Code       ErrorCode      `json:"code"`
	Message    string         `json:"message"`
	Operation  string         `json:"operation,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
	Suggestion string         `json:"suggestion,omitempty"`
	Err        error          `json:"-"`
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Err }

func AsError(err error) *Error {
	var target *Error
	if errors.As(err, &target) {
		return target
	}
	return &Error{Code: CodeExecutionFailed, Message: err.Error(), Err: err}
}

func DecodeOptions(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var options map[string]any
	if err := json.Unmarshal(raw, &options); err != nil {
		return nil, &Error{Code: CodeInvalidInput, Message: "options must be a JSON object", Err: err}
	}
	return options, nil
}
