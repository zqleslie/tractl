// command_test.go tests the Command interface, dispatcher, and RunCommand registration.

package main

import (
	"testing"
)

// fakeCommand is a test double that records Execute calls.
type fakeCommand struct {
	name     string
	executed bool
	args     []string
	exitCode int
}

func (f *fakeCommand) Name() string { return f.name }
func (f *fakeCommand) Execute(args []string) int {
	f.executed = true
	f.args = args
	return f.exitCode
}

func TestDispatch_KnownCommand(t *testing.T) {
	fake := &fakeCommand{name: "run", exitCode: 0}
	cmds := map[string]Command{"run": fake}

	code := dispatch(cmds, []string{"run", "--flag", "value"})

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !fake.executed {
		t.Error("expected Execute to be called")
	}
	if len(fake.args) != 2 || fake.args[0] != "--flag" {
		t.Errorf("unexpected args passed to Execute: %v", fake.args)
	}
}

func TestDispatch_UnknownCommand(t *testing.T) {
	cmds := map[string]Command{"run": &fakeCommand{name: "run"}}

	code := dispatch(cmds, []string{"unknown"})

	if code == 0 {
		t.Error("expected non-zero exit code for unknown sub-command")
	}
}

func TestDispatch_NoArgs(t *testing.T) {
	cmds := map[string]Command{"run": &fakeCommand{name: "run"}}

	code := dispatch(cmds, []string{})

	if code == 0 {
		t.Error("expected non-zero exit code when no args given")
	}
}

func TestRunCommand_Name(t *testing.T) {
	cmd := &RunCommand{}
	if cmd.Name() != "run" {
		t.Errorf("expected Name() == \"run\", got %q", cmd.Name())
	}
}

func TestRunCommand_ImplementsCommand(t *testing.T) {
	// Compile-time check: RunCommand must implement Command.
	var _ Command = &RunCommand{}
}
