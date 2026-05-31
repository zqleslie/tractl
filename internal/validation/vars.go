// vars.go defines package-level regexp patterns and constant maps shared by
// all validation rule files — identifier format, capability format, reserved
// prefixes, and the exhaustive set of valid step kinds.
package validation

import "regexp"

// identifierRe is the canonical user-ID pattern. Reference: tractl_spec.md §4.1
var identifierRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]*$`)

// capPlatformRe matches a valid platform capability: dotted identifier, no @.
// Reference: tractl_spec.md §3.2
var capPlatformRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]*(\.[a-zA-Z][a-zA-Z0-9_.-]*)+$`)

// reservedPrefixes are identifier prefixes forbidden by tractl_spec.md §4.1.
var reservedPrefixes = []string{"traCtl.", "_traCtl."}

// allowedStepKinds is the exhaustive set of valid step kinds. Reference: tractl_spec.md §7.2
var allowedStepKinds = map[string]bool{
	"request":       true,
	"script":        true,
	"extensionCall": true,
	"composite":     true,
}
