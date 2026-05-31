// format_validator.go defines the FormatValidator interface and the ValidatorFor
// factory function used to select a format validator by name.
//
// Design note — why this file lives in internal/validation, not
// internal/validation/common:
//
// The three format-validator packages (yaml, json, toon) all import
// internal/validation/common for shared types (ValidationResult, ValidationError,
// ErrorCode). Placing ValidatorFor in common and importing the format packages
// from there would create a hard import cycle:
//
//	common → yaml → common   (cycle, build fails)
//
// Placing the factory here (internal/validation) avoids the cycle because the
// format packages do NOT import internal/validation — only common.
//
// COORDINATION(Stream C): Replace the format-selector switch in
// internal/engine/ (or overlay/parse.go after split) with:
//
//	v, err := validation.ValidatorFor(format)
//	if err != nil { return nil, err }
//	result := v.Validate(data)
//	if !result.Valid { ... }
package validation

import (
	"fmt"
	"sort"

	validcommon "github.com/tractl/tractl/internal/validation/common"
	jsonvalidator "github.com/tractl/tractl/internal/validation/json"
	toonvalidator "github.com/tractl/tractl/internal/validation/toon"
	yamlvalidator "github.com/tractl/tractl/internal/validation/yaml"
)

// FormatValidator is the common interface satisfied by all format-specific
// validators. Each format validator (yaml.Validate, json.Validate, toon.Validate)
// is adapted to this interface via validatorFunc — no changes to those packages
// are required.
//
// Validate checks raw document bytes for format-specific schema and structural
// errors. It returns a ValidationResult describing all problems found. A Valid
// result means the document conforms to the format subset rules.
type FormatValidator interface {
	Validate(data []byte) *validcommon.ValidationResult
}

// validatorFunc adapts a package-level Validate function to the FormatValidator
// interface. The format validators expose Validate as a package-level function
// (not a method on a struct), so this adapter bridges the gap without requiring
// changes to those packages.
type validatorFunc func([]byte) *validcommon.ValidationResult

func (f validatorFunc) Validate(data []byte) *validcommon.ValidationResult {
	return f(data)
}

// supportedFormats is the registry of all known format validators, initialized
// once at package startup. To add a new format: add one entry here only.
var supportedFormats = map[string]FormatValidator{
	"yaml": validatorFunc(yamlvalidator.Validate),
	"json": validatorFunc(jsonvalidator.Validate),
	"toon": validatorFunc(toonvalidator.Validate),
}

// ValidatorFor returns the FormatValidator for the given format string.
// Supported format values: "yaml", "json", "toon".
// Returns an error for any unrecognised or empty format string.
func ValidatorFor(format string) (FormatValidator, error) {
	v, ok := supportedFormats[format]
	if !ok {
		return nil, fmt.Errorf("validation: unsupported format %q", format)
	}
	return v, nil
}

// SupportedFormats returns a sorted list of all registered format names.
// Useful for error messages and CLI help text.
func SupportedFormats() []string {
	formats := make([]string, 0, len(supportedFormats))
	for f := range supportedFormats {
		formats = append(formats, f)
	}
	sort.Strings(formats)
	return formats
}

// Compile-time interface satisfaction checks.
// These verify that each format package's Validate function has exactly
// the signature required by the validatorFunc adapter, and therefore by
// FormatValidator. A signature drift in any format package will cause a
// compile error here before any tests run.
var (
	_ FormatValidator = validatorFunc(yamlvalidator.Validate)
	_ FormatValidator = validatorFunc(jsonvalidator.Validate)
	_ FormatValidator = validatorFunc(toonvalidator.Validate)
)
