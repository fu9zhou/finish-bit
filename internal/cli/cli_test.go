package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/app"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func newTestCLI(t *testing.T) (*CLI, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	application, err := app.New(app.Config{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	return New(application, stdout, stderr, "test"), stdout, stderr
}

func TestCLICommandFamilies(t *testing.T) {
	command, stdout, stderr := newTestCLI(t)
	assertExit := func(arguments []string, expected int) string {
		t.Helper()
		stdout.Reset()
		stderr.Reset()
		if exit := command.Run(context.Background(), arguments); exit != expected {
			t.Fatalf("Run(%q) exit = %d, want %d; stderr=%s", arguments, exit, expected, stderr.String())
		}
		return stdout.String() + stderr.String()
	}

	for _, arguments := range [][]string{
		{"help"}, {"version", "--json"}, {"search", "format", "json", "--limit", "2"},
		{"describe", "json.format"}, {"capabilities", "--json"},
		{"run", "json.minify", `{"ok": true}`}, {"uuid", "generate", "--count", "2", "--json"},
		{"pkg", "ls"}, {"pkg", "info", "ffmpeg", "--json"}, {"pkg", "remove", "ffmpeg"},
		{"doctor"},
	} {
		if output := assertExit(arguments, 0); output == "" && strings.Join(arguments, " ") != "pkg ls" {
			t.Fatalf("Run(%q) produced no output", arguments)
		}
	}

	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "runner"), []byte("runner"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":1,"name":"cli-extension","version":"1","executable":"runner","operations":[{"id":"cli.echo","summary":"Echo","source":"extension"}]}`
	if err := os.WriteFile(filepath.Join(source, "finishbit-extension.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{
		{"ext", "add", source}, {"ext", "ls", "--json"}, {"ext", "info", "cli-extension"}, {"ext", "remove", "cli-extension", "--json"},
	} {
		assertExit(arguments, 0)
	}

	for _, arguments := range [][]string{
		{"search"}, {"search", "json", "--limit", "bad"}, {"describe"}, {"run"},
		{"json", "format", `{}`, "--missing", "value"}, {"pkg"}, {"pkg", "unknown"}, {"ext"}, {"ext", "unknown"},
	} {
		assertExit(arguments, 2)
	}
}

func TestParseOperationArgumentsCollectsRepeatedStringOptions(t *testing.T) {
	definition := operation.Definition{
		ID:      "test.run",
		Summary: "Run test",
		Source:  "test",
		Options: []operation.Parameter{{Name: "label", Type: operation.TypeStrings, Description: "Labels"}},
	}
	request, err := parseOperationArguments(definition, []string{"--label", "one", "--label=two"})
	if err != nil {
		t.Fatal(err)
	}
	labels, ok := request.Options["label"].([]string)
	if !ok || len(labels) != 2 || labels[0] != "one" || labels[1] != "two" {
		t.Fatalf("labels = %#v", request.Options["label"])
	}
}

func TestDynamicOperationCommand(t *testing.T) {
	command, stdout, stderr := newTestCLI(t)
	exitCode := command.Run(context.Background(), []string{"base64", "encode", "FinishBit"})
	if exitCode != 0 {
		t.Fatalf("exit = %d, stderr = %s", exitCode, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "RmluaXNoQml0" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestJSONErrorIsStructured(t *testing.T) {
	command, _, stderr := newTestCLI(t)
	exitCode := command.Run(context.Background(), []string{"describe", "missing.operation", "--json"})
	if exitCode != 2 {
		t.Fatalf("exit = %d", exitCode)
	}
	if !strings.Contains(stderr.String(), `"code":"operation_not_found"`) {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRejectsExtraOperationInputs(t *testing.T) {
	command, _, stderr := newTestCLI(t)
	exitCode := command.Run(context.Background(), []string{"base64", "encode", "one", "two"})
	if exitCode != 2 {
		t.Fatalf("exit = %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "accepts 1 input") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
