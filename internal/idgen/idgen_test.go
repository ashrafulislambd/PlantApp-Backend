package idgen

import (
	"strings"
	"testing"
)

func TestNew_PrefixesAndUniqueness(t *testing.T) {
	gen := New()

	a := gen.New("pl")
	b := gen.New("pl")

	if !strings.HasPrefix(a, "pl_") {
		t.Errorf("expected id to start with %q, got %q", "pl_", a)
	}
	if a == b {
		t.Errorf("expected two generated ids to differ, both were %q", a)
	}

	const wantLen = len("pl_") + 24 // 12 random bytes -> 24 hex chars
	if len(a) != wantLen {
		t.Errorf("expected id length %d, got %d (%q)", wantLen, len(a), a)
	}
}

func TestNew_DifferentPrefix(t *testing.T) {
	gen := New()
	id := gen.New("fert")
	if !strings.HasPrefix(id, "fert_") {
		t.Errorf("expected id to start with %q, got %q", "fert_", id)
	}
}
