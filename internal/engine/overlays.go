// overlays.go applies overlay documents to a parsed spec, merging field values
// according to the overlay engine's path and strategy rules.

package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tractl/tractl/internal/overlay"
	"github.com/tractl/tractl/internal/spec"
)

// applyOverlays applies each overlay file in order to the spec.
func applyOverlays(s *spec.TraCtlSpec, overlayFiles []string) (*spec.TraCtlSpec, error) {
	eng := overlay.NewEngine()

	for _, path := range overlayFiles {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("overlay: cannot read %q: %w", path, err)
		}

		format := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
		if format == "" {
			format = "yaml"
		}

		doc, err := overlay.Parse(raw, format)
		if err != nil {
			return nil, fmt.Errorf("overlay: parse %q: %w", path, err)
		}

		specJSON, err := json.Marshal(s)
		if err != nil {
			return nil, fmt.Errorf("overlay: marshal spec: %w", err)
		}

		var srcMap map[string]any
		if err := json.Unmarshal(specJSON, &srcMap); err != nil {
			return nil, fmt.Errorf("overlay: unmarshal spec to map: %w", err)
		}

		result, err := eng.Apply(srcMap, []*overlay.OverlayDocument{doc})
		if err != nil {
			return nil, fmt.Errorf("overlay: apply %q: %w", path, err)
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("overlay: marshal result: %w", err)
		}

		var updated spec.TraCtlSpec
		if err := json.Unmarshal(resultJSON, &updated); err != nil {
			return nil, fmt.Errorf("overlay: unmarshal result to spec: %w", err)
		}
		s = &updated
	}

	return s, nil
}
