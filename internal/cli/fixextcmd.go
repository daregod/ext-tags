package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/daregod/ext-tags/internal/formats/ogg"
)

type fixOptions struct {
	apply bool
}

func (a *app) fixextCmd() *cobra.Command {
	opts := &fixOptions{}
	cmd := &cobra.Command{
		Use:   "fixext <path...>",
		Short: "Rename Ogg files mislabeled as .flac to their real extension",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runFixExt(args, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.apply, "apply", false, "actually rename files (default: dry run)")
	return cmd
}

func (a *app) runFixExt(paths []string, opts *fixOptions) error {
	cfg, _, err := a.effectiveConfig()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	recursive := a.opts.recursive || cfg.Files.Recursive
	files, err := expandPaths(paths, recursive, cfg.Files.Extensions)
	if err != nil {
		return err
	}
	for _, path := range files {
		res := fileResult{File: path}
		if !strings.EqualFold(filepath.Ext(path), ".flac") {
			continue
		}
		targetExt := ogg.Extension(path)
		if targetExt == "" {
			res.Status = "skipped"
			res.Reason = "not an Ogg container"
			a.report(res)
			continue
		}
		target := strings.TrimSuffix(path, filepath.Ext(path)) + targetExt
		res.Target = target
		if targetExt == ".flac" {
			res.Status = "skipped"
			res.Reason = "already has the correct extension"
			a.report(res)
			continue
		}
		if _, err := os.Stat(target); err == nil {
			res.Status = "skipped"
			res.Reason = "target already exists"
			a.report(res)
			continue
		}
		oldTag := path + cfg.TagFile.Suffix
		newTag := target + cfg.TagFile.Suffix
		if _, err := os.Stat(newTag); err == nil {
			res.Status = "skipped"
			res.Reason = "target .tag already exists"
			a.report(res)
			continue
		}
		if !opts.apply {
			res.Status = "rename-dry"
			a.report(res)
			continue
		}
		tagMoved := false
		if _, err := os.Stat(oldTag); err == nil {
			if err := os.Rename(oldTag, newTag); err != nil {
				res.Status = "error"
				res.Error = fmt.Sprintf("rename tag: %v", err)
				a.failed = true
				a.report(res)
				continue
			}
			tagMoved = true
		}
		if err := os.Rename(path, target); err != nil {
			if tagMoved {
				_ = os.Rename(newTag, oldTag)
			}
			res.Status = "error"
			res.Error = fmt.Sprintf("rename audio: %v", err)
			a.failed = true
			a.report(res)
			continue
		}
		res.Status = "renamed"
		a.report(res)
	}
	return nil
}
