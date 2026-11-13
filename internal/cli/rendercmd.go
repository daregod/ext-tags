package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/daregod/ext-tags/internal/album"
	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
	"github.com/daregod/ext-tags/internal/rules"
	"github.com/daregod/ext-tags/internal/tagfile"
	"github.com/daregod/ext-tags/internal/tagformat"
)

type renderOptions struct {
	template string
	null     bool
	raw      bool
}

type renderResult struct {
	File  string `json:"file"`
	Value string `json:"value"`
}

func (a *app) renderCmd() *cobra.Command {
	opts := &renderOptions{}
	cmd := &cobra.Command{
		Use:   "render <path...>",
		Short: "Render a template from the final tag fields",
		Long:  "Render a Go template from the same fields and album.yaml overrides used by extract. Nothing is written to disk.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runRender(args, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.template, "template", "t", "", "Go template, for example '{{ .artist }} - {{ .title }}'")
	cmd.Flags().BoolVar(&opts.null, "null", false, "terminate records with NUL instead of newline")
	cmd.Flags().BoolVar(&opts.raw, "raw", false, "use raw container/.tag fields without album.yaml or rules")
	return cmd
}

func (a *app) runRender(paths []string, opts *renderOptions) error {
	if strings.TrimSpace(opts.template) == "" {
		return fmt.Errorf("--template is required")
	}
	cfg, label, err := a.effectiveConfig()
	if err != nil {
		return err
	}
	if a.opts.noMerge {
		cfg.Files.MergeExisting = false
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	engine, err := rules.New(cfg)
	if err != nil {
		return err
	}
	if a.opts.verbose && !a.opts.json {
		fmt.Fprintf(os.Stderr, "config: %s\n", label)
	}
	if opts.raw {
		if _, err := engine.RenderRaw(opts.template, nil, nil, nil, rules.Meta{}); err != nil {
			return fmt.Errorf("template: %w", err)
		}
	} else if _, err := engine.Render(opts.template, nil, nil, rules.Meta{}); err != nil {
		return fmt.Errorf("template: %w", err)
	}
	files, err := expandPaths(paths, cfg.Files.Recursive, cfg.Files.Extensions)
	if err != nil {
		return err
	}
	for _, file := range files {
		if a.isOwnOutput(file, cfg) {
			continue
		}
		a.stats.processed++
		a.renderFile(file, cfg, engine, opts)
	}
	a.printSummary()
	return nil
}

func (a *app) renderFile(path string, cfg *config.Config, engine *rules.Engine, opts *renderOptions) {
	reader, ok := a.getRegistry().ReaderFor(path)
	if !ok {
		a.renderError(path, fmt.Errorf("unsupported format %s", strings.ToLower(filepath.Ext(path))))
		return
	}
	doc, err := reader.Read(path, tagformat.ReadOptions{})
	if errors.Is(err, read.ErrNoTags) {
		doc = metadata.New()
	} else if err != nil {
		a.renderError(path, err)
		return
	}
	var existing *ape.ParsedTag
	if cfg.Files.MergeExisting {
		existing, err = tagfile.ReadExisting(path, cfg.TagFile.Suffix)
		if err != nil {
			a.renderError(path, err)
			return
		}
	}
	meta := rules.Meta{
		Path: path, Filename: baseName(path), Dir: filepath.Dir(path),
		Ext: filepath.Ext(path), Now: time.Now(),
	}
	if opts.raw {
		artwork, err := a.loadArtwork(path, cfg)
		if err != nil {
			a.renderError(path, err)
			return
		}
		extra := artworkTemplateData(artwork)
		for key, value := range artworkTemplateDataFor("container", documentArtwork(doc)) {
			extra[key] = value
		}
		for key, value := range artworkTemplateDataFor("tag", tagfile.ExistingArtwork(existing)) {
			extra[key] = value
		}
		value, err := engine.RenderRawWithExtra(opts.template, doc.Fields, tagfile.ExistingFields(existing), doc.Vars, extra, meta)
		if err != nil {
			a.renderError(path, err)
			return
		}
		a.writeRenderValue(path, value, opts)
		return
	}
	var overrides map[string]string
	if albumCfg, err := album.Load(filepath.Dir(path)); err != nil {
		a.renderError(path, err)
		return
	} else if albumCfg != nil {
		overrides = albumCfg.Apply(doc, path)
	}
	fields, err := engine.BuildWithOverridesPrefer(doc, tagfile.ExistingFields(existing), overrides, meta, preferExisting(cfg, path))
	if err != nil {
		a.renderError(path, err)
		return
	}
	artwork, err := a.loadArtwork(path, cfg)
	if err != nil {
		a.renderError(path, err)
		return
	}
	value, err := engine.RenderWithExtra(opts.template, fields, doc.Vars, artworkTemplateData(artwork), meta)
	if err != nil {
		a.renderError(path, err)
		return
	}
	a.writeRenderValue(path, value, opts)
}

func (a *app) writeRenderValue(path, value string, opts *renderOptions) {
	a.stats.ok++
	if a.opts.json {
		data, err := json.Marshal(renderResult{File: path, Value: value})
		if err != nil {
			a.renderError(path, err)
			return
		}
		_, _ = os.Stdout.Write(append(data, renderSeparator(opts.null)...))
		return
	}
	fmt.Print(value)
	fmt.Print(string(renderSeparator(opts.null)))
}

func renderSeparator(null bool) []byte {
	if null {
		return []byte{0}
	}
	return []byte{'\n'}
}

func (a *app) renderError(path string, err error) {
	a.failed = true
	a.stats.errors++
	fmt.Fprintf(os.Stderr, "error: %s: %v\n", path, err)
}
