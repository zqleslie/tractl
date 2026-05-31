// scalar.go implements YAML scalar value checks enforced during the AST walk:
// YAML 1.1 boolean prohibition, ambiguous null forms, implicit timestamp
// detection, and prohibited numeric forms. Reference: tractl_yaml_spec.md §7.
package yaml

import (
	"regexp"
	"strings"

	goyaml "gopkg.in/yaml.v3"
)

// yaml11BooleanValues is the set of YAML 1.1 boolean literals that MUST NOT
// be coerced to booleans in traCtl YAML. Only "true" and "false" (lowercase)
// are valid. yaml.v3 resolves these to !!bool so we check node.Value, which
// holds the raw authored string.
//
// Spec ref: §7.4.
var yaml11BooleanValues = map[string]bool{
	// Traditional yes/no variants
	"yes": true, "Yes": true, "YES": true,
	"no": true, "No": true, "NO": true,
	// On/off variants
	"on": true, "On": true, "ON": true,
	"off": true, "Off": true, "OFF": true,
	// Single-letter variants
	"y": true, "Y": true,
	"n": true, "N": true,
	// Mixed-case True/False (yaml.v3 resolves these as !!bool)
	"True": true, "TRUE": true,
	"False": true, "FALSE": true,
}

// prohibitedNullValues is the set of YAML null forms that are prohibited as
// typed nulls in traCtl YAML.
//   - "~" is always rejected as ambiguous.
//   - "Null" and "NULL" must be double-quoted if a string value is intended.
//   - "" (empty scalar) is rejected; use "null" explicitly or omit the field.
//
// The canonical "null" (all-lowercase) is accepted.
// Spec ref: §7.5.
var prohibitedNullValues = map[string]bool{
	"~":    true,
	"Null": true,
	"NULL": true,
	// Empty value position: key: <nothing>
	// yaml.v3 resolves this as !!null with Value == "".
	"": true,
}

// prohibitedNumericPrefixRe matches octal (0o/0O), hexadecimal (0x/0X), and
// binary (0b/0B) numeric prefixes that are forbidden in the traCtl subset.
// Spec ref: §7.6.
var prohibitedNumericPrefixRe = regexp.MustCompile(`^[+-]?0[oObBxX]`)

// leadingZeroIntRe matches integer literals with superfluous leading zeros
// (e.g. 007, 0123). JSON integers may not have leading zeros.
// Spec ref: §7.6.
var leadingZeroIntRe = regexp.MustCompile(`^[+-]?0\d`)

// underscoredNumericRe matches underscored numeric literals (e.g. 1_000_000).
// Spec ref: §7.6.
var underscoredNumericRe = regexp.MustCompile(`\d_\d`)

// sexagesimalRe matches YAML 1.1 sexagesimal (base-60) integer literals
// (e.g. 12:34:56). These are a known YAML 1.1 holdover.
// Spec ref: §7.6.
var sexagesimalRe = regexp.MustCompile(`^\d+:\d{2}(:\d{2})*$`)

// specialFloatValues is the set of YAML special float literals that are not
// representable in the JSON number grammar. All must be rejected.
// Spec ref: §7.6.
var specialFloatValues = map[string]bool{
	".inf": true, ".Inf": true, ".INF": true,
	"-.inf": true, "-.Inf": true, "-.INF": true,
	"+.inf": true, "+.Inf": true, "+.INF": true,
	".nan": true, ".NaN": true, ".NAN": true,
	// Some YAML libraries also emit these forms:
	"inf": true, "Inf": true, "INF": true,
	"nan": true, "NaN": true, "NAN": true,
	"Infinity": true, "-Infinity": true, "+Infinity": true,
}

// validateScalar inspects a scalar YAML node for prohibited value forms.
// It is called by the AST walker after structural checks (anchors, tags,
// quoting style) have already been applied to the same node.
func validateScalar(node *goyaml.Node) []ValidationError {
	var errs []ValidationError

	switch node.Tag {
	case "!!bool":
		// yaml.v3 resolves yes/no/on/off/True/False/etc. to !!bool.
		// node.Value holds the raw string the author wrote.
		if yaml11BooleanValues[node.Value] {
			errs = append(errs, ValidationError{
				Code: ErrYAML11Boolean,
				Message: "YAML 1.1 boolean variant \"" + node.Value + "\" is not permitted; " +
					"use \"true\" or \"false\" (all-lowercase only)",
				Line:   node.Line,
				Column: node.Column,
			})
		}

	case "!!null":
		if prohibitedNullValues[node.Value] {
			msg := prohibitedNullMessage(node.Value)
			errs = append(errs, ValidationError{
				Code:    ErrProhibitedNull,
				Message: msg,
				Line:    node.Line,
				Column:  node.Column,
			})
		}

	case "!!timestamp":
		// yaml.v3 implicitly resolves date/time-looking strings to !!timestamp,
		// even without an explicit tag. This is YAML 1.1 behaviour. We reject
		// any scalar the library resolves this way without an explicit tag
		// (explicit tags are caught separately in the walker).
		//
		// Implementation note from spec §7.7: "Many widely-used YAML libraries
		// apply YAML 1.1 implicit timestamp coercion by default."
		if node.Style&goyaml.TaggedStyle == 0 {
			errs = append(errs, ValidationError{
				Code: ErrImplicitTimestamp,
				Message: "implicit timestamp coercion is not permitted for value \"" + node.Value + "\"; " +
					"double-quote date/time strings or represent them through canonical traCtlSpec field schemas",
				Line:   node.Line,
				Column: node.Column,
			})
		}

	case "!!str":
		// yaml.v3 resolves yes/no/on/off/Y/N (and case variants) as !!str under
		// the YAML 1.2 core schema, so they never reach the !!bool branch above.
		// A plain (unquoted) scalar that looks like a YAML 1.1 boolean must be
		// rejected. Quoted values (e.g. "yes") are an explicit authoring choice
		// and are allowed.
		if node.Style == 0 && yaml11BooleanValues[node.Value] {
			errs = append(errs, ValidationError{
				Code: ErrYAML11Boolean,
				Message: "YAML 1.1 boolean variant \"" + node.Value + "\" is not permitted; " +
					"use \"true\" or \"false\" (all-lowercase only)",
				Line:   node.Line,
				Column: node.Column,
			})
		}

	case "!!int":
		errs = append(errs, validateIntScalar(node)...)

	case "!!float":
		errs = append(errs, validateFloatScalar(node)...)
	}

	return errs
}

// validateIntScalar checks an integer-resolved scalar for prohibited numeric forms.
func validateIntScalar(node *goyaml.Node) []ValidationError {
	v := node.Value

	if prohibitedNumericPrefixRe.MatchString(v) {
		return []ValidationError{{
			Code: ErrProhibitedNumber,
			Message: "prohibited numeric literal \"" + v + "\"; " +
				"octal (0o…), hexadecimal (0x…), and binary (0b…) literals are not permitted — use decimal",
			Line:   node.Line,
			Column: node.Column,
		}}
	}

	if strings.HasPrefix(v, "+") {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "positive-sign prefix on numeric literal \"" + v + "\" is not permitted",
			Line:    node.Line,
			Column:  node.Column,
		}}
	}

	if leadingZeroIntRe.MatchString(v) {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "leading zero in integer literal \"" + v + "\" is not permitted",
			Line:    node.Line,
			Column:  node.Column,
		}}
	}

	if underscoredNumericRe.MatchString(v) {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "underscored numeric literal \"" + v + "\" is not permitted; use plain decimal digits",
			Line:    node.Line,
			Column:  node.Column,
		}}
	}

	if sexagesimalRe.MatchString(v) {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "sexagesimal numeric literal \"" + v + "\" is a YAML 1.1 holdover and is not permitted",
			Line:    node.Line,
			Column:  node.Column,
		}}
	}

	return nil
}

// validateFloatScalar checks a float-resolved scalar for prohibited numeric forms.
func validateFloatScalar(node *goyaml.Node) []ValidationError {
	v := node.Value

	if specialFloatValues[v] {
		return []ValidationError{{
			Code: ErrProhibitedNumber,
			Message: "special float literal \"" + v + "\" is not permitted; " +
				".inf and .nan are not representable in the JSON number grammar",
			Line:   node.Line,
			Column: node.Column,
		}}
	}

	if strings.HasPrefix(v, "+") {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "positive-sign prefix on numeric literal \"" + v + "\" is not permitted",
			Line:    node.Line,
			Column:  node.Column,
		}}
	}

	if underscoredNumericRe.MatchString(v) {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "underscored numeric literal \"" + v + "\" is not permitted",
			Line:    node.Line,
			Column:  node.Column,
		}}
	}

	return nil
}

// prohibitedNullMessage returns a tailored diagnostic message for each
// prohibited null form, helping authors understand what to do instead.
func prohibitedNullMessage(value string) string {
	switch value {
	case "~":
		return "\"~\" is an ambiguous YAML null shorthand and is not permitted; use \"null\" (lowercase) or omit the field"
	case "Null":
		return "\"Null\" is ambiguous across YAML versions; use \"null\" (lowercase) for a null value, or double-quote \"Null\" if you intend a string"
	case "NULL":
		return "\"NULL\" is ambiguous across YAML versions; use \"null\" (lowercase) for a null value, or double-quote \"NULL\" if you intend a string"
	case "":
		return "empty value position (key:) is not permitted; use \"null\" explicitly or omit the field entirely"
	default:
		return "prohibited null variant \"" + value + "\"; use \"null\" (lowercase)"
	}
}
