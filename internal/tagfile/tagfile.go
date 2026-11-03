// Package tagfile writes APEv2 .tag files compatible with foo_external_tags.
package tagfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
	Written bool
	Skipped bool
	Reason  string
}

// Build serialises fields and cover art into an APEv2 tag file body.
func Build(fields map[string]string, cover *extractor.Cover, cfg *config.Config) ([]byte, int, error) {
	tag := ape.NewTag()

	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		tag.AddText(key, fields[key])
	}

	if cfg.Cover.Enabled && cover != nil && len(cover.Data) > 0 {
		name := cover.Filename
		if cfg.Cover.Filename != "" {
			name = cfg.Cover.Filename + filepath.Ext(name)
		}
		tag.AddCoverArt(cfg.Cover.Field, name, cover.Data)
	}

	data, err := tag.Serialize()
	if err != nil {
		return nil, 0, err
	}
	return data, len(tag.Items), nil
}

// Write builds and stores <audio>.<suffix>, honouring the on_existing policy.
func Write(audioPath string, fields map[string]string, cover *extractor.Cover, cfg *config.Config, dryRun bool) (Result, error) {
	target := audioPath + cfg.TagFile.Suffix
	res := Result{Path: target}

	data, items, err := Build(fields, cover, cfg)
	if err != nil {
		return res, err
	}
	res.Items = items
	res.Bytes = len(data)
	if items == 0 {
		res.Skipped = true
		res.Reason = "no tags"
		return res, nil
	}

	if _, err := os.Stat(target); err == nil {
		switch cfg.Files.OnExisting {
		case config.OnExistingSkip:
			res.Skipped = true
			res.Reason = "already exists"
			return res, nil
		case config.OnExistingError:
			return res, fmt.Errorf("%s already exists", target)
		case config.OnExistingBackup:
			res.Backup = target + ".bak"
			if !dryRun {
				if err := os.Rename(target, res.Backup); err != nil {
					return res, fmt.Errorf("backup %s: %w", target, err)
				}
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
