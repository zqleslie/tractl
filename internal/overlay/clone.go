package overlay

import (
	"encoding/json"
	"fmt"
)

// cloneSpec deep-clones a value using a JSON marshal/unmarshal round-trip.
// The destination must be a pointer to the same type as src.
// Returns ErrCloneFailed if either step fails.
// Reference: Phase 3 implementation contract; overlay spec §3 (source must not be mutated).
func cloneSpec(src, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return &EngineError{
			Code:   ErrCloneFailed,
			Detail: fmt.Sprintf("marshal failed: %v", err),
		}
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return &EngineError{
			Code:   ErrCloneFailed,
			Detail: fmt.Sprintf("unmarshal failed: %v", err),
		}
	}
	return nil
}
