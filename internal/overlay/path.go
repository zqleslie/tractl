package overlay

import (
	"fmt"
	"strings"
)

// pathResult holds the resolved node and enough context to mutate or remove it.
type pathResult struct {
	// parent is the map or slice that owns the node.
	parent any
	// key is the map key or slice-index string of the node inside parent.
	key string
	// value is the resolved node value.
	value any
}

// resolvePath walks a dot-separated canonical path against a map[string]any document
// (the result of JSON-marshaling a *spec.TraCtlSpec) and returns the terminal node.
//
// The walker is purely structural: each segment is looked up as a map key. This is
// correct because JSON unmarshaling of a struct with json tags produces a map whose
// keys are the json tag names — which are the canonical field names defined in the spec.
//
// Returns ErrPathNotFound if any segment is absent. The function is pure; it does not
// mutate the document.
//
// Reference: overlay spec §9.1
func resolvePath(doc map[string]any, path string) (pathResult, error) {
	segments := strings.Split(path, ".")
	if len(segments) == 0 || path == "" {
		return pathResult{}, fmt.Errorf("%w: empty path", errPathNotFound())
	}

	var parent any = doc
	var current any = doc

	for i, seg := range segments {
		m, ok := current.(map[string]any)
		if !ok {
			return pathResult{}, errPathNotFoundf(path, "segment[%d] %q: parent is not an object", i, seg)
		}
		val, exists := m[seg]
		if !exists {
			return pathResult{}, errPathNotFoundf(path, "segment %q not found", seg)
		}
		parent = current
		current = val
	}

	last := segments[len(segments)-1]
	return pathResult{parent: parent, key: last, value: current}, nil
}

// setAtPath writes value at the terminal segment of path inside doc.
// All intermediate segments must already exist. Returns ErrPathNotFound if any
// intermediate segment is absent.
func setAtPath(doc map[string]any, path string, value any) error {
	segments := strings.Split(path, ".")
	if len(segments) == 0 || path == "" {
		return errPathNotFoundf(path, "empty path")
	}

	current := any(doc)
	for _, seg := range segments[:len(segments)-1] {
		m, ok := current.(map[string]any)
		if !ok {
			return errPathNotFoundf(path, "segment %q: parent is not an object", seg)
		}
		next, exists := m[seg]
		if !exists {
			return errPathNotFoundf(path, "segment %q not found", seg)
		}
		current = next
	}

	m, ok := current.(map[string]any)
	if !ok {
		return errPathNotFoundf(path, "terminal parent is not an object")
	}
	m[segments[len(segments)-1]] = value
	return nil
}

// deleteAtPath removes the terminal segment of path from doc.
// Returns ErrPathNotFound if any intermediate segment is absent.
func deleteAtPath(doc map[string]any, path string) error {
	segments := strings.Split(path, ".")
	if len(segments) == 0 || path == "" {
		return errPathNotFoundf(path, "empty path")
	}

	current := any(doc)
	for _, seg := range segments[:len(segments)-1] {
		m, ok := current.(map[string]any)
		if !ok {
			return errPathNotFoundf(path, "segment %q: parent is not an object", seg)
		}
		next, exists := m[seg]
		if !exists {
			return errPathNotFoundf(path, "segment %q not found", seg)
		}
		current = next
	}

	m, ok := current.(map[string]any)
	if !ok {
		return errPathNotFoundf(path, "terminal parent is not an object")
	}
	delete(m, segments[len(segments)-1])
	return nil
}

func errPathNotFound() error {
	return &EngineError{Code: ErrPathNotFound, Detail: "path not found"}
}

func errPathNotFoundf(path, format string, args ...any) error {
	return &EngineError{
		Code:   ErrPathNotFound,
		Target: path,
		Detail: fmt.Sprintf(format, args...),
	}
}
