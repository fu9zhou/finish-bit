package cli

import (
	"context"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"os"
	"strings"
)

func (c *CLI) executeRequest(ctx context.Context, args []string, jsonOutput bool) error {
	source := ""
	if args[0] == "--request" && len(args) == 2 {
		source = args[1]
	} else if strings.HasPrefix(args[0], "--request=") && len(args) == 1 {
		source = strings.TrimPrefix(args[0], "--request=")
	}
	if source == "" {
		return invalid("run --request requires one JSON file or -; positional inputs and operation options cannot be mixed with a request")
	}
	input := os.Stdin
	if source != "-" {
		file, err := os.Open(source)
		if err != nil {
			return &operation.Error{Code: operation.CodeInvalidInput, Message: "cannot open structured request file", Err: err}
		}
		defer file.Close()
		input = file
	}
	call, err := operation.DecodeCall(input)
	if err != nil {
		return err
	}
	if source == "-" && len(call.Inputs) > 0 && call.Inputs[0] == "-" {
		return invalid("stdin is already used by the structured request; supply operation data inline or in a file")
	}
	result, err := c.app.ExecuteCall(ctx, call)
	if err != nil {
		return err
	}
	return c.print(result, jsonOutput, func() { printResult(c.stdout, result) })
}
