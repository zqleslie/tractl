package provider

// TODO: provider host, lifecycle management
// Reference: HLD §4.2, ADR-004 §6, ADR-007
// Responsibilities:
//   - provider discovery and registration
//   - provider loading and hot lifecycle management
//   - isolation boundaries
//   - capability exposure: register each provider's capability contracts
//     into the canonical capability contract model
//   - capability conflict detection at registration time (before planning)
//   - delegated execution governance
// Constraint: provider lifecycle management is SEPARATE from execution planning
