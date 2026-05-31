// parse.go selects and invokes the correct format parser based on file extension or
// explicit format name. Delegates to the parser registry (internal/parser) for
// format dispatch — adding a new format requires only a registry entry, not a change here.

package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tractl/tractl/internal/parser"
	parseryaml "github.com/tractl/tractl/internal/parser/yaml"
	"github.com/tractl/tractl/internal/spec"
)

// SupportedFormats lists format names advertised to external callers (WASM bridge, docs).
// Aliases such as "yml" and empty string are accepted by IsSupportedFormat and ParseByFormat
// but are omitted from this list.
var SupportedFormats = []string{"yaml", "json", "toon"}

// IsSupportedFormat reports whether format can be parsed by ParseByFormat.
func IsSupportedFormat(format string) bool {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "yaml", "yml", "", "json", "toon":
		return true
	default:
		return false
	}
}

// parseByExtension selects the parser based on the file extension.
// Unknown extensions fall back to YAML to preserve CLI behaviour for files
// without a recognised extension (e.g. no extension, .tractl, etc.).
func parseByExtension(src []byte, filePath string) (*spec.TraCtlSpec, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))
	switch ext {
	case "yaml", "yml", "json", "toon", "":
		return ParseByFormat(src, ext, filePath)
	default:
		// Unknown extension: fall back to YAML to preserve existing CLI behaviour.
		return parseryaml.Parse(src, filePath)
	}
}

// ParseByFormat selects the parser for an in-memory document format via the
// parser registry. Exported so surfaces (e.g. WASM bridge) that receive a
// document without going through Run can parse without duplicating format dispatch.
func ParseByFormat(src []byte, format string, sourceRef string) (*spec.TraCtlSpec, error) {
	norm := strings.ToLower(strings.TrimSpace(format))
	// Normalise yml → yaml and empty → yaml to match registry keys.
	if norm == "yml" || norm == "" {
		norm = "yaml"
	}
	p, err := parser.ForFormat(norm)
	if err != nil {
		return nil, fmt.Errorf("unsupported format %q", format)
	}
	return p.Parse(src, sourceRef)
}
