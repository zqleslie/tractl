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
	"github.com/tractl/tractl/internal/validation"
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

		fv, err := validation.ValidatorFor(format)
		if err != nil {
			return nil, fmt.Errorf("overlay: unsupported format %q in %q", format, path)
		}
		if result := fv.Validate(raw); !result.Valid {
			msgs := make([]string, len(result.Errors))
			for i, e := range result.Errors {
				msgs[i] = e.Error()
			}
			return nil, fmt.Errorf("overlay: format validation failed for %q:\n%s", path, strings.Join(msgs, "\n"))
		}

		doc, err := overlay.Parse(raw, format)
		if err != nil {
			return nil, fmt.Errorf("overlay: parse %q: %w", path, err)
		}

		if errs := overlay.NewOverlayValidator().Validate(doc); len(errs) > 0 {
			msgs := make([]string, len(errs))
			for i, e := range errs {
				msgs[i] = fmt.Sprintf("[%s] %s (field: %s)", e.Code, e.Message, e.Field)
			}
			return nil, fmt.Errorf("overlay: validation failed for %q:\n%s", path, strings.Join(msgs, "\n"))
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
