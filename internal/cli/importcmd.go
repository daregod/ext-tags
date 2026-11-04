package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/daregod/ext-tags/internal/importer"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/tagformat"
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
		Short: "Write .tag fields back into audio files (MP3, FLAC, Ogg, MP4, APEv2, Matroska, ASF)",
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
	case ".mp3", ".mp2", ".mp1", ".flac", ".ogg", ".oga", ".opus", ".spx", ".ape", ".wv", ".mpc", ".tak", ".ac3", ".dts", ".m4a", ".m4b", ".mp4", ".alac", ".mkv", ".webm", ".wma", ".wmv", ".asf", ".wav", ".rf64", ".w64", ".aif", ".aiff":
	default:
		res.Status = "unsupported"
		res.Reason = "import is implemented for MP3, FLAC and Ogg Vorbis/Opus/Speex"
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

	writer, ok := a.getRegistry().WriterFor(path)
	if !ok {
		res.Status = "unsupported"
		res.Reason = "no writer for this format"
		a.report(res)
		return
	}
	reader, ok := a.getRegistry().ReaderFor(path)
	if !ok {
		res.Status = "unsupported"
		res.Reason = "no reader for verification"
		a.report(res)
		return
	}
	if err := importer.WriteAtomic(path, writer, reader, documentFromTagFile(tf), tagformat.WriteOptions{
		Replace:    opts.replace,
		Backup:     opts.backup,
		ID3Version: opts.version,
		ID3V1:      opts.v1,
	}); err != nil {
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
	if strings.EqualFold(filepath.Ext(path), ".flac") && sniffContainer(path) == "ogg" {
		res.Warning = "Ogg/Opus in a .flac file: foobar2000 picks the artwork reader by extension; rename to .opus to see embedded art"
	}
	a.report(res)
}
func sniffContainer(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	head := make([]byte, 4)
	if _, err := f.Read(head); err != nil {
		return ""
	}
	switch string(head) {
	case "OggS":
		return "ogg"
	case "fLaC":
		return "flac"
	}
	return ""
}
func documentFromTagFile(tf *importer.TagFile) *metadata.Document {
	doc := metadata.New()
	for key, value := range tf.Fields {
		doc.Fields[key] = value
	}
	for kind, cover := range tf.Covers {
		doc.Artwork[kind] = cover
	}
	doc.Cover = tf.Cover
	return doc
}
