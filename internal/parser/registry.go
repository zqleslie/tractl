// registry.go defines the Parser interface and the ForFormat factory function.
// Callers select a parser by format name without importing each format package
// concretely.
//
// To add a new format: add one entry to the parsers map — no other file needs
// to change.
//
// COORDINATION(Stream C): Replace the format parser switch in
// internal/engine/engine.go (parseByFormat / parseByExtension, or equivalent)
// with:
//
//	p, err := parser.ForFormat(format)
//	if err != nil { return nil, err }
//	spec, err := p.Parse(data, sourceRef)
//
// After switching, remove direct imports of internal/parser/yaml,
// internal/parser/json, internal/parser/toon from the engine.
package parser

import (
	"fmt"
	"sort"
	"strings"

	parserjson "github.com/tractl/tractl/internal/parser/json"
	parsertoon "github.com/tractl/tractl/internal/parser/toon"
	parseryaml "github.com/tractl/tractl/internal/parser/yaml"
	"github.com/tractl/tractl/internal/spec"
)

// Parser is the common interface satisfied by all format-specific parsers.
// It is declared here so callers can select a parser by format name without
// importing each format package concretely.
//
// Each format parser (yaml, json, toon) exposes a package-level Parse function
// adapted to this interface via parserFunc — no changes to those packages are
// needed.
type Parser interface {
	// Parse converts raw document bytes into a *spec.TraCtlSpec.
	// sourceRef is an opaque string identifying the document origin (file path,
	// URL, or similar) stored in the returned spec's metadata.
	// Returns a non-nil spec on success, or a descriptive error if the input is
	// malformed, fails format validation, or cannot be decoded.
	Parse(src []byte, sourceRef string) (*spec.TraCtlSpec, error)
}

// parserFunc adapts a package-level Parse function to the Parser interface.
// The format parsers expose Parse as a package-level function (not a method on
// a struct), so this adapter bridges the gap without requiring changes to those
// packages.
type parserFunc func([]byte, string) (*spec.TraCtlSpec, error)

func (f parserFunc) Parse(src []byte, sourceRef string) (*spec.TraCtlSpec, error) {
	return f(src, sourceRef)
}

// parsers is the registry of all supported format parsers, initialised once at
// package startup. Add new formats here only — callers need no changes.
var parsers = map[string]Parser{
	"yaml": parserFunc(parseryaml.Parse),
	"json": parserFunc(parserjson.Parse),
	"toon": parserFunc(parsertoon.Parse),
}

// ForFormat returns the Parser for the given format string.
// Supported format values: "yaml", "json", "toon".
// Returns a descriptive error for any unrecognised or empty format string.
func ForFormat(format string) (Parser, error) {
	p, ok := parsers[format]
	if !ok {
		return nil, fmt.Errorf("parser: unsupported format %q (supported: %s)",
			format, strings.Join(SupportedFormats(), ", "))
	}
	return p, nil
}

// SupportedFormats returns a sorted list of all registered format names.
// Useful for error messages and CLI help text.
func SupportedFormats() []string {
	formats := make([]string, 0, len(parsers))
	for f := range parsers {
		formats = append(formats, f)
	}
	sort.Strings(formats)
	return formats
}

// Compile-time interface satisfaction checks.
// These verify that each format package's Parse function has exactly the
// signature required by parserFunc, and therefore by Parser. A signature
// change in any format package will cause a compile error here before any
// test runs.
var (
	_ Parser = parserFunc(parseryaml.Parse)
	_ Parser = parserFunc(parserjson.Parse)
	_ Parser = parserFunc(parsertoon.Parse)
)
