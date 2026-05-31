// run_command.go defines RunCommand, the "run" sub-command for tractl.
// It parses flags, builds an engine.Config, executes the workflow, and formats output.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tractl/tractl/internal/engine"
)

// overlayFlags implements flag.Value to collect repeated --overlay flags.
type overlayFlags []string

func (o *overlayFlags) String() string {
	return fmt.Sprintf("%v", []string(*o))
}

func (o *overlayFlags) Set(v string) error {
	*o = append(*o, v)
	return nil
}

// RunCommand implements the "run" sub-command.
type RunCommand struct{}

// Name returns the sub-command name.
func (c *RunCommand) Name() string { return "run" }

// Execute runs a workflow spec against a target environment.
// args is os.Args[2:] — flags and positional arguments after "run".
func (c *RunCommand) Execute(args []string) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		envName   string
		outputFmt string
		verbose   bool
		overlays  overlayFlags
	)

	fs.StringVar(&envName, "env", "", "active environment name")
	fs.StringVar(&outputFmt, "output", "text", "output format: text or json")
	fs.BoolVar(&verbose, "verbose", false, "include HTTP response status, headers, and body in output")
	fs.Var(&overlays, "overlay", "overlay file path (repeatable)")

	// Parse flags before the workflow file. Go's flag package stops at the
	// first non-flag argument, so flags after the filename are re-parsed.
	if err := fs.Parse(args); err != nil {
		return 2
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "tractl run: workflow file required")
		printUsage()
		return 2
	}

	workflowFile := remaining[0]

	// Re-parse any flags that appeared after the workflow file.
	if len(remaining) > 1 {
		if err := fs.Parse(remaining[1:]); err != nil {
			return 2
		}
	}

	cfg := engine.Config{
		WorkflowFile: workflowFile,
		EnvName:      envName,
		OverlayFiles: []string(overlays),
		Quiet:        outputFmt == "json",
		Verbose:      verbose,
	}

	result := engine.New().Run(cfg)

	switch outputFmt {
	case "json":
		b, err := engine.FormatJSON(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tractl: JSON marshal error: %v\n", err)
			return 2
		}
		fmt.Println(string(b))
	default:
		fmt.Print(engine.FormatText(result))
	}

	return result.ExitCode()
}
