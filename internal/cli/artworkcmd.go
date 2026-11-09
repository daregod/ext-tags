package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/daregod/ext-tags/internal/album"
	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
	"github.com/daregod/ext-tags/internal/tagfile"
	"github.com/daregod/ext-tags/internal/tagformat"
)

type artworkExtractOptions struct {
	slot   string
	output string
	force  bool
}

type artworkResult struct {
	File     string `json:"file"`
	Slot     string `json:"slot"`
	Filename string `json:"filename,omitempty"`
	MIME     string `json:"mime,omitempty"`
	Bytes    int    `json:"bytes,omitempty"`
	Output   string `json:"output,omitempty"`
}

func (a *app) artworkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artwork",
		Short: "Inspect and extract artwork",
	}
	listCmd := &cobra.Command{
		Use:   "list <path...>",
		Short: "List selected artwork slots",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runArtworkList(args)
		},
	}
	opts := &artworkExtractOptions{}
	extractCmd := &cobra.Command{
		Use:   "extract <audio file>",
		Short: "Write one artwork slot to a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runArtworkExtract(args[0], opts)
		},
	}
	extractCmd.Flags().StringVar(&opts.slot, "slot", "front", "front|back|artist|disc|icon")
	extractCmd.Flags().StringVar(&opts.output, "output", "", "output image path")
	extractCmd.Flags().BoolVar(&opts.force, "force", false, "overwrite an existing output file")
	cmd.AddCommand(listCmd, extractCmd)
	return cmd
}

func (a *app) runArtworkList(paths []string) error {
	cfg, _, err := a.effectiveConfig()
	if err != nil {
		return err
	}
	if a.opts.noMerge {
		cfg.Files.MergeExisting = false
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	files, err := expandPaths(paths, cfg.Files.Recursive, cfg.Files.Extensions)
	if err != nil {
		return err
	}
	for _, file := range files {
		if a.isOwnOutput(file, cfg) {
			continue
		}
		covers, err := a.loadArtwork(file, cfg)
		if err != nil {
			a.artworkError(file, err)
			continue
		}
		for _, slot := range artworkKinds {
			cover := covers[slot]
			if cover == nil {
				continue
			}
			res := artworkResult{File: file, Slot: slot, Filename: cover.Filename, MIME: cover.MIME, Bytes: len(cover.Data)}
			if a.opts.json {
				writeJSONRecord(res)
			} else {
				fmt.Printf("%s\t%s\t%s\t%d\n", file, slot, cover.Filename, len(cover.Data))
			}
		}
	}
	return nil
}

func (a *app) runArtworkExtract(path string, opts *artworkExtractOptions) error {
	if !validArtworkSlot(opts.slot) {
		return fmt.Errorf("--slot must be front, back, artist, disc or icon")
	}
	if opts.output == "" {
		return fmt.Errorf("--output is required")
	}
	if !opts.force {
		if _, err := os.Stat(opts.output); err == nil {
			return fmt.Errorf("%s already exists (use --force)", opts.output)
		}
	}
	cfg, _, err := a.effectiveConfig()
	if err != nil {
		return err
	}
	if a.opts.noMerge {
		cfg.Files.MergeExisting = false
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	covers, err := a.loadArtwork(path, cfg)
	if err != nil {
		return err
	}
	cover := covers[opts.slot]
	if cover == nil || len(cover.Data) == 0 {
		return fmt.Errorf("no %s artwork for %s", opts.slot, path)
	}
	if err := os.WriteFile(opts.output, cover.Data, 0644); err != nil {
		return err
	}
	res := artworkResult{File: path, Slot: opts.slot, Filename: cover.Filename, MIME: cover.MIME, Bytes: len(cover.Data), Output: opts.output}
	if a.opts.json {
		writeJSONRecord(res)
	} else {
		fmt.Printf("Wrote %s (%s, %d bytes)\n", opts.output, opts.slot, len(cover.Data))
	}
	return nil
}

func (a *app) loadArtwork(path string, cfg *config.Config) (map[string]*metadata.Cover, error) {
	reader, ok := a.getRegistry().ReaderFor(path)
	if !ok {
		return nil, fmt.Errorf("unsupported format %s", strings.ToLower(filepath.Ext(path)))
	}
	doc, err := reader.Read(path, tagformat.ReadOptions{})
	if errors.Is(err, read.ErrNoTags) {
		doc = metadata.New()
	} else if err != nil {
		return nil, err
	}
	var albumCovers map[string]*metadata.Cover
	if albumCfg, err := album.Load(filepath.Dir(path)); err != nil {
		return nil, err
	} else if albumCfg != nil {
		albumCovers, err = albumCfg.ArtworkCovers()
		if err != nil {
			return nil, err
		}
	}
	var existing *ape.ParsedTag
	if cfg.Files.MergeExisting {
		existing, err = tagfile.ReadExisting(path, cfg.TagFile.Suffix)
		if err != nil {
			return nil, err
		}
	}
	return selectArtwork(copyCovers(doc.Artwork), a.sidecarArtwork(path, cfg), albumCovers, tagfile.ExistingArtwork(existing), cfg), nil
}

func validArtworkSlot(slot string) bool {
	for _, kind := range artworkKinds {
		if slot == kind {
			return true
		}
	}
	return false
}

func (a *app) artworkError(path string, err error) {
	a.failed = true
	fmt.Fprintf(os.Stderr, "error: %s: %v\n", path, err)
}

func writeJSONRecord(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	_, _ = os.Stdout.Write(append(data, '\n'))
}
