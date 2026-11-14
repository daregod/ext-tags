package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/daregod/ext-tags/internal/album"
	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
	"github.com/daregod/ext-tags/internal/rules"
	"github.com/daregod/ext-tags/internal/tagfile"
)

type applyOptions struct {
	tag     string
	output  string
	stdout  bool
	force   bool
	diff    bool
	noAlbum bool
}

type applyResult struct {
	File   string            `json:"file"`
	Status string            `json:"status"`
	Output string            `json:"output,omitempty"`
	Items  int               `json:"items,omitempty"`
	Bytes  int               `json:"bytes,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
	Diff   *diffResult       `json:"diff,omitempty"`
}

func (a *app) applyCmd() *cobra.Command {
	opts := &applyOptions{}
	cmd := &cobra.Command{
		Use:   "apply --tag <file.tag>",
		Short: "Apply rules to an external .tag without an audio file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runApply(opts)
		},
	}
	cmd.Flags().StringVar(&opts.tag, "tag", "", "input .tag file")
	cmd.Flags().StringVar(&opts.output, "output", "", "output .tag path (- for stdout)")
	cmd.Flags().BoolVar(&opts.stdout, "stdout", false, "write the resulting .tag to stdout")
	cmd.Flags().BoolVar(&opts.force, "force", false, "overwrite the output file")
	cmd.Flags().BoolVar(&opts.diff, "diff", false, "show changes without writing (implies --dry-run)")
	cmd.Flags().BoolVar(&opts.noAlbum, "no-album", false, "do not apply album.yaml")
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

func cloneFieldMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func (a *app) runApply(opts *applyOptions) error {
	if opts.tag == "" {
		return fmt.Errorf("--tag is required")
	}
	if opts.diff {
		a.opts.dryRun = true
	}
	if opts.diff && (opts.stdout || opts.output != "") {
		return fmt.Errorf("--diff cannot be combined with --output/--stdout")
	}
	if opts.output == "" && !opts.stdout && !a.opts.dryRun {
		return fmt.Errorf("--output or --stdout is required")
	}
	if (opts.stdout || opts.output == "-") && a.opts.summary {
		return fmt.Errorf("--stdout cannot be combined with --summary")
	}
	cfg, _, err := a.effectiveConfig()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	parsed, err := ape.ReadFile(opts.tag)
	if err != nil {
		return fmt.Errorf("read %s: %w", opts.tag, err)
	}
	doc := metadata.New()
	doc.Fields = tagfile.ExistingFields(parsed)
	doc.Artwork = tagfile.ExistingArtwork(parsed)
	existingFields := cloneFieldMap(doc.Fields)
	existingArt := tagfile.ExistingArtwork(parsed)
	var albumOverrides map[string]string
	if !opts.noAlbum {
		albumCfg, err := album.Load(filepath.Dir(opts.tag))
		if err != nil {
			return err
		}
		if albumCfg != nil {
			audioPath := strings.TrimSuffix(opts.tag, cfg.TagFile.Suffix)
			albumOverrides = albumCfg.Apply(doc, audioPath)
			doc.Fields = cloneFieldMap(existingFields)
			albumCovers, err := albumCfg.ArtworkCovers()
			if err != nil {
				return err
			}
			for kind, cover := range albumCovers {
				doc.Artwork[kind] = cover
			}
			if cover := doc.Artwork[metadata.Front]; cover != nil {
				doc.Cover = cover
			}
		}
	}
	doc.Cover = doc.Artwork[metadata.Front]
	if track, total := read.SplitPair(read.FieldValue(doc.Fields, "Track")); track != "" {
		read.PutVar(doc.Vars, "tracknumber", track)
		read.PutVar(doc.Vars, "totaltracks", total)
	}
	engine, err := rules.New(cfg)
	if err != nil {
		return err
	}
	meta := rules.Meta{
		Path: opts.tag, Filename: strings.TrimSuffix(filepath.Base(opts.tag), filepath.Ext(opts.tag)),
		Dir: filepath.Dir(opts.tag), Ext: filepath.Ext(opts.tag), Now: time.Now(),
	}
	fields, err := engine.BuildWithOverridesPrefer(doc, existingFields, albumOverrides, meta, true)
	if err != nil {
		return err
	}
	if opts.diff {
		resDiff := buildDiff(existingFields, fields, existingArt, doc.Artwork, cfg)
		res := applyResult{File: opts.tag, Status: "diff", Fields: fields, Diff: resDiff}
		a.addResult(fileResult{Status: "diff"})
		if a.opts.json {
			writeJSONRecord(res)
		} else if !a.opts.quiet {
			printFieldDiff(resDiff, opts.tag, "DIFF")
		}
		a.printSummary()
		return nil
	}
	data, items, merged, err := tagfile.BuildFinal(fields, doc.Artwork, parsed, cfg)
	if err != nil {
		return err
	}
	res := applyResult{File: opts.tag, Status: "ok", Items: items, Bytes: len(data), Fields: merged}
	if opts.stdout || opts.output == "-" {
		if !a.opts.dryRun {
			_, err = os.Stdout.Write(data)
			return err
		}
		res.Status = "dry-run"
	} else if !a.opts.dryRun {
		if opts.output == opts.tag && !opts.force {
			return fmt.Errorf("output equals input; use --force")
		}
		if _, err := os.Stat(opts.output); err == nil && !opts.force {
			return fmt.Errorf("%s already exists (use --force)", opts.output)
		}
		if err := os.WriteFile(opts.output, data, 0644); err != nil {
			return err
		}
		res.Output = opts.output
	} else {
		res.Status = "dry-run"
	}
	a.addResult(fileResult{Status: res.Status, Items: res.Items, Bytes: res.Bytes})
	if !a.opts.quiet {
		if a.opts.json {
			writeJSONRecord(res)
		} else if res.Status == "dry-run" {
			fmt.Printf("DRY-RUN %s (%d items, %d bytes)\n", opts.tag, items, len(data))
		} else {
			fmt.Printf("Wrote %s (%d items, %d bytes)\n", res.Output, items, len(data))
		}
	}
	a.printSummary()
	return nil
}
