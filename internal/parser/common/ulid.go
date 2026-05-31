// Package common provides shared utilities for traCtl format parsers.
//
// Isolation invariant: this package MUST NOT import any parser package,
// any format validator package, or internal/spec.
package common

import (
	"crypto/rand"

	"github.com/oklog/ulid/v2"
)

// NewULID generates a new ULID using monotonic time and crypto/rand entropy.
// Each call produces a unique, lexicographically sortable identifier.
func NewULID() string {
	return ulid.MustNew(ulid.Now(), rand.Reader).String()
}
