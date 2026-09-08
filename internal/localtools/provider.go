// Package localtools implements bounded, offline utilities without a service or interpreter.
package localtools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

const maxText = 4 << 20

type spec struct {
	id, summary, alias string
	inputs             []operation.Parameter
	options            []operation.Parameter
	run                func(context.Context, *toolrun.Values, []string) (map[string]any, error)
}

func in(names ...string) []operation.Parameter {
	p := []operation.Parameter{}
	for _, n := range names {
		p = append(p, toolrun.Param(n, "Literal value; use --input-mode file or stdin to read text", true))
	}
	return p
}
func str(n, d, f string) operation.Parameter { return toolrun.Option(n, operation.TypeString, d, f) }
func integer(n, d string, f int) operation.Parameter {
	return toolrun.Option(n, operation.TypeInteger, d, f)
}
func boolean(n, d string, f bool) operation.Parameter {
	return toolrun.Option(n, operation.TypeBoolean, d, f)
}
func list(n, d string) operation.Parameter {
	return toolrun.Option(n, operation.TypeStrings, d, []string{})
}
func invalid(s string) error        { return toolrun.Invalid(s) }
func value(s string) map[string]any { return map[string]any{"text": s} }

func Register(r *operation.Registry) error {
	specs := append(textSpecs(), numberSpecs()...)
	specs = append(specs, timeSpecs()...)
	specs = append(specs, randomSpecs()...)
	specs = append(specs, formatSpecs()...)
	specs = append(specs, chineseSpecs()...)
	specs = append(specs, documentSpecs()...)
	specs = append(specs, extraSpecs()...)
	for _, s := range specs {
		def := operation.Definition{ID: s.id, Summary: s.summary, Description: s.summary + ". Offline local execution. Text inputs default to literal, are UTF-8 and limited to 4 MiB; file/stdin require explicit input-mode. See docs/local-tools.md.", Aliases: []string{s.alias}, Tags: []string{strings.Split(s.id, ".")[0], "offline"}, Source: "core", Inputs: s.inputs, Options: append(append([]operation.Parameter{}, s.options...), str("input-mode", "literal, file (all inputs), or stdin (first input only)", "literal"), str("output", "Optional UTF-8 text or JSON result file", ""), boolean("overwrite", "Replace an existing output", false))}
		if err := r.Register(operation.Capability{Definition: def, Runner: operation.Func(func(ctx context.Context, req operation.Request) (operation.Result, error) {
			if err := operation.ValidateRequest(def, req); err != nil {
				return operation.Result{}, err
			}
			if err := ctx.Err(); err != nil {
				return operation.Result{}, err
			}
			v := &toolrun.Values{Request: req}
			mode := v.Enum("input-mode", "literal", "literal", "file", "stdin")
			output := v.String("output", "")
			overwrite := v.Bool("overwrite", false)
			if v.Err != nil {
				return operation.Result{}, v.Err
			}
			inputs := make([]string, len(req.Inputs))
			for i, x := range req.Inputs {
				var err error
				inputs[i], err = readText(x, mode, i)
				if err != nil {
					return operation.Result{}, err
				}
			}
			data, err := s.run(ctx, v, inputs)
			if v.Err != nil {
				return operation.Result{}, v.Err
			}
			if err != nil {
				return operation.Result{}, err
			}
			if err = ctx.Err(); err != nil {
				return operation.Result{}, err
			}
			encoded, err := json.Marshal(data)
			if err != nil {
				return operation.Result{}, err
			}
			if len(encoded) > 16<<20 {
				return operation.Result{}, invalid("result exceeds 16 MiB")
			}
			result := operation.Result{Operation: s.id, Data: data}
			if output != "" {
				payload := string(encoded) + "\n"
				if t, ok := data["text"].(string); ok {
					payload = t
				}
				err = toolrun.Artifact(ctx, output, overwrite, func(p string) error { return os.WriteFile(p, []byte(payload), 0600) })
				if err != nil {
					return operation.Result{}, err
				}
				result.Outputs = []string{output}
				result.Data = nil
			}
			return result, nil
		})}); err != nil {
			return err
		}
	}
	if err := registerImages(r); err != nil {
		return err
	}
	return registerDocumentCompression(r)
}

func jsonBytes(data map[string]any) ([]byte, error) {
	if t, ok := data["text"].(string); ok {
		return []byte(t), nil
	}
	b, e := json.MarshalIndent(data, "", "  ")
	return append(b, '\n'), e
}
func readText(x, mode string, index int) (string, error) {
	var data []byte
	if mode == "file" {
		p, err := toolrun.LocalFile(x)
		if err != nil {
			return "", err
		}
		f, err := os.Open(p)
		if err != nil {
			return "", err
		}
		defer f.Close()
		data, err = io.ReadAll(io.LimitReader(f, maxText+1))
		if err != nil {
			return "", err
		}
	} else if mode == "stdin" && index == 0 {
		var err error
		data, err = io.ReadAll(io.LimitReader(os.Stdin, maxText+1))
		if err != nil {
			return "", err
		}
	} else {
		data = []byte(x)
	}
	if len(data) > maxText {
		return "", invalid("input exceeds 4 MiB")
	}
	if !utf8.Valid(data) {
		return "", invalid("text input must be valid UTF-8")
	}
	return string(data), nil
}
func requireLen(s string, n int) error {
	if len(s) > n {
		return invalid(fmt.Sprintf("input exceeds %d bytes", n))
	}
	return nil
}
