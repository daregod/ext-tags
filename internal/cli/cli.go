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

	"github.com/daregod/ext-tags/internal/album"
	"github.com/daregod/ext-tags/internal/artwork"
	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
	"github.com/daregod/ext-tags/internal/registry"
	"github.com/daregod/ext-tags/internal/rules"
	"github.com/daregod/ext-tags/internal/tagfile"
	"github.com/daregod/ext-tags/internal/tagformat"
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

	output      string
	stdout      bool
	ignoreLimit bool
}

type app struct {
	opts     *options
	registry *tagformat.Registry
	failed   bool
}

// Execute runs the CLI and returns the process exit code.
func Execute() int {
	a := &app{opts: &options{}, registry: registry.NewDefault()}
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

	root.AddCommand(a.extractCmd(), a.importCmd(), a.fixextCmd(), a.albumCmd(), a.artworkCmd(), a.applyCmd(), a.selfUpdateCmd(), a.renderCmd(), a.configCmd(), a.versionCmd())
	return root
}

func (a *app) extractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract <path...>",
		Short: "Write <file>.<ext>.tag files from audio metadata",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runExtract(cmd, args)
		},
	}
	cmd.Flags().StringVar(&a.opts.output, "output", "", "write the .tag to this path instead of next to the audio file (- = stdout)")
	cmd.Flags().BoolVar(&a.opts.stdout, "stdout", false, "write the .tag bytes to stdout")
	cmd.Flags().BoolVar(&a.opts.ignoreLimit, "ignore-limit", false, "include artwork above cover.hard_max_bytes")
	return cmd
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
	if err := a.validateExtractOutput(); err != nil {
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
	if (a.opts.stdout || a.opts.output != "") && len(files) != 1 {
		return fmt.Errorf("--output/--stdout require exactly one input file, got %d", len(files))
	}
	for _, file := range files {
		if a.isOwnOutput(file, cfg) {
			continue
		}
		a.processFile(file, cfg, engine)
	}
	return nil
}

func (a *app) validateExtractOutput() error {
	if a.opts.output == "-" {
		a.opts.stdout = true
		a.opts.output = ""
	}
	if a.opts.output != "" && a.opts.stdout {
		return fmt.Errorf("--output and --stdout are mutually exclusive")
	}
	if a.opts.stdout && a.opts.json {
		return fmt.Errorf("--stdout cannot be combined with --json: stdout is reserved for tag bytes")
	}
	if a.opts.stdout && a.opts.dryRun {
		return fmt.Errorf("--stdout cannot be combined with --dry-run")
	}
	return nil
}

// isOwnOutput reports whether path is a .tag (or its .bak) we produce ourselves.
// applySidecarArtwork embeds images found next to the audio file for every
// artwork slot that has no embedded picture yet.
func (a *app) applySidecarArtwork(md *metadata.Document, path string, cfg *config.Config) {
	if !cfg.Cover.Enabled || !cfg.Cover.Directory {
		return
	}
	if md.Artwork == nil {
		md.Artwork = map[string]*metadata.Cover{}
	}
	if md.Cover != nil {
		md.Artwork["front"] = md.Cover
	}
	specs := []struct {
		kind  string
		names []string
	}{
		{"front", cfg.Cover.Names},
		{"back", cfg.Cover.Back},
		{"artist", cfg.Cover.Artist},
		{"disc", cfg.Cover.Disc},
		{"icon", cfg.Cover.Icon},
	}
	for _, spec := range specs {
		if len(spec.names) == 0 {
			continue
		}
		if cfg.Cover.Prefer != "sidecar" && md.Artwork[spec.kind] != nil {
			continue
		}
		img, err := artwork.Find(filepath.Dir(path), spec.names)
		if err != nil {
			continue
		}
		name := spec.kind + strings.ToLower(filepath.Ext(img.Filename))
		if spec.kind == "front" {
			name = "cover" + strings.ToLower(filepath.Ext(img.Filename))
		}
		md.Artwork[spec.kind] = &metadata.Cover{Filename: name, MIME: img.MIME, Data: img.Data}
	}
	if front := md.Artwork["front"]; front != nil {
		md.Cover = front
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
	Target  string            `json:"target,omitempty"`
	Reason  string            `json:"reason,omitempty"`
	Warning string            `json:"warning,omitempty"`
	Error   string            `json:"error,omitempty"`
}

func (a *app) processFile(path string, cfg *config.Config, engine *rules.Engine) {
	res := fileResult{File: path}

	if !read.IsSupported(path, cfg.Files.Extensions) {
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

	reader, ok := a.getRegistry().ReaderFor(path)
	if !ok {
		res.Status = "unsupported"
		res.Reason = fmt.Sprintf("unsupported format %s", strings.ToLower(filepath.Ext(path)))
		a.report(res)
		return
	}
	md, err := reader.Read(path, tagformat.ReadOptions{CoverNames: cfg.Cover.Names})
	if errors.Is(err, read.ErrNoTags) {
		md = metadata.New()
	} else if err != nil {
		res.Status = "error"
		res.Error = err.Error()
		a.failed = true
		a.report(res)
		return
	}

	albumCfg, err := album.Load(filepath.Dir(path))
	if err != nil {
		res.Status = "error"
		res.Error = err.Error()
		a.failed = true
		a.report(res)
		return
	}
	var albumOverrides map[string]string
	if albumCfg != nil {
		albumOverrides = albumCfg.Apply(md, path)
	}
	embedded := documentArtwork(md)
	var existing *ape.ParsedTag
	if cfg.Files.MergeExisting {
		existing, err = tagfile.ReadExisting(path, cfg.TagFile.Suffix)
		if err != nil {
			res.Status = "error"
			res.Error = err.Error()
			a.failed = true
			a.report(res)
			return
		}
	}
	var albumCovers map[string]*metadata.Cover
	if albumCfg != nil {
		albumCovers, err = albumCfg.ArtworkCovers()
		if err != nil {
			res.Status = "error"
			res.Error = err.Error()
			a.failed = true
			a.report(res)
			return
		}
	}
	artwork, artworkSkips := selectArtworkWithOptions(embedded, a.sidecarArtwork(path, cfg), albumCovers, tagfile.ExistingArtwork(existing), cfg, a.opts.ignoreLimit)
	md.Artwork = artwork
	if !a.opts.ignoreLimit {
		warnArtworkSkips(path, artworkSkips, "--ignore-limit")
	}
	md.Cover = md.Artwork["front"]
	existingFields := tagfile.ExistingFields(existing)
	fields, err := engine.BuildWithOverridesPrefer(md, existingFields, albumOverrides, meta, preferExisting(cfg, path))
	if err != nil {
		res.Status = "error"
		res.Error = err.Error()
		a.failed = true
		a.report(res)
		return
	}

	if a.opts.stdout {
		data, items, _, err := tagfile.BuildFinal(fields, md.Artwork, nil, cfg)
		if err != nil {
			res.Status = "error"
			res.Error = err.Error()
			a.failed = true
			a.report(res)
			return
		}
		if items == 0 {
			if !a.opts.quiet {
				fmt.Fprintf(os.Stderr, "SKIP    %s (no tags)\n", path)
			}
			return
		}
		if _, err := os.Stdout.Write(data); err != nil {
			res.Status = "error"
			res.Error = err.Error()
			a.failed = true
			a.report(res)
		}
		return
	}

	target := path + cfg.TagFile.Suffix
	if a.opts.output != "" {
		target = a.opts.output
	}
	written, err := tagfile.WriteTo(path, target, fields, md.Artwork, nil, cfg, a.opts.dryRun)
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
	case "rename-dry":
		fmt.Printf("DRY-RUN %s → %s\n", r.File, r.Target)
	case "renamed":
		fmt.Printf("RENAME  %s → %s\n", r.File, r.Target)
	case "imported":
		fmt.Printf("IMPORT  %s ← %s (%d fields)\n", r.File, r.TagFile, r.Items)
		if r.Backup != "" {
			fmt.Printf("        backup: %s\n", r.Backup)
		}
		if r.Warning != "" {
			fmt.Printf("        warning: %s\n", r.Warning)
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
				if !read.IsSupported(path, extensions) {
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
func (a *app) getRegistry() *tagformat.Registry {
	if a.registry == nil {
		a.registry = registry.NewDefault()
	}
	return a.registry
}
