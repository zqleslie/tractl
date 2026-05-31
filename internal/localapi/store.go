package localapi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Store points handlers at a git-native workspace root.
type Store struct {
	rootDir string
}

// NewStore creates the workspace directory skeleton if it is missing.
func NewStore(rootDir string) (*Store, error) {
	if rootDir == "" {
		return nil, errors.New("localapi: empty root directory")
	}
	for _, dir := range []string{
		filepath.Join(rootDir, "requests"),
		filepath.Join(rootDir, "workflows"),
		filepath.Join(rootDir, "collections"),
		filepath.Join(rootDir, "overlays"),
		filepath.Join(rootDir, ".tractl"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("localapi: mkdir %s: %w", filepath.Base(dir), err)
		}
	}
	return &Store{rootDir: rootDir}, nil
}

// DefaultRootDir returns the default git-native workspace directory.
func DefaultRootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "tractl-workspace"), nil
}

func (s *Store) root() string {
	return s.rootDir
}

func (s *Store) resolvePath(rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" || filepath.IsAbs(rel) {
		return "", errors.New("localapi: relative file path required")
	}
	clean := filepath.Clean(rel)
	if clean == "." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", errors.New("localapi: file path escapes workspace")
	}
	path := filepath.Join(s.rootDir, clean)
	root, err := filepath.Abs(s.rootDir)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if absPath != root && !strings.HasPrefix(absPath, root+string(filepath.Separator)) {
		return "", errors.New("localapi: file path escapes workspace")
	}
	return absPath, nil
}
