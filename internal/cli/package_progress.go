package cli

import (
	"fmt"
	"github.com/fu9zhou/finish-bit/pkg/app"
)

func (c *CLI) packageProgress(event app.PackageProgress) {
	switch event.Stage {
	case "downloading":
		if event.Total > 0 {
			fmt.Fprintf(c.stderr, "%s: %.1f / %.1f MiB (%.0f%%) from %s\n", event.Package, float64(event.Bytes)/(1<<20), float64(event.Total)/(1<<20), 100*float64(event.Bytes)/float64(event.Total), event.Source)
		} else {
			fmt.Fprintf(c.stderr, "%s: %.1f MiB from %s\n", event.Package, float64(event.Bytes)/(1<<20), event.Source)
		}
	case "connecting":
		fmt.Fprintf(c.stderr, "%s: connecting to %s\n", event.Package, event.Source)
	case "source-failed":
		fmt.Fprintf(c.stderr, "%s: download from %s failed; trying remaining sources if available\n", event.Package, event.Source)
	case "cache-hit":
		fmt.Fprintf(c.stderr, "%s: using SHA-256 verified download cache\n", event.Package)
	case "already-ready":
		fmt.Fprintf(c.stderr, "%s: installed files verified\n", event.Package)
	case "verifying":
		fmt.Fprintf(c.stderr, "%s: verifying SHA-256\n", event.Package)
	case "extracting":
		fmt.Fprintf(c.stderr, "%s: extracting runtime\n", event.Package)
	case "ready":
		fmt.Fprintf(c.stderr, "%s: installation complete\n", event.Package)
	}
}
