// Package registry wires the default set of format readers and writers.
package registry

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/daregod/ext-tags/internal/format"
	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/formats/asf"
	"github.com/daregod/ext-tags/internal/formats/matroska"
	"github.com/daregod/ext-tags/internal/formats/ogg"
	"github.com/daregod/ext-tags/internal/formats/riff"
	"github.com/daregod/ext-tags/internal/importer"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
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
	return []format.Reader{
		readerSpec{"mp3", []string{".mp3", ".mp2", ".mp1"}, readTaglib},
		readerSpec{"flac", []string{".flac"}, readFLAC},
		readerSpec{"ogg", []string{".ogg", ".oga", ".opus", ".spx"}, readOgg},
		readerSpec{"apev2", []string{".ape", ".wv", ".mpc", ".tak", ".ac3", ".dts"}, readAPEv2},
		readerSpec{"mp4", []string{".m4a", ".m4b", ".mp4", ".alac"}, readTaglib},
		readerSpec{"matroska", []string{".mkv", ".webm"}, readMatroska},
		readerSpec{"asf", []string{".wma", ".wmv", ".asf"}, readASF},
		readerSpec{"riff", []string{".wav", ".rf64", ".w64", ".aif", ".aiff"}, readRIFF},
		readerSpec{"aac", []string{".aac"}, readTaglib},
	}
}

func readTaglib(path string, _ format.ReadOptions) (*metadata.Document, error) {
	return read.Taglib(path)
}

func readFLAC(path string, _ format.ReadOptions) (*metadata.Document, error) {
	magic, err := read.ReadMagic(path, 4)
	if err == nil && string(magic) == "OggS" {
		return ogg.Read(path)
	}
	return read.Taglib(path)
}

func readOgg(path string, _ format.ReadOptions) (*metadata.Document, error) {
	return ogg.Read(path)
}

func readAPEv2(path string, _ format.ReadOptions) (*metadata.Document, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ac3":
		return ape.Read(path, "AC3")
	case ".dts":
		return ape.Read(path, "DTS")
	case ".ape":
		return ape.Read(path, "APE")
	case ".wv":
		return ape.Read(path, "WAVPACK")
	case ".mpc":
		return ape.Read(path, "MUSEPACK")
	case ".tak":
		return ape.Read(path, "TAK")
	}
	return ape.Read(path, "APEv2")
}

func readMatroska(path string, opts format.ReadOptions) (*metadata.Document, error) {
	return matroska.Read(path, opts.CoverNames)
}

func readASF(path string, _ format.ReadOptions) (*metadata.Document, error) {
	return asf.Read(path)
}

func readRIFF(path string, _ format.ReadOptions) (*metadata.Document, error) {
	return riff.Read(path)
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
		writerSpec{"riff", []string{".wav", ".rf64", ".w64", ".aif", ".aiff"}, importer.ImportRIFF},
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
