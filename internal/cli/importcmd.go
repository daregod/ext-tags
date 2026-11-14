package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/importer"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/tagformat"
)

type importOptions struct {
	replace bool
	version int
	v1      bool
	backup  bool
	verify  string
	only    []string
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
	cmd.Flags().StringVar(&opts.verify, "verify", "error", "post-write verification: error, warn or off")
	cmd.Flags().StringSliceVar(&opts.only, "only", nil, "import only these fields; missing selected fields are removed")
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
	if opts.verify == "" {
		opts.verify = "error"
	}
	switch opts.verify {
	case "error", "warn", "off":
	default:
		return fmt.Errorf("--verify must be error, warn or off")
	}
	if len(opts.only) > 0 {
		opts.only = cleanOnly(opts.only)
		if len(opts.only) == 0 {
			return fmt.Errorf("--only requires at least one field")
		}
		if opts.replace {
			fmt.Fprintln(os.Stderr, "warning: --replace has no effect with --only")
			a.stats.warnings++
			opts.replace = false
		}
	}
	if a.opts.verbose && !a.opts.json {
		fmt.Fprintf(os.Stderr, "config: %s\n", label)
	}

	files, err := expandPaths(paths, cfg.Files.Recursive, cfg.Files.Extensions)
	if err != nil {
		return err
	}
	for _, file := range files {
		a.importFile(file, cfg.TagFile.Suffix, cfg, opts)
	}
	a.printSummary()
	return nil
}

func (a *app) importFile(path, suffix string, cfg *config.Config, opts *importOptions) {
	if cfg == nil {
		cfg = config.Default()
	}
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

	if len(opts.only) > 0 && !onlyFormatSupported(path) {
		res.Status = "error"
		res.Error = fmt.Sprintf("--only is not supported for %s", strings.ToLower(filepath.Ext(path)))
		a.failed = true
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
	if len(opts.only) > 0 {
		selected := make(map[string]bool, len(opts.only))
		for _, field := range opts.only {
			selected[strings.ToLower(field)] = true
		}
		filtered := make(map[string]string)
		for field, value := range tf.Fields {
			if selected[strings.ToLower(field)] && !isArtworkName(field) {
				filtered[field] = value
			}
		}
		tf.Fields = filtered
		artOnly := onlyArtwork(opts.only)
		if len(artOnly) == 0 {
			tf.Covers = map[string]*importer.Cover{}
			tf.Cover = nil
		} else {
			covers := make(map[string]*importer.Cover)
			for kind, cover := range tf.Covers {
				if artOnly[kind] {
					covers[kind] = cover
				}
			}
			tf.Covers = covers
			tf.Cover = covers["front"]
		}
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
	doc := documentFromTagFile(tf)
	doc.Artwork = applyCoverBudget(tf.Covers, cfg)
	doc.Cover = doc.Artwork["front"]
	if doc.Cover != nil {
		res.Cover = &coverInfo{Filename: doc.Cover.Filename, Bytes: len(doc.Cover.Data)}
	} else {
		res.Cover = nil
	}
	warnings, err := importer.WriteAtomicResult(path, writer, reader, doc, tagformat.WriteOptions{
		Replace:    opts.replace,
		Backup:     opts.backup,
		ID3Version: opts.version,
		ID3V1:      opts.v1,
		Verify:     opts.verify,
		Only:       opts.only,
	})
	if err != nil {
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
	var notes []string
	notes = append(notes, warnings...)
	if strings.EqualFold(filepath.Ext(path), ".flac") && sniffContainer(path) == "ogg" {
		notes = append(notes, "Ogg/Opus in a .flac file: foobar2000 picks the artwork reader by extension; rename to .opus to see embedded art")
	}
	res.Warning = strings.Join(notes, "; ")
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
func cleanOnly(fields []string) []string {
	seen := make(map[string]bool, len(fields))
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		key := strings.ToLower(field)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, field)
	}
	return out
}
func onlyFormatSupported(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3", ".mp2", ".mp1", ".flac", ".ogg", ".oga", ".opus", ".spx",
		".ape", ".wv", ".mpc", ".tak", ".ac3", ".dts",
		".m4a", ".m4b", ".mp4", ".alac", ".mkv", ".webm",
		".wma", ".wmv", ".asf",
		".wav", ".rf64", ".w64", ".aif", ".aiff":
		return true
	}
	return false
}

var artworkNames = map[string]string{
	"cover art (front)": "front",
	"cover art (back)":  "back",
	"artist picture":    "artist",
	"cover art (disc)":  "disc",
	"cover art (icon)":  "icon",
}

func isArtworkName(name string) bool {
	_, ok := artworkNames[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

func onlyArtwork(fields []string) map[string]bool {
	out := make(map[string]bool)
	for _, field := range fields {
		if kind, ok := artworkNames[strings.ToLower(strings.TrimSpace(field))]; ok {
			out[kind] = true
		}
	}
	return out
}
