package config

import (
	"strings"
	"testing"
)

func TestTagReference(t *testing.T) {
	ref := TagReference()
	if !strings.Contains(ref, "Canonical field names") || !strings.Contains(ref, "Title") {
		t.Fatalf("reference = %q", ref)
	}
	if !strings.Contains(ref, "Cover Art (Front)") || !strings.Contains(ref, "Ignored container fields") {
		t.Fatalf("reference artwork = %q", ref)
	}
}
