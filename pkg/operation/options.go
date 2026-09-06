package operation

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
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
	case json.Number:
		text := string(typed)
		if len(text) > 1024 {
			return 0, invalidOption(name, "integer")
		}
		if index := strings.IndexAny(text, "eE"); index >= 0 {
			exponent, err := strconv.Atoi(text[index+1:])
			if err != nil || exponent < -1024 || exponent > 1024 {
				return 0, invalidOption(name, "integer")
			}
		}
		rational, ok := new(big.Rat).SetString(text)
		if !ok || !rational.IsInt() || !rational.Num().IsInt64() {
			return 0, invalidOption(name, "integer")
		}
		parsed := rational.Num().Int64()
		if int64(int(parsed)) != parsed {
			return 0, invalidOption(name, "integer")
		}
		return int(parsed), nil
	case int:
		return typed, nil
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) || math.Trunc(typed) != typed {
			return 0, invalidOption(name, "integer")
		}
		parsed, err := strconv.ParseInt(strconv.FormatFloat(typed, 'f', -1, 64), 10, strconv.IntSize)
		if err != nil {
			return 0, invalidOption(name, "integer")
		}
		return int(parsed), nil
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
