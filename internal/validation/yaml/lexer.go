// lexer.go implements the raw-byte lexical pre-checks for the YAML validator:
// encoding validation, tab indentation detection, YAML directive rejection,
// and document-end marker rejection. All checks produce accurate line numbers.
package yaml

import (
	"bytes"

	"github.com/tractl/tractl/internal/validation/common"
)

// checkEncoding delegates to the shared encoding checker and normalizes the
// error codes to YAML_ENCODING so callers can match on ErrEncoding.
func checkEncoding(input []byte) ([]byte, []ValidationError) {
	stripped, errs := common.CheckEncoding(input)
	// Remap the generic ENCODING_INVALID code to the YAML-prefixed code so
	// that existing callers matching on ErrEncoding continue to work.
	for i := range errs {
		if errs[i].Code == common.ErrEncoding {
			errs[i].Code = ErrEncoding
		}
	}
	return stripped, errs
}

// normalizeLineEndings replaces CRLF and bare CR sequences with LF.
func normalizeLineEndings(input []byte) []byte {
	return common.NormalizeLineEndings(input)
}

// checkTabIndentation scans the input line by line and reports any tab
// character that appears in an indentation position — that is, before the
// first non-whitespace character on the line.
// Spec ref: §6.6.
func checkTabIndentation(input []byte) []ValidationError {
	var errs []ValidationError
	lineNum := 1

	for _, line := range bytes.Split(input, []byte{'\n'}) {
		for col, b := range line {
			switch b {
			case '\t':
				errs = append(errs, ValidationError{
					Code:    ErrTabIndentation,
					Message: "tab character in indentation position; traCtl YAML requires spaces only",
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

// checkDirectives scans the preamble region of the document for YAML
// directive lines beginning with %.
// Spec ref: §6.3.
func checkDirectives(input []byte) []ValidationError {
	var errs []ValidationError
	lineNum := 1

	for _, line := range bytes.Split(input, []byte{'\n'}) {
		stripped := bytes.TrimRight(line, " \t")

		if bytes.Equal(stripped, []byte("---")) {
			break
		}

		trimmed := bytes.TrimLeft(line, " \t")

		if len(trimmed) == 0 || trimmed[0] == '#' {
			lineNum++
			continue
		}

		if trimmed[0] == '%' {
			errs = append(errs, ValidationError{
				Code:    ErrDirective,
				Message: "YAML directives are not permitted in traCtl YAML documents: " + common.Truncate(string(trimmed), 48),
				Line:    lineNum,
				Column:  bytes.IndexByte(line, '%') + 1,
			})
		} else {
			break
		}

		lineNum++
	}

	return errs
}

// checkDocumentEndMarker scans for the YAML document-end marker (...).
// The document-start marker (---) is permitted; the document-end marker is not.
// Spec ref: §6.2.
func checkDocumentEndMarker(input []byte) []ValidationError {
	var errs []ValidationError
	lineNum := 1

	for _, line := range bytes.Split(input, []byte{'\n'}) {
		if bytes.Equal(bytes.TrimRight(line, " \t\r"), []byte("...")) {
			errs = append(errs, ValidationError{
				Code:    ErrDocumentEndMarker,
				Message: "the document-end marker (...) is not permitted in traCtl YAML documents",
				Line:    lineNum,
				Column:  1,
			})
		}
		lineNum++
	}

	return errs
}
