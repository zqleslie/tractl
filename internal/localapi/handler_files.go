package localapi

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	prefix := strings.TrimPrefix(r.URL.Query().Get("prefix"), "/")
	entries, err := listWorkspaceFiles(s.store.root(), prefix)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, entries)
}

func (s *Server) handleFilePath(w http.ResponseWriter, r *http.Request) {
	rel := r.PathValue("path")
	switch r.Method {
	case http.MethodGet:
		s.handleFileRead(w, rel)
	case http.MethodPost:
		s.handleFileWrite(w, r, rel)
	case http.MethodDelete:
		s.handleFileDelete(w, rel)
	default:
		httpError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleFileRead(w http.ResponseWriter, rel string) {
	path, err := s.store.resolvePath(rel)
	if err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		httpError(w, http.StatusNotFound, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, FileReadResponse{Path: rel, Content: string(raw)})
}

func (s *Server) handleFileWrite(w http.ResponseWriter, r *http.Request, rel string) {
	var req FileWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if err := validateYAML(req.Content); err != nil {
		httpError(w, http.StatusBadRequest, "invalid yaml: "+err.Error())
		return
	}
	path, err := s.store.resolvePath(rel)
	if err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := os.WriteFile(path, []byte(req.Content), 0o644); err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	name := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
	if strings.HasPrefix(filepath.ToSlash(rel), "requests/") {
		s.store.commitPath(path, "tractl: save request "+name)
	} else if strings.HasPrefix(filepath.ToSlash(rel), "workflows/") {
		s.store.commitPath(path, "tractl: save workflow "+name)
	}
	stat, err := os.Stat(path)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, FileWriteResponse{
		Path:       rel,
		ModifiedAt: stat.ModTime().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleFileDelete(w http.ResponseWriter, rel string) {
	path, err := s.store.resolvePath(rel)
	if err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, err := os.Stat(path); err != nil {
		httpError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := os.Remove(path); err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func listWorkspaceFiles(root, prefix string) ([]FileEntry, error) {
	entries := []FileEntry{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(d.Name()) != ".yaml" {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || (prefix != "" && !strings.HasPrefix(rel, prefix)) {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		entries = append(entries, FileEntry{
			Path:       filepath.ToSlash(rel),
			Name:       d.Name(),
			ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
			Size:       info.Size(),
		})
		return nil
	})
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, err
}

func validateYAML(content string) error {
	var decoded any
	return yaml.Unmarshal([]byte(content), &decoded)
}
