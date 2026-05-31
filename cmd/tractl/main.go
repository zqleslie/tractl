// Command tractl is the traCtl CLI entrypoint.
//
// Usage:
//
//	tractl run <workflow-file> [flags]
//
// Flags:
//
//	--env <name>       active environment name (default: none)
//	--overlay <path>   overlay file; repeatable for multiple overlays
//	--output <format>  output format: text (default) or json

package main

import "os"

func main() {
	commands := map[string]Command{
		"run": &RunCommand{},
	}
	os.Exit(dispatch(commands, os.Args[1:]))
}
