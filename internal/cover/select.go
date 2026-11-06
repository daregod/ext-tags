// Package cover chooses between artwork sources without decoding images.
package cover

import (
	"sort"

	"github.com/daregod/ext-tags/internal/metadata"
)

// Candidate is one artwork source for a slot.
type Candidate struct {
	Source string // album, sidecar, embedded, existing
	Cover  *metadata.Cover
}

// Select chooses a candidate using the configured preference and byte limits.
// maxBytes is a preferred budget; hardBytes is an absolute ceiling (0 disables
// either limit). If no candidate fits the preferred budget, the smallest
// candidate below the hard ceiling is used.
func Select(candidates []Candidate, prefer string, maxBytes, hardBytes int64) (Candidate, bool) {
	ordered := order(candidates, prefer)
	valid := make([]Candidate, 0, len(ordered))
	for _, c := range ordered {
		if c.Cover == nil || len(c.Cover.Data) == 0 {
			continue
		}
		if hardBytes > 0 && int64(len(c.Cover.Data)) > hardBytes {
			continue
		}
		valid = append(valid, c)
	}
	if len(valid) == 0 {
		return Candidate{}, false
	}
	if maxBytes > 0 {
		for _, c := range valid {
			if int64(len(c.Cover.Data)) <= maxBytes {
				return c, true
			}
		}
		best := valid[0]
		for _, c := range valid[1:] {
			if len(c.Cover.Data) < len(best.Cover.Data) {
				best = c
			}
		}
		return best, true
	}
	return valid[0], true
}

func order(candidates []Candidate, prefer string) []Candidate {
	priority := map[string]int{
		"album":    0,
		"embedded": 2,
		"sidecar":  1,
		"existing": 4,
	}
	switch prefer {
	case "sidecar":
		priority["sidecar"] = 0
		priority["embedded"] = 1
		priority["existing"] = 2
	case "existing":
		priority["existing"] = 0
		priority["sidecar"] = 1
		priority["embedded"] = 2
	case "embedded", "":
		priority["embedded"] = 1
		priority["sidecar"] = 2
		priority["existing"] = 3
	}
	out := append([]Candidate(nil), candidates...)
	sort.SliceStable(out, func(i, j int) bool {
		return priority[out[i].Source] < priority[out[j].Source]
	})
	return out
}
