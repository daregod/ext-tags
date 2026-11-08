package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/daregod/ext-tags/internal/album"
	"github.com/daregod/ext-tags/internal/artwork"
	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
	"github.com/daregod/ext-tags/internal/rules"
	"github.com/daregod/ext-tags/internal/tagfile"
	"github.com/daregod/ext-tags/internal/tagformat"
)

type albumInitOptions struct {
	force      bool
	dryRun     bool
	mergeAlbum bool
}

func (a *app) albumCmd() *cobra.Command {
	opts := &albumInitOptions{}
	cmd := &cobra.Command{
		Use:   "album",
		Short: "Album metadata helpers",
	}
	initCmd := &cobra.Command{
		Use:   "init <album dir...>",
		Short: "Create an album.yaml skeleton from audio tags",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runAlbumInit(args, opts)
		},
	}
	initCmd.Flags().BoolVar(&opts.force, "force", false, "overwrite an existing album.yaml")
	initCmd.Flags().BoolVar(&opts.mergeAlbum, "merge-album", false, "supplement an existing album.yaml without overwriting its values")
	initCmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "print album.yaml without writing")
	cmd.AddCommand(initCmd)
	return cmd
}

func (a *app) runAlbumInit(paths []string, opts *albumInitOptions) error {
	if opts.mergeAlbum && opts.force {
		return fmt.Errorf("--merge-album and --force are mutually exclusive")
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
	engine, err := rules.New(cfg)
	if err != nil {
		return err
	}
	for _, path := range paths {
		files, dir, err := albumFiles(path, cfg)
		if err != nil {
			return err
		}
		inputs := make([]album.Input, 0, len(files))
		for _, file := range files {
			doc, err := a.readAlbumFile(file, cfg)
			if err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			var existing *ape.ParsedTag
			if cfg.Files.MergeExisting {
				existing, err = tagfile.ReadExisting(file, cfg.TagFile.Suffix)
				if err != nil {
					return fmt.Errorf("%s: %w", file, err)
				}
			}
			fields, err := engine.BuildWithExisting(doc, tagfile.ExistingFields(existing), rules.Meta{
				Path: file, Filename: baseName(file), Dir: dir, Ext: filepath.Ext(file),
			})
			if err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			inputs = append(inputs, album.Input{AudioPath: file, Fields: fields})
		}
		art := albumArtwork(dir, cfg)
		generated := album.Generate(inputs, art)
		if opts.mergeAlbum {
			existing, err := album.Load(dir)
			if err != nil {
				return err
			}
			generated = album.Merge(existing, generated)
		}
		data, err := generated.Marshal()
		if err != nil {
			return err
		}
		if opts.dryRun || a.opts.dryRun {
			fmt.Printf("# %s\n%s", filepath.Join(dir, "album.yaml"), data)
			continue
		}
		target, err := generated.Save(dir, opts.force || opts.mergeAlbum)
		if err != nil {
			return err
		}
		fmt.Println(target)
	}
	return nil
}

func (a *app) readAlbumFile(path string, cfg *config.Config) (*metadata.Document, error) {
	reader, ok := a.getRegistry().ReaderFor(path)
	if !ok {
		return nil, fmt.Errorf("unsupported format %s", filepath.Ext(path))
	}
	doc, err := reader.Read(path, tagformat.ReadOptions{CoverNames: cfg.Cover.Names})
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func albumFiles(path string, cfg *config.Config) ([]string, string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}
	if !info.IsDir() {
		if !read.IsSupported(path, cfg.Files.Extensions) {
			return nil, "", fmt.Errorf("%s: unsupported format", path)
		}
		return []string{path}, filepath.Dir(path), nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, "", err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		file := filepath.Join(path, entry.Name())
		if read.IsSupported(file, cfg.Files.Extensions) {
			files = append(files, file)
		}
	}
	if len(files) == 0 {
		return nil, "", fmt.Errorf("%s: no supported audio files", path)
	}
	sort.Strings(files)
	return files, path, nil
}

func albumArtwork(dir string, cfg *config.Config) map[string]string {
	if !cfg.Cover.Enabled || !cfg.Cover.Directory {
		return nil
	}
	out := map[string]string{}
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
		if img, err := artwork.Find(dir, spec.names); err == nil {
			out[spec.kind] = img.Filename
		}
	}
	return out
}

func baseName(path string) string {
	name := filepath.Base(path)
	return name[:len(name)-len(filepath.Ext(name))]
}
