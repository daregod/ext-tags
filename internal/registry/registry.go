// Package registry wires the default set of format readers and writers.
package registry

import (
	"os"

	"github.com/daregod/ext-tags/internal/extractor"
	"github.com/daregod/ext-tags/internal/format"
	"github.com/daregod/ext-tags/internal/importer"
	"github.com/daregod/ext-tags/internal/metadata"
)

// NewDefault returns the production registry. The list is explicit so tests
// can build a smaller registry with format.New instead.
func NewDefault() *format.Registry {
	return format.New(defaultReaders(), defaultWriters())
}

type readerSpec struct {
	name       string
	extensions []string
	read       func(string, format.ReadOptions) (*metadata.Document, error)
}

func (r readerSpec) Name() string         { return r.name }
func (r readerSpec) Extensions() []string { return r.extensions }
func (r readerSpec) Read(path string, opts format.ReadOptions) (*metadata.Document, error) {
	return r.read(path, opts)
}

func defaultReaders() []format.Reader {
	read := func(path string, opts format.ReadOptions) (*metadata.Document, error) {
		return extractor.ExtractWithOptions(path, extractor.Options{CoverNames: opts.CoverNames})
	}
	return []format.Reader{
		readerSpec{"mp3", []string{".mp3", ".mp2", ".mp1"}, read},
		readerSpec{"flac", []string{".flac"}, read},
		readerSpec{"ogg", []string{".ogg", ".oga", ".opus", ".spx"}, read},
		readerSpec{"apev2", []string{".ape", ".wv", ".mpc", ".tak", ".ac3", ".dts"}, read},
		readerSpec{"mp4", []string{".m4a", ".m4b", ".mp4", ".alac"}, read},
		readerSpec{"matroska", []string{".mkv", ".webm"}, read},
		readerSpec{"asf", []string{".wma", ".wmv", ".asf"}, read},
		readerSpec{"riff", []string{".wav", ".rf64", ".w64", ".aif", ".aiff"}, read},
		readerSpec{"aac", []string{".aac"}, read},
	}
}

type writerSpec struct {
	name       string
	extensions []string
	write      func(string, *importer.TagFile, importer.Options) error
}

func (w writerSpec) Name() string         { return w.name }
func (w writerSpec) Extensions() []string { return w.extensions }
func (w writerSpec) Write(path string, doc *metadata.Document, opts format.WriteOptions) error {
	return w.write(path, tagFile(doc), importer.Options{
		Replace: opts.Replace,
		Backup:  opts.Backup,
		Version: opts.ID3Version,
		V1:      opts.ID3V1,
	})
}

func defaultWriters() []format.Writer {
	return []format.Writer{
		writerSpec{"mp3", []string{".mp3", ".mp2", ".mp1"}, importer.ImportMP3},
		writerSpec{"flac", []string{".flac"}, writeFLAC},
		writerSpec{"ogg", []string{".ogg", ".oga", ".opus", ".spx"}, importer.ImportOgg},
		writerSpec{"apev2", []string{".ape", ".wv", ".mpc", ".tak", ".ac3", ".dts"}, importer.ImportAPEv2},
		writerSpec{"mp4", []string{".m4a", ".m4b", ".mp4", ".alac"}, importer.ImportMP4},
		writerSpec{"matroska", []string{".mkv", ".webm"}, importer.ImportMatroska},
		writerSpec{"asf", []string{".wma", ".wmv", ".asf"}, importer.ImportASF},
	}
}

func writeFLAC(path string, tf *importer.TagFile, opts importer.Options) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	head := make([]byte, 4)
	_, readErr := file.Read(head)
	file.Close()
	if readErr == nil && string(head) == "OggS" {
		return importer.ImportOgg(path, tf, opts)
	}
	return importer.ImportFLAC(path, tf, opts)
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
