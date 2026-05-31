// ulid.go defines the shared ULID generator used by all format parsers to assign unique identifiers to spec entities.
//
// Package common provides shared utilities for traCtl format parsers.
//
// Isolation invariant: this package MUST NOT import any parser package,
// any format validator package, or internal/spec.
package common

import (
	"crypto/rand"
	"fmt"

	"github.com/oklog/ulid/v2"
)

// NewULID generates a new ULID using monotonic time and crypto/rand entropy.
// Each call produces a unique, lexicographically sortable identifier.
// Returns an error if the entropy source is exhausted.
func NewULID() (string, error) {
	id, err := ulid.New(ulid.Now(), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("parser: generate entity ID: %w", err)
	}
	return id.String(), nil
}
