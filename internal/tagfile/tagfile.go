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

type artworkSpec struct {
	kind     string
	key      string
	filename string
}

// Build serialises fields and artwork into an APEv2 tag file body, merging
// text fields from an existing tag when configured.
func Build(fields map[string]string, covers map[string]*metadata.Cover, existing *ape.ParsedTag, cfg *config.Config) ([]byte, int, map[string]string, error) {
	merged := mergeFields(fields, existing, cfg)
	return build(merged, covers, existing, cfg)
}

// BuildFinal serialises already merged fields. Use this when the caller has
// merged existing .tag fields before applying rules; it never re-adds fields
// from the existing tag.
func BuildFinal(fields map[string]string, covers map[string]*metadata.Cover, existing *ape.ParsedTag, cfg *config.Config) ([]byte, int, map[string]string, error) {
	return build(fields, covers, existing, cfg)
}

func mergeFields(fields map[string]string, existing *ape.ParsedTag, cfg *config.Config) map[string]string {
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
	if cfg.Files.MergeExisting && existing != nil {
		for _, item := range existing.Items {
			if item.Flags&2 != 0 {
				continue
			}
			set(item.Key, string(item.Value))
		}
	}
	for key, value := range fields {
		set(key, value)
	}
	return merged
}

func build(merged map[string]string, covers map[string]*metadata.Cover, existing *ape.ParsedTag, cfg *config.Config) ([]byte, int, map[string]string, error) {
	existingArt := make(map[string][]byte)
	if cfg.Files.MergeExisting && existing != nil {
		for _, item := range existing.Items {
			if item.Flags&2 != 0 {
				existingArt[strings.ToLower(item.Key)] = item.Value
			}
		}
	}

	tag := ape.NewTag()
	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := merged[key]
		if strings.EqualFold(key, "Www") || strings.ContainsAny(value, "\r\n") {
			value = normalizeMultiline(value)
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
			if spec.key == "" || !nameAllowed(cfg, spec.key) {
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

// nameAllowed applies fields.include/exclude to text fields and artwork keys.
func nameAllowed(cfg *config.Config, name string) bool {
	if len(cfg.Fields.Include) > 0 {
		found := false
		for _, candidate := range cfg.Fields.Include {
			if strings.EqualFold(candidate, name) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for _, candidate := range cfg.Fields.Exclude {
		if strings.EqualFold(candidate, name) {
			return false
		}
	}
	return true
}

// ReadExisting returns the existing .tag for an audio file, or nil if it does not exist.
func ReadExisting(audioPath, suffix string) (*ape.ParsedTag, error) {
	path := audioPath + suffix
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return ape.ReadFile(path)
}

// ExistingFields returns text fields from an existing tag.
func ExistingFields(existing *ape.ParsedTag) map[string]string {
	out := make(map[string]string)
	if existing == nil {
		return out
	}
	for _, item := range existing.Items {
		if item.Flags&2 == 0 {
			out[item.Key] = string(item.Value)
		}
	}
	return out
}

// Write builds and stores <audio>.<suffix>, honouring the on_existing policy.
func Write(audioPath string, fields map[string]string, covers map[string]*metadata.Cover, cfg *config.Config, dryRun bool) (Result, error) {
	target := audioPath + cfg.TagFile.Suffix
	var existing *ape.ParsedTag
	if cfg.Files.MergeExisting {
		parsed, err := ReadExisting(audioPath, cfg.TagFile.Suffix)
		if err != nil {
			return Result{Path: target}, fmt.Errorf("merge %s: %w", target, err)
		}
		existing = parsed
	}
	return write(audioPath, target, fields, covers, existing, cfg, dryRun, true)
}

// WriteTo writes to an explicit target. existing is the source .tag (next
// to audioPath), not the target .tag. Fields must already contain merged
// existing .tag values, typically from rules.BuildWithExisting.
func WriteTo(audioPath, target string, fields map[string]string, covers map[string]*metadata.Cover, existing *ape.ParsedTag, cfg *config.Config, dryRun bool) (Result, error) {
	return write(audioPath, target, fields, covers, existing, cfg, dryRun, false)
}

func write(audioPath, target string, fields map[string]string, covers map[string]*metadata.Cover, existing *ape.ParsedTag, cfg *config.Config, dryRun, merge bool) (Result, error) {
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

	var data []byte
	var items int
	var mergedFields map[string]string
	var err error
	if merge {
		data, items, mergedFields, err = Build(fields, covers, existing, cfg)
	} else {
		data, items, mergedFields, err = BuildFinal(fields, covers, existing, cfg)
	}
	if err != nil {
		return res, err
	}
	res.Items = items
	res.Bytes = len(data)
	res.Fields = mergedFields
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

func normalizeMultiline(value string) string {
	lines := strings.FieldsFunc(value, func(r rune) bool { return r == '\r' || r == '\n' })
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\r\n")
}
