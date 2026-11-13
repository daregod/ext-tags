package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type fixTagOptions struct {
	apply bool
}

func (a *app) fixTagCmd() *cobra.Command {
	opts := &fixTagOptions{}
	cmd := &cobra.Command{
		Use:   "fix-tag <path...>",
		Short: "Rename external .tag files to match the audio filename",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runFixTag(args, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.apply, "apply", false, "actually rename files (default: dry run)")
	return cmd
}

func findTagSource(path, suffix string) (string, string, error) {
	dir := filepath.Dir(path)
	target := path + suffix
	base := filepath.Base(path)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", err
	}
	var exact, orphan []string
	want := strings.ToLower(filepath.Base(target))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) <= len(suffix) || !strings.EqualFold(name[len(name)-len(suffix):], suffix) {
			continue
		}
		candidate := filepath.Join(dir, name)
		if candidate == target {
			continue
		}
		if strings.ToLower(name) == want {
			exact = append(exact, candidate)
			continue
		}
		tagBase := name[:len(name)-len(suffix)]
		tagStem := strings.TrimSuffix(tagBase, filepath.Ext(tagBase))
		if !strings.EqualFold(tagStem, stem) {
			continue
		}
		implied := filepath.Join(dir, tagBase)
		if tagBase != base {
			if _, err := os.Stat(implied); err == nil {
				continue
			}
		}
		orphan = append(orphan, candidate)
	}
	switch {
	case len(exact) == 1:
		return exact[0], "", nil
	case len(exact) > 1:
		return "", "multiple .tag candidates", nil
	case len(orphan) == 1:
		return orphan[0], "", nil
	case len(orphan) > 1:
		return "", "multiple .tag candidates", nil
	}
	return "", "", nil
}
func (a *app) runFixTag(paths []string, opts *fixTagOptions) error {
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
		target := path + cfg.TagFile.Suffix
		if _, err := os.Stat(target); err == nil {
			continue
		}
		source, reason, err := findTagSource(path, cfg.TagFile.Suffix)
		if err != nil {
			res.Status = "error"
			res.Error = err.Error()
			a.failed = true
			a.report(res)
			continue
		}
		if reason != "" {
			res.Status = "skipped"
			res.Reason = reason
			a.report(res)
			continue
		}
		if source == "" {
			continue
		}
		res.Target = target
		if _, err := os.Stat(target); err == nil {
			res.Status = "skipped"
			res.Reason = "target already exists"
			a.report(res)
			continue
		}
		if !opts.apply {
			res.Status = "rename-dry"
			a.report(res)
			continue
		}
		if err := os.Rename(source, target); err != nil {
			res.Status = "error"
			res.Error = fmt.Sprintf("rename .tag: %v", err)
			a.failed = true
			a.report(res)
			continue
		}
		res.Status = "renamed"
		a.report(res)
	}
	a.printSummary()
	return nil
}
