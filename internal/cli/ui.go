package cli

import (
	"context"
	"flag"
	"strings"

	"github.com/fu9zhou/finish-bit/internal/web"
)

func (c *CLI) ui(ctx context.Context, args []string) error {
	action := "start"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		action, args = args[0], args[1:]
	}
	if action != "start" && action != "stop" && action != "status" && action != "run" {
		return invalid("ui accepts start, stop, status, or run")
	}
	flags := flag.NewFlagSet("ui", flag.ContinueOnError)
	flags.SetOutput(c.stdout)
	port := flags.Int("port", 0, "Local port (0 chooses an available port)")
	noOpen := flags.Bool("no-open", false, "Print the URL without opening a browser")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return invalid(err.Error())
	}
	if flags.NArg() != 0 {
		return invalid("ui accepts --port and --no-open")
	}
	if *port < 0 || *port > 65535 {
		return invalid("port must be between 0 and 65535")
	}
	return web.Manage(ctx, c.app, action, *port, !*noOpen, c.stdout)
}
