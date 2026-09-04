package builtin

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func registerUUID(registry *operation.Registry) error {
	return registry.Register(operation.Capability{Definition: operation.Definition{ID: "uuid.generate", Summary: "Generate UUID v4 values", Description: "Generate cryptographically random RFC 4122 UUID version 4 values.", Aliases: []string{"uuid v4", "生成 uuid"}, Tags: []string{"uuid", "random", "identifier"}, Options: []operation.Parameter{option("count", operation.TypeInteger, "Number of UUIDs", 1)}, Source: "core"}, Runner: operation.Func(runUUID)})
}

func runUUID(_ context.Context, request operation.Request) (operation.Result, error) {
	count, err := operation.IntOption(request, "count", 1)
	if err != nil {
		return operation.Result{}, err
	}
	if count < 1 || count > 1000 {
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: "count must be between 1 and 1000"}
	}
	values := make([]string, count)
	for index := range values {
		var bytes [16]byte
		if _, err := rand.Read(bytes[:]); err != nil {
			return operation.Result{}, fmt.Errorf("generate random UUID: %w", err)
		}
		bytes[6] = (bytes[6] & 0x0f) | 0x40
		bytes[8] = (bytes[8] & 0x3f) | 0x80
		values[index] = fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
	}
	return operation.Result{Operation: "uuid.generate", Data: map[string]any{"values": values}}, nil
}
