// parse.go defines Parse, the format-agnostic overlay document decoder.
// Dispatches to YAML or JSON unmarshal based on the format argument.
// Format validation is the caller's responsibility before calling Parse.
package overlay

import (
	"encoding/json"
	"fmt"
	"strings"

	goyaml "gopkg.in/yaml.v3"
)

// Parse decodes raw overlay bytes in the given format into an *OverlayDocument.
//
// Supported formats: "yaml", "json", "toon". Format is case-insensitive.
//
// Parse does NOT run format validation (internal/validation/{yaml,json,toon}).
// Callers must invoke the appropriate format validator before calling Parse so that
// subset-compliance errors are reported before the unmarshal step.
// (TODO Stream C: call validyaml/validjson/validtoon.Validate at the engine call site
// before passing bytes to Parse.)
//
// Parse does NOT run overlay document validation — that is internal/validation.OverlayValidator's
// responsibility. The caller must run OverlayValidator.Validate on the returned document
// before passing it to Engine.Apply.
//
// Cross-format authoring is first-class per overlay spec §4 and §5: an overlay may be
// authored in any supported format regardless of the source workflow format.
func Parse(src []byte, format string) (*OverlayDocument, error) {
	switch strings.ToLower(format) {
	case "yaml":
		return parseYAML(src)
	case "json":
		return parseJSON(src)
	case "toon":
		return parseTOON(src)
	default:
		return nil, fmt.Errorf("unsupported overlay format %q: must be yaml, json, or toon", format)
	}
}

func parseYAML(src []byte) (*OverlayDocument, error) {
	// Direct unmarshal — format validation is the caller's responsibility (see Parse doc).
	var doc OverlayDocument
	if err := goyaml.Unmarshal(src, &doc); err != nil {
		return nil, fmt.Errorf("overlay yaml unmarshal failed: %w", err)
	}
	return &doc, nil
}

func parseJSON(src []byte) (*OverlayDocument, error) {
	// Direct unmarshal — overlay documents are always JSON-serialisable; format
	// validation is the caller's responsibility (see Parse doc).
	var doc OverlayDocument
	if err := json.Unmarshal(src, &doc); err != nil {
		return nil, fmt.Errorf("overlay json unmarshal failed: %w", err)
	}
	return &doc, nil
}

func parseTOON(src []byte) (*OverlayDocument, error) {
	// TOON is a strict YAML subset; goyaml unmarshal works correctly.
	// Format validation is the caller's responsibility (see Parse doc).
	var doc OverlayDocument
	if err := goyaml.Unmarshal(src, &doc); err != nil {
		return nil, fmt.Errorf("overlay toon unmarshal failed: %w", err)
	}
	return &doc, nil
}
