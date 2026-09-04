// Package builtin provides deterministic operations implemented by the Go runtime.
package builtin

import (
	"fmt"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func Register(registry *operation.Registry) error {
	registrars := []func(*operation.Registry) error{
		registerBase64,
		registerHash,
		registerJSON,
		registerText,
		registerTime,
		registerUUID,
		registerFile,
		registerURL,
	}
	for _, registrar := range registrars {
		if err := registrar(registry); err != nil {
			return fmt.Errorf("register built-in operation: %w", err)
		}
	}
	return nil
}
