package common

import (
	"bytes"
	"unicode/utf8"
)

// ErrEncoding is produced when the input uses a prohibited encoding
// (UTF-16, UTF-32) or contains invalid UTF-8 byte sequences.
const ErrEncoding ErrorCode = "ENCODING_INVALID"

// CheckEncoding validates byte-level encoding compliance and strips a leading
// UTF-8 BOM when present.
//
// Returns the (possibly BOM-stripped) input and any encoding violations.
// If violations are returned, further analysis of the input is unreliable and
// the caller should not proceed to parse.
//
// Spec ref: tractl_yaml_spec.md §6.4, tractl_json_spec.md §6.4, tractl_toon_spec.md §6.1.
func CheckEncoding(input []byte) ([]byte, []ValidationError) {
	// Detect UTF-32 before UTF-16: UTF-32 LE shares its first two bytes with UTF-16 LE.
	if len(input) >= 4 {
		utf32be := input[0] == 0x00 && input[1] == 0x00 && input[2] == 0xFE && input[3] == 0xFF
		utf32le := input[0] == 0xFF && input[1] == 0xFE && input[2] == 0x00 && input[3] == 0x00
		if utf32be || utf32le {
			return input, []ValidationError{{
				Code:    ErrEncoding,
				Message: "document encoding must be UTF-8; UTF-32 encoding detected (BOM present)",
			}}
		}
	}

	if len(input) >= 2 {
		utf16be := input[0] == 0xFE && input[1] == 0xFF
		utf16le := input[0] == 0xFF && input[1] == 0xFE
		if utf16be || utf16le {
			return input, []ValidationError{{
				Code:    ErrEncoding,
				Message: "document encoding must be UTF-8; UTF-16 encoding detected (BOM present)",
			}}
		}
	}

	// Strip a leading UTF-8 BOM (EF BB BF) silently.
	// Editors such as Notepad on Windows emit this; we tolerate it on input
	// but never emit it on output.
	if bytes.HasPrefix(input, []byte{0xEF, 0xBB, 0xBF}) {
		input = input[3:]
	}

	if !utf8.Valid(input) {
		return input, []ValidationError{{
			Code:    ErrEncoding,
			Message: "document contains invalid UTF-8 byte sequences",
		}}
	}

	return input, nil
}

// NormalizeLineEndings replaces CRLF and bare CR sequences with LF.
// This must be called before any line-based scanning or parsing.
func NormalizeLineEndings(input []byte) []byte {
	out := bytes.ReplaceAll(input, []byte{'\r', '\n'}, []byte{'\n'})
	return bytes.ReplaceAll(out, []byte{'\r'}, []byte{'\n'})
}
