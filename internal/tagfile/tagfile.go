// Package tagfile writes APEv2 .tag files compatible with foo_external_tags.
package tagfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"external-tags/internal/ape"
	"external-tags/internal/config"
	"external-tags/internal/extractor"
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
func Build(fields map[string]string, cover *extractor.Cover, existing *ape.ParsedTag, cfg *config.Config) ([]byte, int, map[string]string, error) {
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

	var existingCover []byte
	if cfg.Files.MergeExisting && existing != nil {
		for _, item := range existing.Items {
			if item.Flags&2 != 0 {
				if existingCover == nil && strings.EqualFold(item.Key, cfg.Cover.Field) {
					existingCover = item.Value
				}
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
		tag.AddText(key, merged[key])
	}

	if cfg.Cover.Enabled {
		switch {
		case cover != nil && len(cover.Data) > 0:
			name := cover.Filename
			if cfg.Cover.Filename != "" {
				name = cfg.Cover.Filename + filepath.Ext(name)
			}
			tag.AddCoverArt(cfg.Cover.Field, name, cover.Data)
		case len(existingCover) > 0:
			tag.AddBinary(cfg.Cover.Field, existingCover)
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
func Write(audioPath string, fields map[string]string, cover *extractor.Cover, cfg *config.Config, dryRun bool) (Result, error) {
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

	data, items, merged, err := Build(fields, cover, existing, cfg)
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
