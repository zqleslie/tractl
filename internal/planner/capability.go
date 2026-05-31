// Package planner resolves capability contracts and builds the execution DAG plan.
package planner

// CapabilityRegistry resolves capability contracts to concrete runtime bindings.
type CapabilityRegistry interface {
	// Resolve returns the resolved capability for a given contract string.
	// Returns ErrUnknownCapability if the contract cannot be resolved.
	Resolve(contract string) (*ResolvedCapability, error)
}

// ResolvedCapability describes the runtime binding of a capability contract.
type ResolvedCapability struct {
	// Contract is the capability contract identifier (e.g., "protocol.http").
	Contract string
	// Runtime is the executor runtime name (e.g., "http").
	Runtime string
	// Version is the runtime version (e.g., "1").
	Version string
}

// defaultRegistry implements CapabilityRegistry with MVP support for protocol.http.
type defaultRegistry struct{}

// Resolve implements CapabilityRegistry.
func (r *defaultRegistry) Resolve(contract string) (*ResolvedCapability, error) {
	switch contract {
	case "protocol.http":
		return &ResolvedCapability{
			Contract: "protocol.http",
			Runtime:  "http",
			Version:  "1",
		}, nil
	default:
		return nil, plannerErr(ErrUnknownCapability, "", "", "unknown capability contract: "+contract)
	}
}

// DefaultRegistry returns a registry with MVP protocol.http support only.
// All other contracts return ErrUnknownCapability.
func DefaultRegistry() CapabilityRegistry {
	return &defaultRegistry{}
}
