package cli

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/tagfile"
)

type fieldDiff struct {
	Field  string `json:"field"`
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Change string `json:"change"`
}

type coverDiff struct {
	Slot   string       `json:"slot"`
	From   *artworkInfo `json:"from,omitempty"`
	To     *artworkInfo `json:"to,omitempty"`
	Change string       `json:"change"`
}

type diffResult struct {
	Fields  []fieldDiff `json:"fields,omitempty"`
	Artwork []coverDiff `json:"artwork,omitempty"`
}

func buildDiff(existing, final map[string]string, existingArt, selectedArt map[string]*metadata.Cover, cfg *config.Config) *diffResult {
	out := &diffResult{}
	oldFields := lowerFields(existing)
	newFields := lowerFields(final)
	keys := make(map[string]bool, len(oldFields)+len(newFields))
	for key := range oldFields {
		keys[key] = true
	}
	for key := range newFields {
		keys[key] = true
	}
	sorted := make([]string, 0, len(keys))
	for key := range keys {
		sorted = append(sorted, key)
	}
	sort.Strings(sorted)
	for _, lower := range sorted {
		oldValue, oldOK := oldFields[lower]
		newValue, newOK := newFields[lower]
		if oldOK && newOK && oldValue == newValue {
			continue
		}
		change := "changed"
		switch {
		case !oldOK:
			change = "added"
		case !newOK:
			change = "removed"
		}
		out.Fields = append(out.Fields, fieldDiff{Field: fieldName(existing, final, lower), From: oldValue, To: newValue, Change: change})
	}
	for _, slot := range artworkKinds {
		oldCover := existingArt[slot]
		newCover := finalArtwork(slot, selectedArt[slot], oldCover, cfg)
		if oldCover == nil && newCover == nil {
			continue
		}
		if oldCover != nil && newCover != nil && bytes.Equal(oldCover.Data, newCover.Data) {
			continue
		}
		change := "changed"
		switch {
		case oldCover == nil:
			change = "added"
		case newCover == nil:
			change = "removed"
		}
		out.Artwork = append(out.Artwork, coverDiff{Slot: slot, From: makeArtworkInfo(oldCover), To: makeArtworkInfo(newCover), Change: change})
	}
	return out
}

func lowerFields(fields map[string]string) map[string]string {
	out := make(map[string]string, len(fields))
	for key, value := range fields {
		out[strings.ToLower(strings.TrimSpace(key))] = value
	}
	return out
}

func fieldName(old, current map[string]string, lower string) string {
	for key := range current {
		if strings.ToLower(strings.TrimSpace(key)) == lower {
			return key
		}
	}
	for key := range old {
		if strings.ToLower(strings.TrimSpace(key)) == lower {
			return key
		}
	}
	return lower
}

func finalArtwork(slot string, selected, existing *metadata.Cover, cfg *config.Config) *metadata.Cover {
	if !cfg.Cover.Enabled {
		return nil
	}
	if !tagfile.NameAllowed(cfg, artworkField(slot, cfg)) {
		return nil
	}
	if cfg.Cover.Prefer == "existing" && existing != nil {
		return existing
	}
	if selected != nil {
		return selected
	}
	return existing
}

func artworkField(slot string, cfg *config.Config) string {
	switch slot {
	case "front":
		return cfg.Cover.Field
	case "back":
		return "Cover Art (Back)"
	case "artist":
		return "Artist Picture"
	case "disc":
		return "Cover Art (Disc)"
	case "icon":
		return "Cover Art (Icon)"
	}
	return slot
}

func printFieldDiff(diff *diffResult, path, label string) {
	fmt.Printf("%s    %s\n", label, path)
	if diff == nil {
		return
	}
	for _, change := range diff.Fields {
		switch change.Change {
		case "added":
			fmt.Printf("        + %s: %q\n", change.Field, change.To)
		case "removed":
			fmt.Printf("        - %s: %q\n", change.Field, change.From)
		default:
			fmt.Printf("        ~ %s: %q -> %q\n", change.Field, change.From, change.To)
		}
	}
	for _, change := range diff.Artwork {
		fmt.Printf("        cover %s: %s -> %s\n", change.Slot, formatCover(change.From), formatCover(change.To))
	}
}

func formatCover(info *artworkInfo) string {
	if info == nil {
		return "none"
	}
	parts := []string{info.Filename}
	if info.MIME != "" {
		parts = append(parts, info.MIME)
	}
	if info.Width > 0 && info.Height > 0 {
		parts = append(parts, fmt.Sprintf("%dx%d", info.Width, info.Height))
	}
	parts = append(parts, fmt.Sprintf("%d bytes", info.Bytes))
	return strings.Join(parts, ", ")
}
