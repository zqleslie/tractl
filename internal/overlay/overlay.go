// Package overlay defines the canonical overlay document model.
package overlay

// MergeAction is the typed enum for how a patch is applied to its target.
type MergeAction string

// MergeAction constants.
const (
	ActionReplace      MergeAction = "replace"
	ActionDeepMerge    MergeAction = "deepMerge"
	ActionAppend       MergeAction = "append"
	ActionAppendUnique MergeAction = "appendUnique"
	ActionRemove       MergeAction = "remove"
)

// TargetMode controls whether a selector matches one or all qualifying nodes.
type TargetMode string

// TargetMode constants.
const (
	TargetModeOne TargetMode = "one" // default when empty
	TargetModeAll TargetMode = "all"
)

// MatchSelector is an open-ended set of semantic match properties. Reference: overlay spec §9.2
type MatchSelector map[string]any

// SourceSelector identifies a source-native adapter and its adapter-defined fields.
type SourceSelector struct {
	Type   string         `json:"type"             yaml:"type"`
	Fields map[string]any `json:"fields,omitempty" yaml:"fields,omitempty"`
}

// Target describes what the patch operates on. Exactly one of Path, Match,
// or Source must be set. Mode is only meaningful for Match and Source targets.
type Target struct {
	Path   string          `json:"path,omitempty"   yaml:"path,omitempty"`
	Match  MatchSelector   `json:"match,omitempty"  yaml:"match,omitempty"`
	Source *SourceSelector `json:"source,omitempty" yaml:"source,omitempty"`
	Mode   TargetMode      `json:"mode,omitempty"   yaml:"mode,omitempty"`
}

// Patch is a single overlay operation: a target selector plus a merge action and data.
type Patch struct {
	Target Target         `json:"target"           yaml:"target"`
	Action MergeAction    `json:"action,omitempty" yaml:"action,omitempty"`
	Data   map[string]any `json:"data,omitempty"   yaml:"data,omitempty"`
}

// OverlayMetadata holds human-readable descriptors for an overlay document.
type OverlayMetadata struct { //nolint:revive
	Name        string `json:"name,omitempty"        yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// OverlayDocument is the document root for an overlay file.
type OverlayDocument struct { //nolint:revive
	Metadata *OverlayMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Patches  []Patch          `json:"patches"            yaml:"patches"`
}

// IdentityContractArrays is the normative set of array field names on which
// appendUnique is permitted. Reference: overlay spec §10.5
var IdentityContractArrays = map[string]bool{
	"assertions": true,
	"extracts":   true,
	"workflows":  true,
	"steps":      true,
	"extensions": true,
	"bindings":   true,
}
