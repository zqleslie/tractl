package json

import (
	"github.com/tractl/tractl/internal/validation/common"
)

// checkEncoding validates byte-level encoding and strips a UTF-8 BOM.
// Remaps the generic error code to the JSON-prefixed one.
func checkEncoding(input []byte) ([]byte, []ValidationError) {
	stripped, errs := common.CheckEncoding(input)
	for i := range errs {
		if errs[i].Code == common.ErrEncoding {
			errs[i].Code = ErrEncoding
		}
	}
	return stripped, errs
}

// checkForbiddenSyntax scans the raw bytes for JSON5/JSONC constructs that
// Go's encoding/json does not reject on its own:
//   - // and /* */ comments
//   - single-quoted strings
//
// It does NOT validate structural JSON; that is done by the parser.
// Line numbers are tracked for accurate error positions.
func checkForbiddenSyntax(input []byte) []ValidationError {
	var errs []ValidationError
	line := 1
	i := 0
	n := len(input)

	for i < n {
		b := input[i]

		switch b {
		case '\n':
			line++
			i++

		case '"':
			// Skip over a double-quoted string so we don't misidentify
			// // or ' characters inside string content.
			i++ // consume opening "
			for i < n {
				c := input[i]
				if c == '\\' {
					i += 2 // skip escape sequence
					continue
				}
				if c == '"' {
					i++ // consume closing "
					break
				}
				if c == '\n' {
					line++
				}
				i++
			}

		case '\'':
			// Single-quoted string — forbidden.
			errs = append(errs, ValidationError{
				Code:    ErrSingleQuote,
				Message: "single-quoted strings are not permitted; use double-quoted strings per RFC 8259",
				Line:    line,
				Column:  1,
			})
			// Skip to the closing ' to avoid cascading errors.
			i++
			for i < n {
				c := input[i]
				if c == '\\' {
					i += 2
					continue
				}
				if c == '\'' {
					i++
					break
				}
				if c == '\n' {
					line++
				}
				i++
			}

		case '/':
			if i+1 < n {
				next := input[i+1]
				if next == '/' {
					// Line comment — forbidden.
					errs = append(errs, ValidationError{
						Code:    ErrComment,
						Message: "// line comments are not permitted in traCtl JSON documents (JSONC is not accepted)",
						Line:    line,
						Column:  1,
					})
					// Skip to end of line.
					for i < n && input[i] != '\n' {
						i++
					}
					continue
				}
				if next == '*' {
					// Block comment — forbidden.
					errs = append(errs, ValidationError{
						Code:    ErrComment,
						Message: "/* */ block comments are not permitted in traCtl JSON documents (JSONC is not accepted)",
						Line:    line,
						Column:  1,
					})
					// Skip to closing */.
					i += 2
					for i+1 < n {
						if input[i] == '\n' {
							line++
						}
						if input[i] == '*' && input[i+1] == '/' {
							i += 2
							break
						}
						i++
					}
					continue
				}
			}
			i++

		default:
			i++
		}
	}

	return errs
}

// checkTrailingCommas scans for trailing commas before ] or } tokens.
// A trailing comma is a JSON5 feature and is not permitted.
//
// Strategy: look for a comma that is followed (ignoring whitespace) by ] or }.
// We operate on raw bytes after the forbidden-syntax scan, so string content
// can still contain those bytes — but since we already validated no broken
// strings exist, this produces at most one false positive per malformed doc,
// which is acceptable (the parse failure will also fire).
func checkTrailingCommas(input []byte) []ValidationError {
	var errs []ValidationError
	line := 1
	i := 0
	n := len(input)

	for i < n {
		b := input[i]

		if b == '\n' {
			line++
			i++
			continue
		}

		// Skip double-quoted strings.
		if b == '"' {
			i++
			for i < n {
				c := input[i]
				if c == '\\' {
					i += 2
					continue
				}
				if c == '"' {
					i++
					break
				}
				if c == '\n' {
					line++
				}
				i++
			}
			continue
		}

		if b == ',' {
			commaLine := line
			j := i + 1
			// Skip whitespace and newlines after the comma.
			for j < n && isWhitespace(input[j]) {
				if input[j] == '\n' {
					line++
				}
				j++
			}
			if j < n && (input[j] == ']' || input[j] == '}') {
				errs = append(errs, ValidationError{
					Code:    ErrTrailingComma,
					Message: "trailing commas are not permitted in traCtl JSON documents",
					Line:    commaLine,
					Column:  1,
				})
			}
		}

		i++
	}

	return errs
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}
