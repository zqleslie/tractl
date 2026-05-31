// lexer.go implements raw-byte pre-checks for the TOON validator: UTF-8
// encoding, tab indentation rejection, YAML directive rejection, document
// separator rejection, and invalid escape sequence detection in double-quoted strings.
package toon

import (
	"bytes"
	"strings"

	"github.com/tractl/tractl/internal/validation/common"
)

// checkEncoding delegates to the shared encoding checker, remapping the
// generic error code to the TOON-prefixed one.
func checkEncoding(input []byte) ([]byte, []ValidationError) {
	stripped, errs := common.CheckEncoding(input)
	for i := range errs {
		if errs[i].Code == common.ErrEncoding {
			errs[i].Code = ErrEncoding
		}
	}
	return stripped, errs
}

// normalizeLineEndings replaces CRLF and bare CR with LF.
// Spec ref: §6.2.
func normalizeLineEndings(input []byte) []byte {
	return common.NormalizeLineEndings(input)
}

// checkTabIndentation scans for tab characters in indentation positions.
// TOON requires 2-space indentation; tabs are a lexical error.
// Spec ref: §6.3.
func checkTabIndentation(input []byte) []ValidationError {
	var errs []ValidationError
	lineNum := 1

	for _, line := range bytes.Split(input, []byte{'\n'}) {
		for col, b := range line {
			switch b {
			case '\t':
				errs = append(errs, ValidationError{
					Code:    ErrTabIndentation,
					Message: "tab character in indentation position; TOON requires 2-space indentation",
					Line:    lineNum,
					Column:  col + 1,
				})
				goto nextLine
			case ' ':
				continue
			default:
				goto nextLine
			}
		}
	nextLine:
		lineNum++
	}

	return errs
}

// checkDocumentSeparators scans for YAML document markers (--- and ...).
// TOON documents are single-document and must not contain these markers.
// Spec ref: §20.
func checkDocumentSeparators(input []byte) []ValidationError {
	var errs []ValidationError
	lineNum := 1

	for _, line := range bytes.Split(input, []byte{'\n'}) {
		trimmed := bytes.TrimRight(line, " \t\r")
		if bytes.Equal(trimmed, []byte("---")) || bytes.Equal(trimmed, []byte("...")) {
			marker := string(trimmed)
			errs = append(errs, ValidationError{
				Code:    ErrDocumentSeparator,
				Message: "TOON documents are single-document; the \"" + marker + "\" document marker is not permitted",
				Line:    lineNum,
				Column:  1,
			})
		}
		lineNum++
	}

	return errs
}

// checkDirectives scans for YAML directive lines (%YAML, %TAG) in the
// document preamble. YAML directives are not applicable to TOON.
// Spec ref: §20.
func checkDirectives(input []byte) []ValidationError {
	var errs []ValidationError
	lineNum := 1

	for _, line := range bytes.Split(input, []byte{'\n'}) {
		trimmed := bytes.TrimLeft(line, " \t")
		if len(trimmed) == 0 || trimmed[0] == '#' {
			lineNum++
			continue
		}
		if trimmed[0] == '%' {
			errs = append(errs, ValidationError{
				Code:    ErrDirective,
				Message: "YAML directives are not applicable to TOON: " + common.Truncate(string(trimmed), 48),
				Line:    lineNum,
				Column:  bytes.IndexByte(line, '%') + 1,
			})
		} else {
			// Non-blank, non-comment, non-directive content reached.
			break
		}
		lineNum++
	}

	return errs
}

// checkInvalidEscapes scans double-quoted string content for escape sequences
// not in the TOON-permitted set: \", \\, \n, \r, \t, \uXXXX.
// Any other \X sequence is a lexical error per spec §7.1.1.
func checkInvalidEscapes(input []byte) []ValidationError {
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
		if b != '"' {
			i++
			continue
		}

		// Inside a double-quoted string.
		i++ // consume opening "
		for i < n {
			c := input[i]
			if c == '\n' {
				line++
				i++
				continue
			}
			if c == '"' {
				i++ // consume closing "
				break
			}
			if c == '\\' {
				if i+1 >= n {
					i++
					break
				}
				next := input[i+1]
				switch next {
				case '"', '\\', 'n', 'r', 't':
					i += 2
				case 'u':
					// \uXXXX — require exactly 4 hex digits.
					if i+5 < n && isHexDigits(input[i+2:i+6]) {
						i += 6
					} else {
						errs = append(errs, ValidationError{
							Code:    ErrInvalidEscape,
							Message: `unrecognized escape sequence "\` + string(next) + `" in double-quoted string; recognized escapes are \", \\, \n, \r, \t, \uXXXX`,
							Line:    line,
							Column:  1,
						})
						i += 2
					}
				default:
					errs = append(errs, ValidationError{
						Code:    ErrInvalidEscape,
						Message: `unrecognized escape sequence "\` + string(next) + `" in double-quoted string; recognized escapes are \", \\, \n, \r, \t, \uXXXX`,
						Line:    line,
						Column:  1,
					})
					i += 2
				}
				continue
			}
			i++
		}
	}

	return errs
}

func isHexDigits(b []byte) bool {
	for _, c := range b {
		if !strings.ContainsRune("0123456789abcdefABCDEF", rune(c)) {
			return false
		}
	}
	return true
}
