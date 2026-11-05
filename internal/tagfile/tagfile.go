// Package tagfile writes APEv2 .tag files compatible with foo_external_tags.
package tagfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/metadata"
)

// Result describes what happened to a single output file.
type Result struct {
	Path    string
	Backup  string
	Items   int
	Bytes   int
	Fields  map[string]string
	Written bool
	Skipped bool
	Reason  string
}

// Build serialises fields and cover art into an APEv2 tag file body.
// When existing is not nil and merge_existing is enabled, fields that are
// present only in the existing tag (for example Rating from the foobar2000
// media library) are preserved.
// artworkSpec describes one artwork slot stored in the .tag file.
type artworkSpec struct {
	kind     string
	key      string
	filename string
}

// Build serialises fields and artwork into an APEv2 tag file body.
// When existing is not nil and merge_existing is enabled, fields and artwork
// that are present only in the existing tag are preserved.
func Build(fields map[string]string, covers map[string]*metadata.Cover, existing *ape.ParsedTag, cfg *config.Config) ([]byte, int, map[string]string, error) {
	merged := make(map[string]string, len(fields))
	index := make(map[string]string, len(fields))
	set := func(key, value string) {
		lower := strings.ToLower(key)
		if canonical, ok := index[lower]; ok {
			merged[canonical] = value
			return
		}
		index[lower] = key
		merged[key] = value
	}

	existingArt := make(map[string][]byte)
	if cfg.Files.MergeExisting && existing != nil {
		for _, item := range existing.Items {
			if item.Flags&2 != 0 {
				existingArt[strings.ToLower(item.Key)] = item.Value
				continue
			}
			set(item.Key, string(item.Value))
		}
	}
	for key, value := range fields {
		set(key, value)
	}

	tag := ape.NewTag()

	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := merged[key]
		if strings.EqualFold(key, "Www") {
			value = normalizeWww(value)
		}
		tag.AddText(key, value)
	}

	if cfg.Cover.Enabled {
		specs := []artworkSpec{
			{"front", cfg.Cover.Field, cfg.Cover.Filename},
			{"back", "Cover Art (Back)", "back"},
			{"artist", "Artist Picture", "artist"},
			{"disc", "Cover Art (Disc)", "disc"},
			{"icon", "Cover Art (Icon)", "icon"},
		}
		for _, spec := range specs {
			if spec.key == "" {
				continue
			}
			if cfg.Cover.Prefer == "existing" {
				if raw, ok := existingArt[strings.ToLower(spec.key)]; ok && len(raw) > 0 {
					tag.AddBinary(spec.key, raw)
					continue
				}
			}
			if cover := covers[spec.kind]; cover != nil && len(cover.Data) > 0 {
				name := cover.Filename
				if spec.filename != "" {
					name = spec.filename + filepath.Ext(name)
				}
				tag.AddCoverArt(spec.key, name, cover.Data)
				continue
			}
			if raw, ok := existingArt[strings.ToLower(spec.key)]; ok && len(raw) > 0 {
				tag.AddBinary(spec.key, raw)
			}
		}
	}

	data, err := tag.Serialize()
	if err != nil {
		return nil, 0, nil, err
	}
	return data, len(tag.Items), merged, nil
}

// Write builds and stores <audio>.<suffix>, honouring the on_existing policy
// and merging fields from an existing tag file when configured.
func Write(audioPath string, fields map[string]string, covers map[string]*metadata.Cover, cfg *config.Config, dryRun bool) (Result, error) {
	target := audioPath + cfg.TagFile.Suffix
	res := Result{Path: target}

	_, statErr := os.Stat(target)
	exists := statErr == nil

	if exists {
		switch cfg.Files.OnExisting {
		case config.OnExistingSkip:
			res.Skipped = true
			res.Reason = "already exists"
			return res, nil
		case config.OnExistingError:
			return res, fmt.Errorf("%s already exists", target)
		}
	}

	var existing *ape.ParsedTag
	if exists && cfg.Files.MergeExisting {
		parsed, err := ape.ReadFile(target)
		if err != nil {
			return res, fmt.Errorf("merge %s: %w", target, err)
		}
		existing = parsed
	}

	data, items, merged, err := Build(fields, covers, existing, cfg)
	if err != nil {
		return res, err
	}
	res.Items = items
	res.Bytes = len(data)
	res.Fields = merged
	if items == 0 {
		res.Skipped = true
		res.Reason = "no tags"
		return res, nil
	}

	if exists && cfg.Files.OnExisting == config.OnExistingBackup {
		res.Backup = target + ".bak"
		if !dryRun {
			if err := os.Rename(target, res.Backup); err != nil {
				return res, fmt.Errorf("backup %s: %w", target, err)
			}
		}
	}

	if dryRun {
		res.Reason = "dry-run"
		return res, nil
	}

	if err := os.WriteFile(target, data, 0644); err != nil {
		return res, fmt.Errorf("write %s: %w", target, err)
	}
	res.Written = true
	return res, nil
}
func normalizeWww(value string) string {
	lines := strings.FieldsFunc(value, func(r rune) bool { return r == '\r' || r == '\n' })
	return strings.Join(lines, "\r\n")
}
