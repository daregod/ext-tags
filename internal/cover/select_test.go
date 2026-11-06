package cover

import (
	"testing"

	"github.com/daregod/ext-tags/internal/metadata"
)

func c(size int, src string) Candidate {
	return Candidate{Source: src, Cover: &metadata.Cover{Data: make([]byte, size)}}
}

func TestSelectPreferenceAndBudget(t *testing.T) {
	sidecar := c(3<<20, "sidecar")
	embedded := c(1<<20, "embedded")
	got, ok := Select([]Candidate{embedded, sidecar}, "sidecar", 2<<20, 8<<20)
	if !ok || got.Source != "embedded" {
		t.Fatalf("got %q, %v; want embedded", got.Source, ok)
	}
	got, ok = Select([]Candidate{embedded, sidecar}, "sidecar", 4<<20, 8<<20)
	if !ok || got.Source != "sidecar" {
		t.Fatalf("got %q, %v; want sidecar", got.Source, ok)
	}
}

func TestSelectFallbackToSmallestUnderHardLimit(t *testing.T) {
	sidecar := c(7<<20, "sidecar")
	embedded := c(3<<20, "embedded")
	got, ok := Select([]Candidate{sidecar, embedded}, "sidecar", 2<<20, 8<<20)
	if !ok || got.Source != "embedded" {
		t.Fatalf("fallback = %q, %v; want embedded", got.Source, ok)
	}
	got, ok = Select([]Candidate{sidecar, embedded}, "sidecar", 2<<20, 5<<20)
	if !ok || got.Source != "embedded" {
		t.Fatalf("fallback = %q, %v; want embedded", got.Source, ok)
	}
	if _, ok := Select([]Candidate{c(9<<20, "sidecar"), c(9<<20, "album")}, "sidecar", 2<<20, 8<<20); ok {
		t.Fatal("hard limit must reject all")
	}
}

func TestSelectNoLimit(t *testing.T) {
	got, ok := Select([]Candidate{c(6<<20, "sidecar"), c(1<<20, "embedded")}, "sidecar", 0, 0)
	if !ok || got.Source != "sidecar" {
		t.Fatalf("got %q, %v", got.Source, ok)
	}
}
