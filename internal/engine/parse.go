// parse.go selects and invokes the correct format parser based on file extension or
// explicit format name. Format detection uses a switch over the three supported formats
// (yaml, json, toon). When Stream A's parser factory lands, this file becomes a thin
// wrapper over parser.FormatFor().

package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	parserjson "github.com/tractl/tractl/internal/parser/json"
	parsertoon "github.com/tractl/tractl/internal/parser/toon"
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
		return parseryaml.Parse(src, filePath)
	}
}

// ParseByFormat selects the parser for an in-memory document format.
// Exported so surfaces (e.g. WASM bridge) that receive a document without going
// through Run can parse without duplicating the format-dispatch switch.
func ParseByFormat(src []byte, format string, sourceRef string) (*spec.TraCtlSpec, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "yaml", "yml", "":
		return parseryaml.Parse(src, sourceRef)
	case "json":
		return parserjson.Parse(src, sourceRef)
	case "toon":
		return parsertoon.Parse(src, sourceRef)
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}
