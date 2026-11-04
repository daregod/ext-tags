// Package registry wires the default set of format readers and writers.
package registry

import (
	"os"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/extractor"
	"github.com/daregod/ext-tags/internal/format"
	"github.com/daregod/ext-tags/internal/importer"
	"github.com/daregod/ext-tags/internal/metadata"
)

// NewDefault returns the production registry. The list is explicit so tests
// can build a smaller registry with format.New instead.
func NewDefault() *format.Registry {
	return format.New(
		[]format.Reader{autoReader{extensions: config.Default().Files.Extensions}},
		defaultWriters(),
	)
}

type autoReader struct {
	extensions []string
}

func (r autoReader) Name() string         { return "auto" }
func (r autoReader) Extensions() []string { return r.extensions }
func (r autoReader) Read(path string, opts format.ReadOptions) (*metadata.Document, error) {
	return extractor.ExtractWithOptions(path, extractor.Options{CoverNames: opts.CoverNames})
}

type writerFunc struct {
	name       string
	extensions []string
	write      func(string, *importer.TagFile, importer.Options) error
}

func (w writerFunc) Name() string         { return w.name }
func (w writerFunc) Extensions() []string { return w.extensions }
func (w writerFunc) Write(path string, doc *metadata.Document, opts format.WriteOptions) error {
	return w.write(path, tagFile(doc), importer.Options{
		Replace: opts.Replace,
		Backup:  opts.Backup,
		Version: opts.ID3Version,
		V1:      opts.ID3V1,
	})
}

func defaultWriters() []format.Writer {
	return []format.Writer{
		writerFunc{"mp3", []string{".mp3", ".mp2", ".mp1"}, importer.ImportMP3},
		flacWriter{},
		writerFunc{"ogg", []string{".ogg", ".oga", ".opus", ".spx"}, importer.ImportOgg},
		writerFunc{"apev2", []string{".ape", ".wv", ".mpc", ".tak", ".ac3", ".dts"}, importer.ImportAPEv2},
		writerFunc{"mp4", []string{".m4a", ".m4b", ".mp4"}, importer.ImportMP4},
		writerFunc{"matroska", []string{".mkv", ".webm"}, importer.ImportMatroska},
		writerFunc{"asf", []string{".wma", ".wmv", ".asf"}, importer.ImportASF},
	}
}

func tagFile(doc *metadata.Document) *importer.TagFile {
	fields := make(map[string]string, len(doc.Fields))
	for key, value := range doc.Fields {
		fields[key] = value
	}
	covers := make(map[string]*importer.Cover, len(doc.Artwork))
	for kind, cover := range doc.Artwork {
		covers[kind] = cover
	}
	if doc.Cover != nil && covers[metadata.Front] == nil {
		covers[metadata.Front] = doc.Cover
	}
	return &importer.TagFile{Fields: fields, Cover: doc.Cover, Covers: covers}
}

type flacWriter struct{}

func (flacWriter) Name() string         { return "flac" }
func (flacWriter) Extensions() []string { return []string{".flac"} }
func (flacWriter) Write(path string, doc *metadata.Document, opts format.WriteOptions) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	head := make([]byte, 4)
	_, readErr := file.Read(head)
	file.Close()
	if readErr == nil && string(head) == "OggS" {
		return importer.ImportOgg(path, tagFile(doc), importer.Options{
			Replace: opts.Replace,
			Backup:  opts.Backup,
		})
	}
	return importer.ImportFLAC(path, tagFile(doc), importer.Options{
		Replace: opts.Replace,
		Backup:  opts.Backup,
	})
}
