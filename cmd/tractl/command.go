// command.go defines the Command interface and the command dispatcher for the tractl CLI.
// Each sub-command is a self-contained struct implementing Command.
// To add a new sub-command: implement Command and register it in main.go.

package main

import (
	"fmt"
	"os"
)

// Command is implemented by each tractl sub-command.
type Command interface {
	// Name returns the sub-command name as typed on the CLI (e.g. "run").
	Name() string
	// Execute runs the sub-command with the given arguments (excluding the sub-command name).
	// It returns an exit code: 0 for success, non-zero for failure.
	Execute(args []string) int
}

// dispatch routes args to the correct Command and returns its exit code.
// args is os.Args[1:] — the first element is expected to be the sub-command name.
func dispatch(commands map[string]Command, args []string) int {
	if len(args) == 0 {
		printUsage()
		return 2
	}
	cmd, ok := commands[args[0]]
	if !ok {
		printUsage()
		return 2
	}
	return cmd.Execute(args[1:])
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: tractl run [--env <name>] [--overlay <path>] [--output text|json] [--verbose] <workflow-file>")
}
