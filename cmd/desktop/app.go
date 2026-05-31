// app.go defines code for the desktop package.

package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/tractl/tractl/internal/localapi"
)

// App holds the desktop application state.
// Methods on App are bound to the Wails frontend.
type App struct {
	ctx context.Context
	api *localapi.Server
}

type WorkspaceItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type EnvironmentSummary struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	VariableCount int    `json:"variableCount"`
}

type RunHistorySummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Outcome  string `json:"outcome"`
	Duration int    `json:"durationMs"`
}

type GitStatusSummary struct {
	State              string   `json:"state"`
	Branch             string   `json:"branch"`
	ChangedTraCtlFiles []string `json:"changedTraCtlFiles"`
	UncommittedCount   int      `json:"uncommittedCount"`
}

type EngineStatusSummary struct {
	State   string `json:"state"`
	Version string `json:"version"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	root, err := localapi.DefaultRootDir()
	if err != nil {
		fmt.Printf("tractl: cannot resolve data directory: %v\n", maskCredentials(err.Error()))
		return
	}

	store, err := localapi.NewStore(root)
	if err != nil {
		fmt.Printf("tractl: cannot open workspace store: %v\n", maskCredentials(err.Error()))
		return
	}

	a.api = localapi.NewServer("127.0.0.1:7428", store)
	go func() {
		if err := a.api.Start(ctx); err != nil {
			fmt.Printf("tractl: local API stopped: %v\n", maskCredentials(err.Error()))
		}
	}()
}

// maskCredentials replaces common credential patterns in startup error strings.
func maskCredentials(s string) string {
	patterns := []struct {
		re   string
		repl string
	}{
		{`(?i)(bearer\s+)\S+`, `$1[REDACTED]`},
		{`(?i)(basic\s+)\S+`, `$1[REDACTED]`},
		{`(?i)([?&](?:token|key|secret|password)=)[^&\s"]+`, `$1[REDACTED]`},
	}
	result := s
	for _, p := range patterns {
		re := regexp.MustCompile(p.re)
		result = re.ReplaceAllString(result, p.repl)
	}
	return result
}

func (a *App) ListWorkspaceItems() []WorkspaceItem {
	return []WorkspaceItem{}
}

func (a *App) ListEnvironments() []EnvironmentSummary {
	return []EnvironmentSummary{}
}

func (a *App) ListRunHistory() []RunHistorySummary {
	return []RunHistorySummary{}
}

func (a *App) GetGitStatus() GitStatusSummary {
	return GitStatusSummary{
		State:              "untracked",
		Branch:             "",
		ChangedTraCtlFiles: []string{},
		UncommittedCount:   0,
	}
}

func (a *App) GetEngineStatus() EngineStatusSummary {
	return EngineStatusSummary{
		State:   "ready",
		Version: "0.1.0-alpha",
	}
}
