package localapi

import (
	"net/http"
	"os/exec"
	"strings"
)

func (s *Server) handleWorkspaceStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()
	status := WorkspaceStatus{RootDir: s.store.rootDir}

	if err := exec.CommandContext(ctx, "git", "-C", s.store.rootDir, "rev-parse", "--git-dir").Run(); err == nil {
		status.IsGitRepo = true

		if out, err := exec.CommandContext(ctx, "git", "-C", s.store.rootDir, "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
			status.Branch = strings.TrimSpace(string(out))
		}

		if out, err := exec.CommandContext(ctx, "git", "-C", s.store.rootDir, "status", "--porcelain").Output(); err == nil {
			count := 0
			for _, line := range strings.Split(string(out), "\n") {
				if strings.TrimSpace(line) != "" {
					count++
				}
			}
			status.DirtyFiles = count
		}
	}

	jsonResponse(w, http.StatusOK, status)
}
