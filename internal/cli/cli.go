// Package cli adapts command-line arguments to the transport-independent application service.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/fu9zhou/finish-bit/pkg/app"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type CLI struct {
	app            *app.App
	stdout, stderr io.Writer
	version        string
}

func New(application *app.App, stdout, stderr io.Writer, version string) *CLI {
	return &CLI{app: application, stdout: stdout, stderr: stderr, version: version}
}

func (c *CLI) Run(ctx context.Context, arguments []string) int {
	jsonOutput, arguments := takeFlag(arguments, "--json")
	if len(arguments) == 0 || arguments[0] == "help" || arguments[0] == "--help" || arguments[0] == "-h" {
		c.help()
		return 0
	}
	if err := c.dispatch(ctx, arguments, jsonOutput); err != nil {
		return c.printError(err, jsonOutput)
	}
	return 0
}

func (c *CLI) dispatch(ctx context.Context, arguments []string, jsonOutput bool) error {
	switch arguments[0] {
	case "version", "--version", "-v":
		return c.print(map[string]any{"version": c.version}, jsonOutput, func() { fmt.Fprintf(c.stdout, "fnsh %s\n", c.version) })
	case "search":
		return c.search(arguments[1:], jsonOutput)
	case "describe":
		return c.describe(arguments[1:], jsonOutput)
	case "capabilities":
		return c.print(c.app.Capabilities(), jsonOutput, func() {
			for _, definition := range c.app.Capabilities() {
				fmt.Fprintf(c.stdout, "%-24s %s\n", definition.ID, definition.Summary)
			}
		})
	case "run":
		return c.execute(ctx, arguments[1:], jsonOutput)
	case "pkg":
		return c.packages(ctx, arguments[1:], jsonOutput)
	case "ext":
		return c.extensions(arguments[1:], jsonOutput)
	case "doctor":
		return c.doctor(jsonOutput)
	default:
		if len(arguments) >= 2 {
			if _, err := c.app.Describe(arguments[0] + "." + arguments[1]); err == nil {
				return c.execute(ctx, append([]string{arguments[0] + "." + arguments[1]}, arguments[2:]...), jsonOutput)
			}
		}
		return &operation.Error{Code: operation.CodeInvalidInput, Message: fmt.Sprintf("unknown command %q", arguments[0]), Suggestion: "fnsh help"}
	}
}

func (c *CLI) search(arguments []string, jsonOutput bool) error {
	limit := 5
	query := []string{}
	for index := 0; index < len(arguments); index++ {
		if arguments[index] == "--limit" {
			if index+1 >= len(arguments) {
				return invalid("--limit requires a value")
			}
			parsed, err := strconv.Atoi(arguments[index+1])
			if err != nil {
				return invalid("--limit must be an integer")
			}
			limit = parsed
			index++
			continue
		}
		query = append(query, arguments[index])
	}
	if len(query) == 0 {
		return invalid("search requires a query")
	}
	matches := c.app.Search(strings.Join(query, " "), limit)
	return c.print(matches, jsonOutput, func() {
		for _, match := range matches {
			fmt.Fprintf(c.stdout, "%-24s %s\n", match.ID, match.Summary)
		}
	})
}

func (c *CLI) describe(arguments []string, jsonOutput bool) error {
	if len(arguments) != 1 {
		return invalid("describe requires exactly one operation ID")
	}
	definition, err := c.app.Describe(arguments[0])
	if err != nil {
		return err
	}
	return c.print(definition, jsonOutput, func() {
		fmt.Fprintf(c.stdout, "%s — %s\n\n%s\n", definition.ID, definition.Summary, definition.Description)
		if len(definition.Inputs) > 0 {
			fmt.Fprintln(c.stdout, "\nInputs:")
			printParameters(c.stdout, definition.Inputs)
		}
		if len(definition.Options) > 0 {
			fmt.Fprintln(c.stdout, "\nOptions:")
			printParameters(c.stdout, definition.Options)
		}
		if len(definition.Requirements) > 0 {
			fmt.Fprintln(c.stdout, "\nRequires:")
			for _, requirement := range definition.Requirements {
				fmt.Fprintf(c.stdout, "  %s\n", requirement.Package)
			}
		}
	})
}

func printParameters(output io.Writer, parameters []operation.Parameter) {
	for _, parameter := range parameters {
		required := ""
		if parameter.Required {
			required = " (required)"
		}
		fmt.Fprintf(output, "  %-16s %-8s %s%s\n", parameter.Name, parameter.Type, parameter.Description, required)
	}
}

func (c *CLI) execute(ctx context.Context, arguments []string, jsonOutput bool) error {
	if len(arguments) == 0 {
		return invalid("run requires an operation ID")
	}
	definition, err := c.app.Describe(arguments[0])
	if err != nil {
		return err
	}
	request, err := parseOperationArguments(definition, arguments[1:])
	if err != nil {
		return err
	}
	result, err := c.app.Execute(ctx, definition.ID, request)
	if err != nil {
		return err
	}
	return c.print(result, jsonOutput, func() { printResult(c.stdout, result) })
}

func parseOperationArguments(definition operation.Definition, arguments []string) (operation.Request, error) {
	parameters := map[string]operation.Parameter{}
	for _, parameter := range definition.Options {
		parameters[parameter.Name] = parameter
	}
	request := operation.Request{Options: map[string]any{}}
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if argument == "-o" {
			argument = "--output"
		}
		if !strings.HasPrefix(argument, "--") {
			request.Inputs = append(request.Inputs, argument)
			continue
		}
		nameValue := strings.SplitN(strings.TrimPrefix(argument, "--"), "=", 2)
		parameter, ok := parameters[nameValue[0]]
		if !ok {
			return operation.Request{}, invalid(fmt.Sprintf("operation %s has no option --%s", definition.ID, nameValue[0]))
		}
		value := "true"
		if len(nameValue) == 2 {
			value = nameValue[1]
		} else if parameter.Type != operation.TypeBoolean {
			if index+1 >= len(arguments) {
				return operation.Request{}, invalid(fmt.Sprintf("--%s requires a value", parameter.Name))
			}
			index++
			value = arguments[index]
		}
		converted, err := convertValue(parameter, value)
		if err != nil {
			return operation.Request{}, err
		}
		request.Options[parameter.Name] = converted
	}
	for index, input := range definition.Inputs {
		if input.Required && index >= len(request.Inputs) {
			return operation.Request{}, invalid(fmt.Sprintf("missing required input %q", input.Name))
		}
	}
	if len(request.Inputs) > len(definition.Inputs) {
		return operation.Request{}, invalid(fmt.Sprintf("operation %s accepts %d input(s), got %d", definition.ID, len(definition.Inputs), len(request.Inputs)))
	}
	for _, option := range definition.Options {
		if option.Required {
			if _, present := request.Options[option.Name]; !present {
				return operation.Request{}, invalid(fmt.Sprintf("missing required option --%s", option.Name))
			}
		}
	}
	return request, nil
}

func convertValue(parameter operation.Parameter, value string) (any, error) {
	switch parameter.Type {
	case operation.TypeString:
		return value, nil
	case operation.TypeInteger:
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return nil, invalid(fmt.Sprintf("--%s must be an integer", parameter.Name))
		}
		return parsed, nil
	case operation.TypeBoolean:
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, invalid(fmt.Sprintf("--%s must be a boolean", parameter.Name))
		}
		return parsed, nil
	default:
		return nil, invalid(fmt.Sprintf("unsupported option type %q", parameter.Type))
	}
}

func (c *CLI) packages(ctx context.Context, arguments []string, jsonOutput bool) error {
	if len(arguments) == 0 {
		return invalid("pkg requires add, remove, ls, info, or repair")
	}
	switch arguments[0] {
	case "ls":
		return c.print(c.app.Packages(), jsonOutput, func() {
			for _, item := range c.app.Packages() {
				fmt.Fprintf(c.stdout, "%s %s %s\n", item.Name, item.Version, item.Platform)
			}
		})
	case "add", "repair":
		if len(arguments) != 2 {
			return invalid("pkg " + arguments[0] + " requires a package name")
		}
		var value any
		var err error
		if arguments[0] == "add" {
			value, err = c.app.InstallPackage(ctx, arguments[1])
		} else {
			value, err = c.app.RepairPackage(ctx, arguments[1])
		}
		if err != nil {
			return err
		}
		return c.print(value, jsonOutput, func() { fmt.Fprintf(c.stdout, "%s is ready\n", arguments[1]) })
	case "remove":
		if len(arguments) != 2 {
			return invalid("pkg remove requires a package name")
		}
		if err := c.app.RemovePackage(arguments[1]); err != nil {
			return err
		}
		return c.print(map[string]any{"removed": arguments[1]}, jsonOutput, func() { fmt.Fprintf(c.stdout, "removed %s\n", arguments[1]) })
	case "info":
		if len(arguments) != 2 {
			return invalid("pkg info requires a package name")
		}
		value, err := c.app.PackageInfo(arguments[1])
		if err != nil {
			return err
		}
		return c.print(value, jsonOutput, func() {
			state := "not installed"
			if value.Installed != nil {
				state = "installed"
			}
			fmt.Fprintf(c.stdout, "%s %s %s (%s, %s)\n", value.Name, value.Version, value.Platform, state, value.License)
		})
	default:
		return invalid("pkg requires add, remove, ls, info, or repair")
	}
}

func (c *CLI) extensions(arguments []string, jsonOutput bool) error {
	if len(arguments) == 0 {
		return invalid("ext requires add, remove, ls, or info")
	}
	switch arguments[0] {
	case "ls":
		values := c.app.Extensions()
		return c.print(values, jsonOutput, func() {
			for _, value := range values {
				fmt.Fprintf(c.stdout, "%s %s\n", value.Name, value.Version)
			}
		})
	case "add":
		if len(arguments) != 2 {
			return invalid("ext add requires a directory or zip path")
		}
		value, err := c.app.InstallExtension(arguments[1])
		if err != nil {
			return err
		}
		return c.print(value, jsonOutput, func() { fmt.Fprintf(c.stdout, "installed %s %s\n", value.Name, value.Version) })
	case "remove":
		if len(arguments) != 2 {
			return invalid("ext remove requires a name")
		}
		if err := c.app.RemoveExtension(arguments[1]); err != nil {
			return err
		}
		return c.print(map[string]any{"removed": arguments[1]}, jsonOutput, func() { fmt.Fprintf(c.stdout, "removed %s\n", arguments[1]) })
	case "info":
		if len(arguments) != 2 {
			return invalid("ext info requires a name")
		}
		value, err := c.app.ExtensionInfo(arguments[1])
		if err != nil {
			return err
		}
		return c.print(value, jsonOutput, func() {
			fmt.Fprintf(c.stdout, "%s %s (%d operations)\n", value.Name, value.Version, len(value.Operations))
		})
	default:
		return invalid("ext requires add, remove, ls, or info")
	}
}

func (c *CLI) doctor(jsonOutput bool) error {
	checks := c.app.Doctor()
	failed := false
	for _, check := range checks {
		if !check.OK {
			failed = true
		}
	}
	err := c.print(checks, jsonOutput, func() {
		for _, check := range checks {
			state := "ok"
			if !check.OK {
				state = "missing"
			}
			fmt.Fprintf(c.stdout, "%-8s %-24s %s\n", state, check.Name, check.Message)
		}
	})
	if err != nil {
		return err
	}
	if failed {
		return errors.New("one or more doctor checks need attention")
	}
	return nil
}

func printResult(output io.Writer, result operation.Result) {
	if text, ok := result.Data["text"].(string); ok {
		fmt.Fprint(output, text)
		if !strings.HasSuffix(text, "\n") {
			fmt.Fprintln(output)
		}
		return
	}
	if len(result.Outputs) > 0 {
		for _, path := range result.Outputs {
			fmt.Fprintln(output, path)
		}
		return
	}
	data, _ := json.MarshalIndent(result.Data, "", "  ")
	fmt.Fprintln(output, string(data))
}

func (c *CLI) print(value any, asJSON bool, human func()) error {
	if !asJSON {
		human()
		return nil
	}
	encoder := json.NewEncoder(c.stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func (c *CLI) printError(err error, asJSON bool) int {
	typed := operation.AsError(err)
	if asJSON {
		_ = json.NewEncoder(c.stderr).Encode(map[string]any{"error": typed})
	} else {
		fmt.Fprintf(c.stderr, "error: %s\n", typed.Message)
		if typed.Suggestion != "" {
			fmt.Fprintf(c.stderr, "try: %s\n", typed.Suggestion)
		}
	}
	if typed.Code == operation.CodeInvalidInput || typed.Code == operation.CodeNotFound {
		return 2
	}
	if typed.Code == operation.CodeDependencyMissing {
		return 3
	}
	return 1
}

func invalid(message string) error {
	return &operation.Error{Code: operation.CodeInvalidInput, Message: message}
}

func takeFlag(arguments []string, flag string) (bool, []string) {
	found := false
	rest := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		if argument == flag {
			found = true
		} else {
			rest = append(rest, argument)
		}
	}
	return found, rest
}

func (c *CLI) help() {
	fmt.Fprint(c.stdout, `fnsh — Search. Reuse. Finish.

Usage:
  fnsh search <query> [--limit N] [--json]
  fnsh describe <operation> [--json]
  fnsh <domain> <action> [inputs...] [options...] [--json]
  fnsh run <operation> [inputs...] [options...] [--json]
  fnsh capabilities [--json]
  fnsh pkg <add|remove|ls|info|repair> [name] [--json]
  fnsh ext <add|remove|ls|info> [name-or-path] [--json]
  fnsh doctor [--json]
  fnsh version [--json]

Examples:
  fnsh search "裁剪并压缩视频"
  fnsh describe video.trim --json
  fnsh json format data.json --indent 2
  fnsh video trim input.mp4 --start 10s --duration 20s -o clip.mp4
`)
}
