package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/daregod/ext-tags/internal/importer"
)

type importOptions struct {
	replace bool
	version int
	v1      bool
	backup  bool
}

func (a *app) importCmd() *cobra.Command {
	opts := &importOptions{}
	cmd := &cobra.Command{
		Use:   "import <audio files...>",
		Short: "Write .tag fields back into audio files (MP3, FLAC)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runImport(args, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.replace, "replace", false, "drop audio tag frames that are not in the .tag file")
	cmd.Flags().IntVar(&opts.version, "id3v2-version", 3, "ID3v2 version to write: 2, 3 or 4")
	cmd.Flags().BoolVar(&opts.v1, "id3v1", false, "also write an ID3v1 tag")
	cmd.Flags().BoolVar(&opts.backup, "backup", false, "keep a copy of the audio file as <name>.bak")
	return cmd
}

func (a *app) runImport(paths []string, opts *importOptions) error {
	cfg, label, err := a.effectiveConfig()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if opts.version != 2 && opts.version != 3 && opts.version != 4 {
		return fmt.Errorf("--id3v2-version must be 2, 3 or 4")
	}
	if a.opts.verbose && !a.opts.json {
		fmt.Fprintf(os.Stderr, "config: %s\n", label)
	}

	files, err := expandPaths(paths, cfg.Files.Recursive, cfg.Files.Extensions)
	if err != nil {
		return err
	}
	for _, file := range files {
		a.importFile(file, cfg.TagFile.Suffix, opts)
	}
	return nil
}

func (a *app) importFile(path, suffix string, opts *importOptions) {
	res := fileResult{File: path}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3", ".mp2", ".mp1", ".flac":
	default:
		res.Status = "unsupported"
		res.Reason = "import is implemented for MP3/MP2/MP1 and FLAC only"
		a.report(res)
		return
	}

	tagPath := path + suffix
	if _, err := os.Stat(tagPath); err != nil {
		res.Status = "skipped"
		res.Reason = "no " + filepath.Base(tagPath)
		a.report(res)
		return
	}

	tf, err := importer.ReadTagFile(tagPath)
	if err != nil {
		res.Status = "error"
		res.Error = err.Error()
		a.failed = true
		a.report(res)
		return
	}
	res.TagFile = tagPath
	res.Items = len(tf.Fields)
	res.Fields = tf.Fields
	if tf.Cover != nil {
		res.Cover = &coverInfo{Filename: tf.Cover.Filename, Bytes: len(tf.Cover.Data)}
	}

	if a.opts.dryRun {
		res.Status = "import-dry"
		a.report(res)
		return
	}

	if err := importTags(path, tf, opts); err != nil {
		res.Status = "error"
		res.Error = err.Error()
		a.failed = true
		a.report(res)
		return
	}
	if opts.backup {
		res.Backup = path + ".bak"
	}
	res.Status = "imported"
	a.report(res)
}
func importTags(path string, tf *importer.TagFile, opts *importOptions) error {
	if strings.EqualFold(filepath.Ext(path), ".flac") {
		return importer.ImportFLAC(path, tf, importer.Options{
			Replace: opts.replace,
			Backup:  opts.backup,
		})
	}
	return importer.ImportMP3(path, tf, importer.Options{
		Version: opts.version,
		V1:      opts.v1,
		Replace: opts.replace,
		Backup:  opts.backup,
	})
}
