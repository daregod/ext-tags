// Package cli implements the ext-tags command line interface.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/daregod/ext-tags/internal/artwork"
	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/extractor"
	"github.com/daregod/ext-tags/internal/rules"
	"github.com/daregod/ext-tags/internal/tagfile"
	"github.com/daregod/ext-tags/internal/version"
)

type options struct {
	configPath string
	profile    string
	dryRun     bool
	json       bool
	quiet      bool
	verbose    bool

	recursive  bool
	onExisting string
	cover      bool
	merge      bool
	noMerge    bool
}

type app struct {
	opts   *options
	failed bool
}

// Execute runs the CLI and returns the process exit code.
func Execute() int {
	a := &app{opts: &options{}}
	if err := a.rootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if a.failed {
		return 1
	}
	return 0
}

func (a *app) rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "ext-tags [flags] <command> [paths...]",
		Short:         "Extract audio metadata into foobar2000 external .tag files",
		Long:          "ext-tags reads tags from audio files and writes <file>.<ext>.tag files\ncompatible with the foo_external_tags component of foobar2000.",
		Version:       version.String(),
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return a.runExtract(cmd, args)
		},
	}
	root.SetVersionTemplate("ext-tags {{.Version}}\n")

	pf := root.PersistentFlags()
	pf.StringVar(&a.opts.configPath, "config", "", "config file (default: ext-tags.yaml, then user config dir)")
	pf.StringVar(&a.opts.profile, "profile", "", "named profile from the config file")
	pf.BoolVar(&a.opts.dryRun, "dry-run", false, "report what would be written without touching files")
	pf.BoolVar(&a.opts.json, "json", false, "print one JSON object per file")
	pf.BoolVar(&a.opts.quiet, "quiet", false, "only print errors")
	pf.BoolVarP(&a.opts.verbose, "verbose", "v", false, "print the resolved config")
	pf.BoolVarP(&a.opts.recursive, "recursive", "r", false, "recurse into subdirectories (extract)")
	pf.StringVar(&a.opts.onExisting, "on-existing", "", "overwrite|skip|backup|error (default: from config)")
	pf.BoolVar(&a.opts.cover, "cover", true, "embed cover art (extract)")
	pf.BoolVar(&a.opts.merge, "merge", true, "merge fields from an existing .tag (default from config)")
	pf.BoolVar(&a.opts.noMerge, "no-merge", false, "do not merge fields from an existing .tag")

	root.AddCommand(a.extractCmd(), a.importCmd(), a.configCmd(), a.versionCmd())
	return root
}

func (a *app) extractCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "extract <path...>",
		Short: "Write <file>.<ext>.tag files from audio metadata",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runExtract(cmd, args)
		},
	}
}
func (a *app) versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the build version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ext-tags %s\n", version.String())
		},
	}
}

func (a *app) runExtract(cmd *cobra.Command, paths []string) error {
	cfg, label, err := a.effectiveConfig()
	if err != nil {
		return err
	}
	if err := a.applyOverrides(cfg, cmd.Flags().Changed); err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	for _, warning := range cfg.Warnings() {
		if a.opts.verbose {
			fmt.Fprintf(os.Stderr, "warning: %s\n", warning)
		}
	}
	engine, err := rules.New(cfg)
	if err != nil {
		return err
	}
	if a.opts.verbose && !a.opts.json {
		fmt.Fprintf(os.Stderr, "config: %s\n", label)
	}

	files, err := expandPaths(paths, cfg.Files.Recursive, cfg.Files.Extensions)
	if err != nil {
		return err
	}
	for _, file := range files {
		if a.isOwnOutput(file, cfg) {
			continue
		}
		a.processFile(file, cfg, engine)
	}
	return nil
}

// isOwnOutput reports whether path is a .tag (or its .bak) we produce ourselves.
// applySidecarCover embeds a cover image found next to the audio file when
// the file itself has no embedded picture.
func (a *app) applySidecarCover(md *extractor.Result, path string, cfg *config.Config) {
	if md.Cover != nil || !cfg.Cover.Enabled || !cfg.Cover.Directory {
		return
	}
	img, err := artwork.Find(filepath.Dir(path), cfg.Cover.Names)
	if err != nil {
		return
	}
	md.Cover = &extractor.Cover{
		Filename: "cover" + strings.ToLower(filepath.Ext(img.Filename)),
		MIME:     img.MIME,
		Data:     img.Data,
	}
}

func (a *app) isOwnOutput(path string, cfg *config.Config) bool {
	lower := strings.ToLower(path)
	suffix := strings.ToLower(cfg.TagFile.Suffix)
	return strings.HasSuffix(lower, suffix) || strings.HasSuffix(lower, suffix+".bak")
}

func (a *app) applyOverrides(cfg *config.Config, changed func(string) bool) error {
	if changed("merge") && changed("no-merge") {
		return fmt.Errorf("--merge and --no-merge are mutually exclusive")
	}
	if changed("recursive") {
		cfg.Files.Recursive = a.opts.recursive
	}
	if changed("on-existing") {
		cfg.Files.OnExisting = a.opts.onExisting
	}
	if changed("cover") {
		cfg.Cover.Enabled = a.opts.cover
	}
	if changed("merge") {
		cfg.Files.MergeExisting = a.opts.merge
	}
	if changed("no-merge") {
		cfg.Files.MergeExisting = false
	}
	return nil
}

func (a *app) effectiveConfig() (*config.Config, string, error) {
	explicit := a.opts.configPath
	if explicit == "" {
		explicit = os.Getenv(config.EnvConfig)
	}
	cfg, layers, err := config.LoadLayers(a.opts.profile, explicit)
	if err != nil {
		return nil, "", err
	}
	return cfg, layers.Label(), nil
}

type coverInfo struct {
	Filename string `json:"filename"`
	Bytes    int    `json:"bytes"`
}

type fileResult struct {
	File    string            `json:"file"`
	Status  string            `json:"status"`
	TagFile string            `json:"tagfile,omitempty"`
	Backup  string            `json:"backup,omitempty"`
	Items   int               `json:"items,omitempty"`
	Bytes   int               `json:"bytes,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
	Cover   *coverInfo        `json:"cover,omitempty"`
	Reason  string            `json:"reason,omitempty"`
	Error   string            `json:"error,omitempty"`
}

func (a *app) processFile(path string, cfg *config.Config, engine *rules.Engine) {
	res := fileResult{File: path}

	if !extractor.IsSupported(path, cfg.Files.Extensions) {
		res.Status = "unsupported"
		res.Reason = fmt.Sprintf("unsupported format %s", strings.ToLower(filepath.Ext(path)))
		a.report(res)
		return
	}

	ext := strings.ToLower(filepath.Ext(path))
	meta := rules.Meta{
		Path:     path,
		Filename: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Dir:      filepath.Dir(path),
		Ext:      ext,
		Now:      time.Now(),
	}

	md, err := extractor.ExtractWithOptions(path, extractor.Options{CoverNames: cfg.Cover.Names})
	if errors.Is(err, extractor.ErrNoTags) {
		res.Status = "no-tags"
		a.report(res)
		return
	}
	if err != nil {
		res.Status = "error"
		res.Error = err.Error()
		a.failed = true
		a.report(res)
		return
	}

	a.applySidecarCover(md, path, cfg)

	fields, err := engine.Build(md, meta)
	if err != nil {
		res.Status = "error"
		res.Error = err.Error()
		a.failed = true
		a.report(res)
		return
	}

	written, err := tagfile.Write(path, fields, md.Cover, cfg, a.opts.dryRun)
	if err != nil {
		res.Status = "error"
		res.Error = err.Error()
		a.failed = true
		a.report(res)
		return
	}

	res.TagFile = written.Path
	res.Backup = written.Backup
	res.Items = written.Items
	res.Bytes = written.Bytes
	res.Fields = written.Fields
	if res.Fields == nil {
		res.Fields = fields
	}
	if cfg.Cover.Enabled && md.Cover != nil {
		res.Cover = &coverInfo{Filename: md.Cover.Filename, Bytes: len(md.Cover.Data)}
	}

	switch {
	case written.Skipped:
		res.Status = "skipped"
		res.Reason = written.Reason
	case a.opts.dryRun:
		res.Status = "dry-run"
	default:
		res.Status = "ok"
	}
	a.report(res)
}

func (a *app) report(r fileResult) {
	if a.opts.json {
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(r); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			a.failed = true
		}
		return
	}

	if a.opts.quiet && r.Status != "error" {
		return
	}

	switch r.Status {
	case "ok":
		fmt.Printf("OK      %s → %s (%d items, %d bytes)\n", r.File, r.TagFile, r.Items, r.Bytes)
		if r.Backup != "" {
			fmt.Printf("        backup: %s\n", r.Backup)
		}
	case "dry-run":
		fmt.Printf("DRY-RUN %s → %s (%d items, %d bytes)\n", r.File, r.TagFile, r.Items, r.Bytes)
	case "imported":
		fmt.Printf("IMPORT  %s ← %s (%d fields)\n", r.File, r.TagFile, r.Items)
		if r.Backup != "" {
			fmt.Printf("        backup: %s\n", r.Backup)
		}
	case "import-dry":
		fmt.Printf("DRY-RUN %s ← %s (%d fields)\n", r.File, r.TagFile, r.Items)
	case "skipped":
		fmt.Printf("SKIP    %s (%s)\n", r.File, r.Reason)
	case "no-tags":
		fmt.Printf("SKIP    %s (no tags)\n", r.File)
	case "unsupported":
		fmt.Printf("SKIP    %s (%s)\n", r.File, r.Reason)
	default:
		fmt.Fprintf(os.Stderr, "ERROR   %s: %s\n", r.File, r.Error)
	}
}

func expandPaths(paths []string, recursive bool, extensions []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		matches, err := filepath.Glob(p)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			matches = []string{p}
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				// Let the per-file step report the real error.
				files = append(files, match)
				continue
			}
			if !info.IsDir() {
				files = append(files, match)
				continue
			}
			err = filepath.WalkDir(match, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					if path != match && !recursive {
						return fs.SkipDir
					}
					return nil
				}
				if !extractor.IsSupported(path, extensions) {
					return nil
				}
				files = append(files, path)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return files, nil
}
