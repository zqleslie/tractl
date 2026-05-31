// keys.go defines shared key validation helpers used by the YAML, JSON, and
// TOON AST walkers: _ulid authoring prohibition, reserved prefix enforcement,
// and duplicate key detection.
package common

import "strings"

// ErrULIDKey is produced when a mapping key named "_ulid" appears in an
// authored document. _ulid is system-assigned and must never be authored.
// Spec ref: tractl_spec.md §4.2.
const ErrULIDKey ErrorCode = "KEY_ULID_AUTHORED"

// ErrReservedPrefix is produced when a mapping key uses a reserved
// namespace prefix (_traCtl. or traCtl.).
// Spec ref: tractl_spec.md §4.1.
const ErrReservedPrefix ErrorCode = "KEY_RESERVED_PREFIX"

// ErrDuplicateKey is produced when a mapping contains two or more entries
// with the same string key.
const ErrDuplicateKey ErrorCode = "KEY_DUPLICATE"

// reservedKeyPrefixes are the namespace prefixes that must not appear in
// authored documents across all formats.
var reservedKeyPrefixes = []string{
	"_traCtl.",
	"traCtl.",
}

// CheckKey inspects a single mapping key string and returns any key-level
// violations: ULID key, reserved prefix. Duplicate-key detection is
// format-specific and must be handled by each validator.
func CheckKey(key string, line, col int) []ValidationError {
	var errs []ValidationError

	if key == "_ulid" {
		errs = append(errs, ValidationError{
			Code:    ErrULIDKey,
			Message: "\"_ulid\" is a system-assigned identity field and must not appear in authored documents",
			Line:    line,
			Column:  col,
		})
	}

	for _, prefix := range reservedKeyPrefixes {
		if strings.HasPrefix(key, prefix) {
			errs = append(errs, ValidationError{
				Code:    ErrReservedPrefix,
				Message: "key \"" + key + "\" uses the reserved namespace prefix \"" + prefix + "\"",
				Line:    line,
				Column:  col,
			})
			break
		}
	}

	return errs
}

// Truncate returns s capped at n runes with a trailing ellipsis if truncated.
// Used for embedding raw document fragments in error messages.
func Truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
