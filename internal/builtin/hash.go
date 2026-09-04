package builtin

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func registerHash(registry *operation.Registry) error {
	return registry.Register(operation.Capability{Definition: operation.Definition{
		ID: "hash.calculate", Summary: "Calculate a deterministic content hash", Description: "Hash literal text, a file, or stdin (-) with SHA-256, SHA-1, or MD5.",
		Aliases: []string{"checksum", "文件哈希", "计算哈希"}, Tags: []string{"hash", "checksum", "file"},
		Inputs:  []operation.Parameter{param("input", "Literal text, file path, or - for stdin", true)},
		Options: []operation.Parameter{option("algorithm", operation.TypeString, "sha256, sha1, or md5", "sha256")}, Source: "core",
	}, Runner: operation.Func(runHash)})
}

func runHash(_ context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return operation.Result{}, err
	}
	algorithm, err := operation.StringOption(request, "algorithm", "sha256")
	if err != nil {
		return operation.Result{}, err
	}
	var hasher hash.Hash
	switch algorithm {
	case "sha256":
		hasher = sha256.New()
	case "sha1":
		hasher = sha1.New()
	case "md5":
		hasher = md5.New()
	default:
		return operation.Result{}, &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("unsupported hash algorithm %q", algorithm)}
	}
	data, err := readInput(request.Inputs[0])
	if err != nil {
		return operation.Result{}, err
	}
	_, _ = hasher.Write(data)
	return operation.Result{Operation: "hash.calculate", Data: map[string]any{"algorithm": algorithm, "digest": hex.EncodeToString(hasher.Sum(nil))}}, nil
}
