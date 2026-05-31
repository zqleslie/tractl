package json

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/tractl/tractl/internal/validation/common"
)

// prohibitedNumberRe matches number-like tokens that are forbidden in traCtl JSON.
// These are caught by scanning the raw token string from the decoder.
var prohibitedNumberPrefixRe = regexp.MustCompile(`^[+-]?0[xXoObB]`)
var leadingZeroRe = regexp.MustCompile(`^-?0\d`)
var underscoredRe = regexp.MustCompile(`\d_\d`)

// walkDocument parses the JSON input using a streaming decoder and validates:
//   - top-level value is an object
//   - no multi-value stream
//   - no duplicate keys (recursively)
//   - no _ulid keys or reserved-prefix keys
//   - no prohibited number forms
func walkDocument(input []byte) []ValidationError {
	dec := json.NewDecoder(strings.NewReader(string(input)))
	dec.UseNumber() // give us raw number tokens so we can inspect their form

	var errs []ValidationError

	// Read the first token — must be '{' (object start).
	tok, err := dec.Token()
	if err != nil {
		errs = append(errs, ValidationError{
			Code:    ErrParseFailed,
			Message: "JSON parse error: " + err.Error(),
		})
		return errs
	}

	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		errs = append(errs, ValidationError{
			Code:    ErrNotObject,
			Message: "the top-level value in a traCtl JSON document must be an object ({…}), not " + fmt.Sprintf("%T", tok),
		})
		return errs
	}

	errs = append(errs, walkObject(dec)...)

	// Check for a second top-level value (multi-value stream).
	tok2, err2 := dec.Token()
	if err2 == nil || (err2 != nil && err2 != io.EOF) {
		msg := "multi-value JSON streams (NDJSON, concatenated values) are not permitted; a traCtl JSON document must contain exactly one top-level object"
		if err2 == nil {
			msg += fmt.Sprintf(" (unexpected token after document: %v)", tok2)
		}
		errs = append(errs, ValidationError{
			Code:    ErrMultiValue,
			Message: msg,
		})
	}

	return errs
}

// walkObject consumes an already-opened '{' and validates all key/value pairs.
// It recurses into nested objects and arrays.
func walkObject(dec *json.Decoder) []ValidationError {
	var errs []ValidationError
	seen := make(map[string]bool)

	for dec.More() {
		// Read the key token.
		keyTok, err := dec.Token()
		if err != nil {
			errs = append(errs, ValidationError{
				Code:    ErrParseFailed,
				Message: "JSON parse error reading object key: " + err.Error(),
			})
			return errs
		}

		key, ok := keyTok.(string)
		if !ok {
			errs = append(errs, ValidationError{
				Code:    ErrParseFailed,
				Message: fmt.Sprintf("expected string key, got %T", keyTok),
			})
			return errs
		}

		// Duplicate key check.
		if seen[key] {
			errs = append(errs, ValidationError{
				Code:    ErrDuplicateKey,
				Message: "duplicate object key \"" + key + "\"",
			})
		}
		seen[key] = true

		// _ulid / reserved prefix checks (remap generic codes to JSON codes).
		for _, ce := range common.CheckKey(key, 0, 0) {
			switch ce.Code {
			case common.ErrULIDKey:
				ce.Code = ErrULIDKey
			case common.ErrReservedPrefix:
				ce.Code = ErrReservedPrefix
			case common.ErrEncoding, common.ErrDuplicateKey:
				// CheckKey never produces these codes; handled elsewhere
			}
			errs = append(errs, ce)
		}

		// Read the value.
		errs = append(errs, walkValue(dec)...)
	}

	// Consume the closing '}'.
	if _, err := dec.Token(); err != nil {
		errs = append(errs, ValidationError{
			Code:    ErrParseFailed,
			Message: "JSON parse error reading object close: " + err.Error(),
		})
	}

	return errs
}

// walkArray consumes an already-opened '[' and validates all elements.
func walkArray(dec *json.Decoder) []ValidationError {
	var errs []ValidationError

	for dec.More() {
		errs = append(errs, walkValue(dec)...)
	}

	if _, err := dec.Token(); err != nil {
		errs = append(errs, ValidationError{
			Code:    ErrParseFailed,
			Message: "JSON parse error reading array close: " + err.Error(),
		})
	}

	return errs
}

// walkValue reads and validates a single JSON value (object, array, or scalar).
func walkValue(dec *json.Decoder) []ValidationError {
	tok, err := dec.Token()
	if err != nil {
		return []ValidationError{{
			Code:    ErrParseFailed,
			Message: "JSON parse error reading value: " + err.Error(),
		}}
	}

	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			return walkObject(dec)
		case '[':
			return walkArray(dec)
		}

	case json.Number:
		return validateNumber(string(v))
	}

	return nil
}

// validateNumber checks a raw JSON number token for prohibited forms.
func validateNumber(v string) []ValidationError {
	// Radix prefix (0x…, 0o…, 0b…)
	if prohibitedNumberPrefixRe.MatchString(v) {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "prohibited numeric literal \"" + v + "\"; octal (0o…), hexadecimal (0x…), and binary (0b…) literals are not permitted",
		}}
	}

	// Positive sign prefix
	if strings.HasPrefix(v, "+") {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "positive-sign prefix on numeric literal \"" + v + "\" is not permitted",
		}}
	}

	// Leading zeros (e.g. 0123) — but not 0 itself or 0.x forms
	if leadingZeroRe.MatchString(v) {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "leading zero in integer literal \"" + v + "\" is not permitted",
		}}
	}

	// Underscored numeric
	if underscoredRe.MatchString(v) {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "underscored numeric literal \"" + v + "\" is not permitted",
		}}
	}

	// Missing integer part (.5) or missing fractional part (5.)
	if strings.HasPrefix(v, ".") {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "numeric literal \"" + v + "\" has no integer part; leading decimal point is not permitted",
		}}
	}
	if strings.HasSuffix(v, ".") {
		return []ValidationError{{
			Code:    ErrProhibitedNumber,
			Message: "numeric literal \"" + v + "\" has no fractional part after decimal point",
		}}
	}

	return nil
}
