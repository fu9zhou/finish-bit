package operation

import (
	"fmt"
	"strconv"
)

func StringOption(request Request, name, fallback string) (string, error) {
	value, ok := request.Options[name]
	if !ok {
		return fallback, nil
	}
	text, ok := value.(string)
	if !ok {
		return "", invalidOption(name, "string")
	}
	return text, nil
}

func BoolOption(request Request, name string, fallback bool) (bool, error) {
	value, ok := request.Options[name]
	if !ok {
		return fallback, nil
	}
	switch typed := value.(type) {
	case bool:
		return typed, nil
	case string:
		parsed, err := strconv.ParseBool(typed)
		if err != nil {
			return false, invalidOption(name, "boolean")
		}
		return parsed, nil
	default:
		return false, invalidOption(name, "boolean")
	}
}

func IntOption(request Request, name string, fallback int) (int, error) {
	value, ok := request.Options[name]
	if !ok {
		return fallback, nil
	}
	switch typed := value.(type) {
	case int:
		return typed, nil
	case float64:
		return int(typed), nil
	case string:
		parsed, err := strconv.Atoi(typed)
		if err != nil {
			return 0, invalidOption(name, "integer")
		}
		return parsed, nil
	default:
		return 0, invalidOption(name, "integer")
	}
}

func RequireInputs(request Request, count int) error {
	if len(request.Inputs) < count {
		return &Error{Code: CodeInvalidInput, Message: fmt.Sprintf("expected at least %d input(s), got %d", count, len(request.Inputs))}
	}
	return nil
}

func invalidOption(name, expected string) error {
	return &Error{Code: CodeInvalidInput, Message: fmt.Sprintf("option %q must be a %s", name, expected)}
}
