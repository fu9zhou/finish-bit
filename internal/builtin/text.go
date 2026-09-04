package builtin

import (
	"context"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func registerText(registry *operation.Registry) error {
	input := []operation.Parameter{param("input", "Text, file path, or - for stdin", true)}
	output := []operation.Parameter{option("output", operation.TypeString, "Write the result to this file", "")}
	return registerAll(registry,
		operation.Capability{Definition: operation.Definition{ID: "text.count", Summary: "Count text bytes, characters, words, and lines", Description: "Return deterministic text statistics.", Aliases: []string{"word count", "文本统计"}, Tags: []string{"text", "count"}, Inputs: input, Source: "core"}, Runner: operation.Func(runTextCount)},
		operation.Capability{Definition: operation.Definition{ID: "text.replace", Summary: "Replace text literally", Description: "Replace literal occurrences without regular expressions.", Aliases: []string{"find replace", "文本替换"}, Tags: []string{"text", "replace"}, Inputs: append(input, param("old", "Text to find", true), param("new", "Replacement text", true)), Options: append([]operation.Parameter{option("count", operation.TypeInteger, "Maximum replacements; -1 means all", -1)}, output...), Source: "core"}, Runner: operation.Func(runTextReplace)},
		operation.Capability{Definition: operation.Definition{ID: "text.sort", Summary: "Sort lines", Description: "Sort text lines lexicographically.", Aliases: []string{"sort lines", "文本排序"}, Tags: []string{"text", "lines", "sort"}, Inputs: input, Options: append([]operation.Parameter{option("descending", operation.TypeBoolean, "Sort in descending order", false)}, output...), Source: "core"}, Runner: operation.Func(runTextSort)},
		operation.Capability{Definition: operation.Definition{ID: "text.unique", Summary: "Remove duplicate lines", Description: "Keep the first occurrence of each line.", Aliases: []string{"deduplicate lines", "文本去重"}, Tags: []string{"text", "lines", "unique"}, Inputs: input, Options: output, Source: "core"}, Runner: operation.Func(runTextUnique)},
	)
}

func textValue(request operation.Request) (string, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return "", err
	}
	data, err := readInput(request.Inputs[0])
	return string(data), err
}

func runTextCount(_ context.Context, request operation.Request) (operation.Result, error) {
	text, err := textValue(request)
	if err != nil {
		return operation.Result{}, err
	}
	lines := 0
	if text != "" {
		lines = strings.Count(text, "\n")
		if !strings.HasSuffix(text, "\n") {
			lines++
		}
	}
	return operation.Result{Operation: "text.count", Data: map[string]any{"bytes": len([]byte(text)), "characters": utf8.RuneCountInString(text), "words": len(strings.Fields(text)), "lines": lines}}, nil
}

func runTextReplace(_ context.Context, request operation.Request) (operation.Result, error) {
	if err := operation.RequireInputs(request, 3); err != nil {
		return operation.Result{}, err
	}
	text, err := textValue(request)
	if err != nil {
		return operation.Result{}, err
	}
	count, err := operation.IntOption(request, "count", -1)
	if err != nil {
		return operation.Result{}, err
	}
	output, err := outputOption(request)
	if err != nil {
		return operation.Result{}, err
	}
	return finishText("text.replace", strings.Replace(text, request.Inputs[1], request.Inputs[2], count), output)
}

func scanLines(text string) []string {
	if text == "" {
		return nil
	}
	text = strings.TrimSuffix(text, "\n")
	lines := strings.Split(text, "\n")
	for index := range lines {
		lines[index] = strings.TrimSuffix(lines[index], "\r")
	}
	return lines
}

func runTextSort(_ context.Context, request operation.Request) (operation.Result, error) {
	text, err := textValue(request)
	if err != nil {
		return operation.Result{}, err
	}
	lines := scanLines(text)
	sort.Strings(lines)
	descending, err := operation.BoolOption(request, "descending", false)
	if err != nil {
		return operation.Result{}, err
	}
	if descending {
		for left, right := 0, len(lines)-1; left < right; left, right = left+1, right-1 {
			lines[left], lines[right] = lines[right], lines[left]
		}
	}
	output, err := outputOption(request)
	if err != nil {
		return operation.Result{}, err
	}
	return finishText("text.sort", strings.Join(lines, "\n"), output)
}

func runTextUnique(_ context.Context, request operation.Request) (operation.Result, error) {
	text, err := textValue(request)
	if err != nil {
		return operation.Result{}, err
	}
	seen := map[string]bool{}
	unique := []string{}
	for _, line := range scanLines(text) {
		if !seen[line] {
			seen[line] = true
			unique = append(unique, line)
		}
	}
	output, err := outputOption(request)
	if err != nil {
		return operation.Result{}, err
	}
	return finishText("text.unique", strings.Join(unique, "\n"), output)
}
