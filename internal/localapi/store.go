package localapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"gopkg.in/yaml.v3"
)

// FileRecord tracks persisted workspace file metadata.
type FileRecord struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	UpdatedAt string `json:"updatedAt"`
}

type manifest struct {
	Files []FileRecord `json:"files"`
}

// Store persists request/workflow YAML files under the traCtl app directory.
type Store struct {
	rootDir string
}

// NewStore creates a store at the given root (e.g. ~/.config/tractl).
func NewStore(rootDir string) (*Store, error) {
	if rootDir == "" {
		return nil, errors.New("localapi: empty root directory")
	}
	requestsDir := filepath.Join(rootDir, "requests")
	if err := os.MkdirAll(requestsDir, 0o755); err != nil {
		return nil, fmt.Errorf("localapi: mkdir requests: %w", err)
	}
	return &Store{rootDir: rootDir}, nil
}

// DefaultRootDir returns the OS-appropriate traCtl app data directory.
func DefaultRootDir() (string, error) {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "tractl"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "tractl"), nil
}

func (s *Store) manifestPath() string {
	return filepath.Join(s.rootDir, "manifest.json")
}

func (s *Store) loadManifest() (manifest, error) {
	path := s.manifestPath()
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return manifest{Files: []FileRecord{}}, nil
		}
		return manifest{}, err
	}
	var m manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return manifest{}, err
	}
	if m.Files == nil {
		m.Files = []FileRecord{}
	}
	return m, nil
}

func (s *Store) saveManifest(m manifest) error {
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.manifestPath(), raw, 0o644)
}

// SaveYAML persists document bytes and updates the manifest.
func (s *Store) SaveYAML(id, kind, name string, yamlBytes []byte) (FileRecord, error) {
	if strings.TrimSpace(name) == "" {
		name = "Untitled request"
	}
	if id == "" {
		id = ulid.Make().String()
	}

	path := filepath.Join(s.rootDir, "requests", id+".yaml")
	if err := os.WriteFile(path, yamlBytes, 0o644); err != nil {
		return FileRecord{}, fmt.Errorf("localapi: write file: %w", err)
	}

	updatedAt := time.Now().UTC().Format(time.RFC3339)
	record := FileRecord{
		ID:        id,
		Kind:      kind,
		Name:      name,
		Path:      path,
		UpdatedAt: updatedAt,
	}

	m, err := s.loadManifest()
	if err != nil {
		return FileRecord{}, err
	}

	found := false
	for i, existing := range m.Files {
		if existing.ID == id {
			m.Files[i] = record
			found = true
			break
		}
	}
	if !found {
		m.Files = append(m.Files, record)
	}

	if err := s.saveManifest(m); err != nil {
		return FileRecord{}, err
	}

	return record, nil
}

// PathForID returns the on-disk YAML path for a file id.
func (s *Store) PathForID(id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", errors.New("localapi: file id required")
	}
	path := filepath.Join(s.rootDir, "requests", id+".yaml")
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("localapi: file %q not found", id)
	}
	return path, nil
}

// MarshalDocumentYAML converts a JSON-encoded traCtlSpec document to YAML bytes.
func MarshalDocumentYAML(documentJSON []byte) ([]byte, error) {
	var doc any
	if err := json.Unmarshal(documentJSON, &doc); err != nil {
		return nil, fmt.Errorf("localapi: invalid document json: %w", err)
	}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("localapi: marshal yaml: %w", err)
	}
	return out, nil
}
