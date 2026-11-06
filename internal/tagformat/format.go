// Package format defines reader/writer interfaces and a format registry.
package tagformat

import (
	"path/filepath"
	"strings"

	"github.com/daregod/ext-tags/internal/metadata"
)

// ReadOptions tune container parsing.
type ReadOptions struct {
	CoverNames []string
}

// WriteOptions tune tag writing.
type WriteOptions struct {
	Replace    bool
	Backup     bool
	ID3Version int
	ID3V1      bool
	// Verify controls post-write verification: ""/"error", "warn" or "off".
	Verify string
}

// Reader reads a container into a metadata document.
type Reader interface {
	Name() string
	Extensions() []string
	Read(path string, opts ReadOptions) (*metadata.Document, error)
}

// Writer writes a metadata document into a container.
type Writer interface {
	Name() string
	Extensions() []string
	Write(path string, doc *metadata.Document, opts WriteOptions) error
}

// Registry is an ordered list of readers and writers.
type Registry struct {
	readers []Reader
	writers []Writer
}

// New builds a registry from explicit reader/writer lists. Order matters:
// the first matching extension wins.
func New(readers []Reader, writers []Writer) *Registry {
	return &Registry{
		readers: append([]Reader(nil), readers...),
		writers: append([]Writer(nil), writers...),
	}
}

// ReaderFor returns the first reader accepting path's extension.
func (r *Registry) ReaderFor(path string) (Reader, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	for _, reader := range r.readers {
		if hasExtension(reader.Extensions(), ext) {
			return reader, true
		}
	}
	return nil, false
}

// WriterFor returns the first writer accepting path's extension.
func (r *Registry) WriterFor(path string) (Writer, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	for _, writer := range r.writers {
		if hasExtension(writer.Extensions(), ext) {
			return writer, true
		}
	}
	return nil, false
}

// Readers returns a copy of the registered readers.
func (r *Registry) Readers() []Reader {
	return append([]Reader(nil), r.readers...)
}

// Writers returns a copy of the registered writers.
func (r *Registry) Writers() []Writer {
	return append([]Writer(nil), r.writers...)
}

func hasExtension(extensions []string, ext string) bool {
	for _, candidate := range extensions {
		if strings.EqualFold(candidate, ext) {
			return true
		}
	}
	return false
}
