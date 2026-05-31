package overlay

import (
	"encoding/json"
	"fmt"
	"strings"

	goyaml "gopkg.in/yaml.v3"

	validjson "github.com/tractl/tractl/internal/validation/json"
	validtoon "github.com/tractl/tractl/internal/validation/toon"
	validyaml "github.com/tractl/tractl/internal/validation/yaml"
)

// Parse decodes raw overlay bytes in the given format into an *OverlayDocument.
//
// Supported formats: "yaml", "json", "toon". Format is case-insensitive.
//
// Before unmarshaling, the appropriate format validator from internal/validation is run.
// If format validation fails, Parse returns a descriptive error without attempting
// to unmarshal.
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
	result := validyaml.Validate(src)
	if !result.Valid {
		msgs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			msgs[i] = e.Message
		}
		return nil, fmt.Errorf("overlay yaml validation failed: %s", strings.Join(msgs, "; "))
	}

	var doc OverlayDocument
	if err := goyaml.Unmarshal(src, &doc); err != nil {
		return nil, fmt.Errorf("overlay yaml unmarshal failed: %w", err)
	}
	return &doc, nil
}

func parseJSON(src []byte) (*OverlayDocument, error) {
	result := validjson.Validate(src)
	if !result.Valid {
		msgs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			msgs[i] = e.Message
		}
		return nil, fmt.Errorf("overlay json validation failed: %s", strings.Join(msgs, "; "))
	}

	var doc OverlayDocument
	if err := json.Unmarshal(src, &doc); err != nil {
		return nil, fmt.Errorf("overlay json unmarshal failed: %w", err)
	}
	return &doc, nil
}

func parseTOON(src []byte) (*OverlayDocument, error) {
	result := validtoon.Validate(src)
	if !result.Valid {
		msgs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			msgs[i] = e.Message
		}
		return nil, fmt.Errorf("overlay toon validation failed: %s", strings.Join(msgs, "; "))
	}

	// TOON is a strict YAML subset; UnmarshalYAMLBytes (goyaml) works correctly.
	var doc OverlayDocument
	if err := goyaml.Unmarshal(src, &doc); err != nil {
		return nil, fmt.Errorf("overlay toon unmarshal failed: %w", err)
	}
	return &doc, nil
}
