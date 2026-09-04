package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/app"
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
