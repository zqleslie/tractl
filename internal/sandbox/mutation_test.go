package sandbox_test

import (
	"testing"

	"github.com/tractl/tractl/internal/sandbox"
)

func TestMutationSet_Empty_AllZero(t *testing.T) {
	m := sandbox.MutationSet{}
	if !m.Empty() {
		t.Fatalf("expected empty mutation set")
	}
}

func TestMutationSet_NotEmpty_HasVariables(t *testing.T) {
	m := sandbox.MutationSet{Variables: map[string]string{"x": "1"}}
	if m.Empty() {
		t.Fatalf("expected not empty")
	}
}

func TestMutationSet_NotEmpty_CancelTrue(t *testing.T) {
	m := sandbox.MutationSet{Cancel: true}
	if m.Empty() {
		t.Fatalf("expected not empty")
	}
}

func TestMutationSet_NotEmpty_HasExtracts(t *testing.T) {
	m := sandbox.MutationSet{Extracts: map[string]string{"token": "abc"}}
	if m.Empty() {
		t.Fatalf("expected not empty when Extracts is set")
	}
}
